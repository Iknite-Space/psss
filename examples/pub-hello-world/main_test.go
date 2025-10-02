package main

import (
	"context"
	"testing"

	"github.com/Iknite-Space/psss/models"
	"github.com/Iknite-Space/psss/pub/pub.mocks"
	"github.com/stretchr/testify/require"

	"google.golang.org/protobuf/proto"
)

func TestPublish(t *testing.T) {

	newMockPubSvc := pubmocks.NewMockPublisher[proto.Message]()
	eventMsg := models.ProtoMutationEvent[proto.Message]{
		EventID:      "392f8b1e-4f1c-4d2a-9c3e-1a2b3c4d5e6f",
		EventType:    models.EventTypeCreated,
		ResourceType: "example-resource",
		ResourceID:   "resource-1",
		UserID:       "user-1",
	}
	
	err := newMockPubSvc.MockPublish(context.Background(), eventMsg)
	require.NoError(t, err)
}
