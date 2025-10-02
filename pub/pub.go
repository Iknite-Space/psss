package pub

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

type SNSPublisher[T proto.Message] struct {
	SnsClient *sns.Client
	topicArn  string
	logger    zerolog.Logger
}

func NewPubService(SnsClient *sns.Client, topicArn string) *SNSPublisher[proto.Message] {
	return &SNSPublisher[proto.Message]{
		SnsClient: SnsClient,
		logger:    zerolog.Nop(),
		topicArn:  topicArn,
	}
}

// WithLogger sets the logger for the SNSPublisher.
func (s *SNSPublisher[T]) WithLogger(logger zerolog.Logger) *SNSPublisher[T] {
	s.logger = logger
	return s
}

// marshalProtoMutationEventToJSON marshals a ProtoMutationEvent with proto.Message fields to JSON.
func marshalProtoMutationEventToJSON[T proto.Message](e ProtoMutationEvent[T]) ([]byte, error) {
	beforeBytes, err := protojson.Marshal(e.Before)
	if err != nil {
		return nil, fmt.Errorf("marshaling 'Before': %w", err)
	}

	afterBytes, err := protojson.Marshal(e.After)
	if err != nil {
		return nil, fmt.Errorf("marshaling 'After': %w", err)
	}

	metaBytes, err := protojson.Marshal(e.MetaData)
	if err != nil {
		return nil, fmt.Errorf("marshaling 'MetaData': %w", err)
	}

	// actual struct for JSON encoding for top level fields using json package.
	payload := publishedProtoMutationEvent{
		EventID:       e.EventID,
		EventType:     e.EventType,
		EventTime:     e.EventTime,
		Source:        e.Source,
		CorrelationID: e.CorrelationID,
		ResourceType:  e.ResourceType,
		ResourceID:    e.ResourceID,
		UserID:        e.UserID,
		Reason:        e.Reason,
		Before:        beforeBytes,
		After:         afterBytes,
		MetaData:      metaBytes,
	}

	return json.Marshal(payload)
}

// Publish publishes messages to a specified message broker.
func (s *SNSPublisher[T]) Publish(ctx context.Context, message ProtoMutationEvent[proto.Message]) error {
	eventBytes, err := marshalProtoMutationEventToJSON(message)
	if err != nil {
		return fmt.Errorf("failed to marshal mutation event: %w", err)
	}

	s.logger.Debug().
		RawJSON("sns_message", eventBytes).Str("correlation_id", message.CorrelationID).Msg("Publishing event to SNS")

	// Publish to SNS
	response, err := s.SnsClient.Publish(ctx, &sns.PublishInput{
		TopicArn: aws.String(s.topicArn),
		Message:  aws.String(string(eventBytes)),
	})
	if err != nil {
		return fmt.Errorf("failed to publish message to SNS: %w", err)
	}

	s.logger.Info().
		Str("message_id", *response.MessageId).
		Str("correlation_id", message.CorrelationID).
		Msg("Message published to SNS successfully")

	return nil
}

type EventType uint8

const (
	EventTypeCreated EventType = 1
	EventTypeUpdated EventType = 2
	EventTypeDeleted EventType = 3
)

// publishedProtoMutationEvent is used for JSON encoding of the  proto.message fields using protojson.marshal fxn.
type publishedProtoMutationEvent struct {
	EventID       string          `json:"event_id"`
	EventType     EventType       `json:"event_type"`
	EventTime     time.Time       `json:"timestamp"`
	Source        string          `json:"source"`
	CorrelationID string          `json:"correlation_id"`
	ResourceType  string          `json:"resource_type"`
	ResourceID    string          `json:"resource_id"`
	UserID        string          `json:"user_id"`
	Reason        string          `json:"reason"`
	Before        json.RawMessage `json:"before,omitempty"`
	After         json.RawMessage `json:"after,omitempty"`
	MetaData      json.RawMessage `json:"metadata,omitempty"`
}

// Represents a generic event with proto fields encoding occurring/generated in the system.
type ProtoMutationEvent[T proto.Message] struct {
	// Unique identifier for the event
	EventID string

	// Type or category of the event (e.g., "notes.created", "notes.updated")
	EventType EventType

	// EventTime when the event occurred (in UTC)
	EventTime time.Time

	// Source service or component that generated the event
	Source string

	// ID used to correlate this event with other related events
	CorrelationID string

	// Type of resource involved in the event (e.g., "document")
	ResourceType string

	// Unique identifier for the affected resource
	ResourceID string

	// 'user_id' of the user who performed the action
	UserID string

	// Explanation or justification for the event (if applicable)
	Reason string

	// State of the resource before the event occurred
	Before T

	// State of the resource after the event occurred
	After T

	// Additional metadata related to the event
	MetaData T
}
