package pub

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Iknite-Space/psss/models"
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
func marshalProtoMutationEventToJSON[T proto.Message](e models.ProtoMutationEvent[T]) ([]byte, error) {
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
func (s *SNSPublisher[T]) Publish(ctx context.Context, message models.ProtoMutationEvent[proto.Message]) error {
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

// publishedProtoMutationEvent is used for JSON encoding of the  proto.message fields using protojson.marshal fxn.
type publishedProtoMutationEvent struct {
	EventID       string           `json:"event_id"`
	EventType     models.EventType `json:"event_type"`
	EventTime     time.Time        `json:"timestamp"`
	Source        string           `json:"source"`
	CorrelationID string           `json:"correlation_id"`
	ResourceType  string           `json:"resource_type"`
	ResourceID    string           `json:"resource_id"`
	UserID        string           `json:"user_id"`
	Reason        string           `json:"reason"`
	Before        json.RawMessage  `json:"before,omitempty"`
	After         json.RawMessage  `json:"after,omitempty"`
	MetaData      json.RawMessage  `json:"metadata,omitempty"`
}
