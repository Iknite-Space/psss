package pub

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"google.golang.org/protobuf/proto"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

type SNSPublisher struct {
	SnsClient *sns.Client
	logger    zerolog.Logger
}

func NewPubService(SnsClient *sns.Client) *SNSPublisher {
	return &SNSPublisher{logger: zerolog.Nop(), SnsClient: SnsClient}
}

// WithLogger sets the logger for the SNSPublisher.
func (s *SNSPublisher) WithLogger(logger zerolog.Logger) *SNSPublisher {
	s.logger = logger
	return s
}

// Publish publishes messages to a specified SNS topic ARN.
func (s *SNSPublisher) Publish(ctx context.Context, topicArn string, message ProtoMutationEvent[proto.Message]) error {

	payload := ProtoMutationEvent[proto.Message]{
		EventID:       message.EventID,
		EventType:     message.EventType,
		EventTime:     message.EventTime,
		Source:        message.Source,
		CorrelationID: message.CorrelationID,
		ResourceType:  message.ResourceType,
		ResourceID:    message.ResourceID,
		PerformedBy:   message.PerformedBy,
		Reason:        message.Reason,
		Before:        message.Before,
		After:         message.After,
		MetaData:      message.MetaData,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal message payload: %w", err)
	}

	publishInput := sns.PublishInput{
		TopicArn: aws.String(topicArn),
		Message:  aws.String(string(payloadBytes)),
	}

	_, err = s.SnsClient.Publish(ctx, &publishInput)
	if err != nil {
		return fmt.Errorf("failed to publish to SNS topic: %w", err)
	}

	return nil
}

type EventType uint8

const (
	EventTypeCreated EventType = 1
	EventTypeUpdated EventType = 2
	EventTypeDeleted EventType = 3
)

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

	// Information about the user who performed the action
	PerformedBy string

	// Explanation or justification for the event (if applicable)
	Reason string

	// State of the resource before the event occurred
	Before T

	// State of the resource after the event occurred
	After T

	// Additional metadata related to the event
	MetaData T
}
