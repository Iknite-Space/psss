package pubmocks

import (
	"context"

	"github.com/Iknite-Space/psss/models"
	"github.com/Iknite-Space/psss/pub"
	"google.golang.org/protobuf/proto"
)

// MockPublisherImplementation is a manual mock implementation of Publisher[T]
type MockPublisherImplementation[T proto.Message] struct {
	PublishMockFunc func(ctx context.Context, message models.ProtoMutationEvent[T]) error
}

var _ pub.Publisher[proto.Message] = (*MockPublisherImplementation[proto.Message])(nil)

// NewMockPublisher creates a new instance of MockPublisherImplementation[T]
func NewMockPublish[T proto.Message](Publish func(ctx context.Context, message models.ProtoMutationEvent[T]) error) *MockPublisherImplementation[T] {
	return &MockPublisherImplementation[T]{
		PublishMockFunc: Publish,
	}
}

// MockPublish sets the mock implementation for the Publish method
func (m *MockPublisherImplementation[T]) Publish(ctx context.Context, message models.ProtoMutationEvent[T]) error {
	return m.Publish(ctx, message)
}
