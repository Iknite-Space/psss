package sub

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type EventType uint8

const (
	EventTypeCreated EventType = 1
	EventTypeUpdated EventType = 2
	EventTypeDeleted EventType = 3
)

// MutationEventSqsProcessor is a processor that reads events from an SQS queue and processes them using the
// provided handler function. It uses the AWS SDK for Go v2 to interact with SQS.
type MutationEventSqsProcessor struct {
	svc       *sqs.Client
	queueURL  string
	handlerFn StringHandler
	logger    zerolog.Logger
}


func NewMutationEventSqsProcessorx[T proto.Message](
	svc *sqs.Client,
	queueURL string,
	newMessage func() T,
	handler EventMutationProtoHandlerFn[T],
) *SqsEventProcessor {

	mutationEventHandler := MutationEventHandlerToStringHandler(handler, newMessage)

	return &SqsEventProcessor{
		svc:       svc,
		queueURL:  queueURL,
		logger:    zerolog.Nop(),
		handlerFn: jsonEventHandlerToSqsHandlerFn(),
	}
}

type EventMutationProtoHandlerFn[T proto.Message] func(context.Context, ProtoMutationEvent[T]) error

// Represents a generic event with top level fields encoding occurring/generated in the system.
type MutationEvent struct {
	// Unique identifier for the event
	EventID string `json:"event_id"`

	// Type or category of the event (e.g., "notes.created", "notes.updated")
	EventType EventType `json:"event_type"`

	// EventTime when the event occurred (in UTC)
	EventTime time.Time `json:"timestamp"`

	// Source service or component that generated the event
	Source string `json:"source"`

	// ID used to correlate this event with other related events
	CorrelationID string `json:"correlation_id"`

	// Type of resource involved in the event (e.g., "document")
	ResourceType string `json:"resource_type"`

	// Unique identifier for the affected resource
	ResourceID string `json:"resource_id"`

	// Information about the user who performed the action
	PerformedBy string `json:"performed_by"`

	// Explanation or justification for the event (if applicable)
	Reason string `json:"reason"`

	// State of the resource before the event occurred
	Before []byte `json:"before"`

	// State of the resource after the event occurred
	After []byte `json:"after"`

	// Additional metadata related to the event
	MetaData []byte `json:"metadata"`
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

func MutationEventHandlerToStringHandler[T proto.Message](handler EventMutationProtoHandlerFn[T], newMessage func() T) StringHandler {
	return func(ctx context.Context, s string) error {
		msg := &MutationEvent{}
		err := json.Unmarshal([]byte(s), msg)
		if err != nil {
			return fmt.Errorf("error unmarshaling JSON mutation event: %w", err)
		}

		before := newMessage()
		err = protojson.Unmarshal([]byte(msg.Before), before)
		if err != nil {
			return fmt.Errorf("error unmarshaling 'Before' field from protobuf: %w", err)
		}

		after := newMessage()
		err = protojson.Unmarshal([]byte(msg.After), after)
		if err != nil {
			return fmt.Errorf("error unmarshaling 'After' field from protobuf: %w", err)
		}

		meta := newMessage()
		err = protojson.Unmarshal([]byte(msg.MetaData), meta)
		if err != nil {
			return fmt.Errorf("error unmarshaling 'MetaData' field from protobuf: %w", err)
		}

		input := ProtoMutationEvent[T]{
			EventID:       msg.EventID,
			EventType:     msg.EventType,
			EventTime:     msg.EventTime,
			Source:        msg.Source,
			CorrelationID: msg.CorrelationID,
			ResourceType:  msg.ResourceType,
			ResourceID:    msg.ResourceID,
			PerformedBy:   msg.PerformedBy,
			Reason:        msg.Reason,
			Before:        before,
			After:         after,
			MetaData:      meta,
		}

		return handler(ctx, input)
	}
}
