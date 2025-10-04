package pubmocks

import (
	"context"

	"github.com/Iknite-Space/psss/models"
	"google.golang.org/protobuf/proto"
)

// MockPublisherImplementation is a manual mock implementation of Publisher[T]
type MockPublisherImplementation[T proto.Message] struct {
	Publish func(ctx context.Context, message models.ProtoMutationEvent[T]) error
}

// NewMockPublisher creates a new instance of MockPublisherImplementation[T]
func NewMockPublish[T proto.Message](Publish func(ctx context.Context, message models.ProtoMutationEvent[T]) error) *MockPublisherImplementation[T] {
	return &MockPublisherImplementation[T]{
		Publish: Publish,
	}
}

// MockPublish sets the mock implementation for the Publish method
func (m *MockPublisherImplementation[T]) MockPublish(ctx context.Context, message models.ProtoMutationEvent[T]) error {
	if m.Publish != nil {
		return m.Publish(ctx, message)
	}

	return nil
}
