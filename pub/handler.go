package pub

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Iknite-Space/psss/models"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/proto"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

type SNSPublisher struct {
	SnsClient *sns.Client
	logger    zerolog.Logger
}

func NewPubService(SnsClient *sns.Client) *SNSPublisher {
	return &SNSPublisher{logger: zerolog.Nop(), SnsClient: SnsClient}
}

// WithLogger sets the logger for the SNSPublisher.
func (s *SNSPublisher) WithLogger(logger zerolog.Logger) *SNSPublisher {
	s.logger = logger
	return s
}

// Publish publishes messages to a specified SNS topic ARN.
func (s *SNSPublisher) Publish(ctx context.Context, topicArn string, message models.ProtoMutationEvent[proto.Message]) error {

	payload := models.ProtoMutationEvent[proto.Message]{
		EventID:       message.EventID,
		EventType:     message.EventType,
		EventTime:     message.EventTime,
		Source:        message.Source,
		CorrelationID: message.CorrelationID,
		ResourceType:  message.ResourceType,
		ResourceID:    message.ResourceID,
		PerformedBy:   message.PerformedBy,
		Reason:        message.Reason,
		Before:        message.Before,
		After:         message.After,
		MetaData:      message.MetaData,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal message payload: %w", err)
	}

	publishInput := sns.PublishInput{
		TopicArn: aws.String(topicArn),
		Message:  aws.String(string(payloadBytes)),
	}

	_, err = s.SnsClient.Publish(ctx, &publishInput)
	if err != nil {
		return fmt.Errorf("failed to publish to SNS topic: %w", err)
	}

	return nil
}
