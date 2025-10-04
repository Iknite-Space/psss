package pub

import (
	"context"

	"github.com/Iknite-Space/psss/models"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/proto"

	"github.com/aws/aws-sdk-go-v2/service/sns"
)


func MockNewPublisherService(SnsClient *sns.Client, topicArn string) *SNSPublisher[proto.Message] {
	return &SNSPublisher[proto.Message]{
		SnsClient: SnsClient,
		logger:    zerolog.Nop(),
		topicArn:  topicArn,
	}
}

// MockPublish mocks the implementation of the sending messages to a message broker.
func (s *SNSPublisher[T]) MockPublish(ctx context.Context, message models.ProtoMutationEvent[proto.Message]) error {
	 
	return nil
}
 