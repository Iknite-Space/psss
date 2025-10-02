package pubmocks

import (
	"context"

	"github.com/Iknite-Space/psss/models"
	"google.golang.org/protobuf/proto"
)

// MockPublisherImplementation is a manual mock implementation of Publisher[T]
type MockPublisherImplementation[T proto.Message] struct {
	mockPublishFunc func(ctx context.Context, message models.ProtoMutationEvent[T]) error
}

// Publisher defines the interface for publishing messages
type Publisher[T proto.Message] interface {
	MockPublish(ctx context.Context, message models.ProtoMutationEvent[T]) error
}

var _ Publisher[proto.Message] = (*MockPublisherImplementation[proto.Message])(nil)

// NewMockPublisher creates a new instance of MockPublisher[T]
func NewMockPublisher[T proto.Message]() *MockPublisherImplementation[T] {
	return &MockPublisherImplementation[T]{}
}

// MockPublish sets the mock implementation for the Publish method
func (m *MockPublisherImplementation[T]) MockPublish(ctx context.Context, message models.ProtoMutationEvent[T]) error {
	if m.mockPublishFunc != nil {
		return m.mockPublishFunc(ctx, message)
	}

	return nil
}
