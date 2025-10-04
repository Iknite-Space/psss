package pub

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Iknite-Space/psss/models"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

// Publisher defines the interface for publishing messages
type Publisher[T proto.Message] interface {
	Publish(ctx context.Context, message models.ProtoMutationEvent[T]) error
}

type SNSPublisher[T proto.Message] struct {
	SnsClient *sns.Client
	topicArn  string
	logger    zerolog.Logger
}

var _ Publisher[proto.Message] = (*SNSPublisher[proto.Message])(nil)

func NewPubService(SnsClient *sns.Client, topicArn string) *SNSPublisher[proto.Message] {
	return &SNSPublisher[proto.Message]{
		SnsClient: SnsClient,
		logger:    zerolog.Nop(),
		topicArn:  topicArn,
	}
}

// WithLogger sets the logger for the SNSPublisher.
func (s *SNSPublisher[T]) WithLogger(logger zerolog.Logger) *SNSPublisher[T] {
	s.logger = logger
	return s
}

// marshalProtoMutationEventToJSON marshals a ProtoMutationEvent with proto.Message fields to JSON.
func marshalProtoMutationEventToJSON[T proto.Message](e models.ProtoMutationEvent[T]) ([]byte, error) {
	beforeBytes, err := protojson.Marshal(e.Before)
	if err != nil {
		return nil, fmt.Errorf("marshaling 'Before': %w", err)
	}

	afterBytes, err := protojson.Marshal(e.After)
	if err != nil {
		return nil, fmt.Errorf("marshaling 'After': %w", err)
	}

	// actual struct for JSON encoding for top level fields using json package.
	payload := models.PublishedProtoMutationEvent{
		EventID:       e.EventID,
		EventType:     e.EventType,
		EventTime:     e.EventTime,
		Source:        e.Source,
		CorrelationID: e.CorrelationID,
		ResourceType:  e.ResourceType,
		ResourceID:    e.ResourceID,
		UserID:        e.UserID,
		Reason:        e.Reason,
		Before:        beforeBytes,
		After:         afterBytes,
		MetaData:      e.MetaData,
	}

	return json.Marshal(payload)
}

// Publish publishes messages to a specified message broker.
func (s *SNSPublisher[T]) Publish(ctx context.Context, message models.ProtoMutationEvent[proto.Message]) error {
	eventBytes, err := marshalProtoMutationEventToJSON(message)
	if err != nil {
		return fmt.Errorf("failed to marshal mutation event: %w", err)
	}

	s.logger.Debug().
		RawJSON("sns_message", eventBytes).Str("correlation_id", message.CorrelationID).Msg("Publishing event to SNS")

	// Publish to SNS
	response, err := s.SnsClient.Publish(ctx, &sns.PublishInput{
		TopicArn: aws.String(s.topicArn),
		Message:  aws.String(string(eventBytes)),
	})
	if err != nil {
		return fmt.Errorf("failed to publish message to SNS: %w", err)
	}

	s.logger.Info().
		Str("message_id", *response.MessageId).Str("correlation_id", message.CorrelationID).
		Msg("Message published to SNS successfully")

	return nil
}
