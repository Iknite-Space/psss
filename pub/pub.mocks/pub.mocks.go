package pubmocks

import (
	"context"

	"github.com/Iknite-Space/psss/models"
	"google.golang.org/protobuf/proto"
)

// MockPublisher is a manual mock implementation of Publisher[T]
type MockPublisher[T proto.Message] struct {
	mockPublishFunc func(ctx context.Context, message models.ProtoMutationEvent[T]) error
}

// Publisher defines the interface for publishing messages
type Publisher[T proto.Message] interface {
	MockPublish(ctx context.Context, message models.ProtoMutationEvent[T]) error
}

// NewMockPublisher creates a new instance of MockPublisher[T]
func NewMockPublisher[T proto.Message]() *MockPublisher[T] {
	return &MockPublisher[T]{}
}

var _ Publisher[proto.Message] = (*MockPublisher[proto.Message])(nil)

// MockPublish sets the mock implementation for the Publish method
func (m *MockPublisher[T]) MockPublish(ctx context.Context, message models.ProtoMutationEvent[T]) error {
	if m.mockPublishFunc != nil {
		return m.mockPublishFunc(ctx, message)
	}

	return nil
}
