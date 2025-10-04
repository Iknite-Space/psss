package pubmocks

import (
	"context"

	"github.com/Iknite-Space/psss/models"
	"github.com/Iknite-Space/psss/pub"
	"google.golang.org/protobuf/proto"
)

// MockPublisher is a manual mock implementation of Publisher[T]
type MockPublisher[T proto.Message] struct {
}

var _ pub.Publisher[proto.Message] = (*MockPublisher[proto.Message])(nil)

// MockPublish sets the mock implementation for the Publish method
func (m *MockPublisher[T]) Publish(ctx context.Context, message models.ProtoMutationEvent[T]) error {
	return nil
}
