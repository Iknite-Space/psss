package pubmocks

import (
	"context"

	"github.com/Iknite-Space/psss/models"
	"github.com/Iknite-Space/psss/pub"
	"google.golang.org/protobuf/proto"
)

// MockPublisherImpl is a manual mock implementation of Publisher[T]
type MockPublisherImpl[T proto.Message] struct {
}

var _ pub.Publisher[proto.Message] = (*MockPublisherImpl[proto.Message])(nil)

// MockPublish sets the mock implementation for the Publish method
func (m *MockPublisherImpl[T]) Publish(ctx context.Context, message models.ProtoMutationEvent[T]) error {
	return nil
}
