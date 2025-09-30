package sub

import (
	"context"
	"time"
)

type EventMutationHandlerFn func(context.Context, MutationEvents) error

// Represents a generic event occurring/generated in the system.
type MutationEvents struct {
	// Unique identifier for the event
	EventID string `json:"event_id"`

	// Type or category of the event (e.g., "notes.created", "notes.updated")
	EventType string `json:"event_type"`

	// Timestamp when the event occurred (in UTC)
	Timestamp time.Time `json:"timestamp"`

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
	Reason string

	// State of the resource before the event occurred
	Before []byte `json:"before"`

	// State of the resource after the event occurred
	After []byte `json:"after"`

	// Additional metadata related to the event
	MetaData []byte `json:"metadata"`
}

func MutationEventHandlerToStringHandlder(handler EventMutationHandlerFn) StringHandler {
	return func(ctx context.Context, s string) error {
		return nil
	}
}
