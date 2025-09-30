package sub

import (
	"context"
	"time"

	"google.golang.org/protobuf/proto"
)

type EventType uint8

const (
	EventTypeCreated EventType = 1
	EventTypeUpdated EventType = 2
	EventTypeDeleted EventType = 3
)

type EventMutationHandlerFn[T proto.Message] func(context.Context, MutationEvents[T]) error

// Represents a generic event occurring/generated in the system.
type MutationEvents[T proto.Message] struct {
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

	// Additional metadata related to the event
	MetaData T `json:"metadata"`

	// State of the resource after the event occurred
	After T `json:"after"`
}

func MutationEventHandlerToStringHandlder[T proto.Message](handler EventMutationHandlerFn[T]) StringHandler {
	return func(ctx context.Context, s string) error {
		return nil
	}
}
