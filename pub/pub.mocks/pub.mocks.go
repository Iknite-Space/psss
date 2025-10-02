package pubmocks

import (
	"context"

	"github.com/Iknite-Space/psss/models"
	"github.com/Iknite-Space/psss/pub"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/proto"
)

// MockPublisher is a manual mock implementation of Publisher[T]
type MockPublisher[T proto.Message] struct {
	mockPublishFunc func(ctx context.Context, message models.ProtoMutationEvent[T]) error
	WithLoggerFunc  func(logger zerolog.Logger) *pub.SNSPublisher[T]
}

// NewMockPublisher creates a new instance of MockPublisher[T]
func NewMockPublisher[T proto.Message]() *MockPublisher[T] {
	return &MockPublisher[T]{}
}

// MockPublish sets the mock implementation for the Publish method
func (m *MockPublisher[T]) MockPublish(ctx context.Context, message models.ProtoMutationEvent[T]) error {
	if m.mockPublishFunc != nil {
		return m.mockPublishFunc(ctx, message)
	}
	return nil
}

// MockWithLogger sets the mock implementation for the WithLogger method
func (m *MockPublisher[T]) MockWithLogger(logger zerolog.Logger) *pub.SNSPublisher[T] {
	return &pub.SNSPublisher[T]{}
}
