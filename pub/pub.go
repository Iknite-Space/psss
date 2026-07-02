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
	snstypes "github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// MaxBatchSize is the maximum number of messages allowed in a single batch
// publish request for both SQS and SNS.
const MaxBatchSize = 10

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

func NewPubService[T proto.Message](SnsClient *sns.Client, topicArn string) *SNSPublisher[T] {
	return &SNSPublisher[T]{
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
func (s *SNSPublisher[T]) Publish(ctx context.Context, message models.ProtoMutationEvent[T]) error {
	eventBytes, err := marshalProtoMutationEventToJSON(message)
	if err != nil {
		return fmt.Errorf("failed to marshal mutation event: %w", err)
	}

	s.logger.Debug().
		RawJSON("sns_message", eventBytes).
		Str("correlation_id", message.CorrelationID).Msg("Publishing event to SNS")

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

// SQSBatchPublisher sends pre-marshalled message batches directly to an SQS
// queue. Unlike SNSPublisher (single proto event → SNS topic), this matches
// the outbox-relay pattern where multiple pre-serialised payloads are sent in
// one SQS batch call.
type SQSBatchPublisher struct {
	sqsClient *sqs.Client
	logger    zerolog.Logger
}

func NewSQSBatchPublisher(sqsClient *sqs.Client) *SQSBatchPublisher {
	return &SQSBatchPublisher{
		sqsClient: sqsClient,
		logger:    zerolog.Nop(),
	}
}

// WithLogger sets the logger for the SQSBatchPublisher.
func (p *SQSBatchPublisher) WithLogger(logger zerolog.Logger) *SQSBatchPublisher {
	p.logger = logger
	return p
}

// SendBatch sends a batch of entries to the given queue URL in a single
// SQS SendMessageBatch call. Returns the number of successfully sent
// entries, or an error if the batch call itself failed. Individual failed
// entries (partial failure) are also returned as an error with details.
//
// The caller is responsible for marshalling each entry's MessageBody and
// assigning a unique Id within the batch.
func (p *SQSBatchPublisher) SendBatch(ctx context.Context, queueURL string, entries []types.SendMessageBatchRequestEntry) (int, error) {
	if len(entries) == 0 {
		return 0, nil
	}
	if len(entries) > MaxBatchSize {
		return 0, fmt.Errorf("sqs batch size %d exceeds maximum %d", len(entries), MaxBatchSize)
	}

	sent, err := p.sqsClient.SendMessageBatch(ctx, &sqs.SendMessageBatchInput{
		QueueUrl: aws.String(queueURL),
		Entries:  entries,
	})
	if err != nil {
		return 0, fmt.Errorf("send sqs batch: %w", err)
	}
	if len(sent.Failed) > 0 {
		return len(sent.Successful), fmt.Errorf("partial sqs batch failure: %v", sent.Failed)
	}
	return len(sent.Successful), nil
}

type snsBatchAPI interface {
	PublishBatch(ctx context.Context, params *sns.PublishBatchInput, optFns ...func(*sns.Options)) (*sns.PublishBatchOutput, error)
}

type SNSBatchPublisher[T proto.Message] struct {
	snsClient snsBatchAPI
	topicArn  string
	logger    zerolog.Logger
}

func NewSNSBatchPublisher[T proto.Message](snsClient snsBatchAPI, topicArn string) *SNSBatchPublisher[T] {
	return &SNSBatchPublisher[T]{
		snsClient: snsClient,
		topicArn:  topicArn,
		logger:    zerolog.Nop(),
	}
}

func (p *SNSBatchPublisher[T]) WithLogger(logger zerolog.Logger) *SNSBatchPublisher[T] {
	p.logger = logger
	return p
}

func (p *SNSBatchPublisher[T]) PublishBatch(ctx context.Context, messages []models.ProtoMutationEvent[T]) (int, error) {
	if len(messages) == 0 {
		return 0, nil
	}
	if len(messages) > MaxBatchSize {
		return 0, fmt.Errorf("sns batch size %d exceeds maximum %d", len(messages), MaxBatchSize)
	}

	entries := make([]snstypes.PublishBatchRequestEntry, len(messages))
	for i, msg := range messages {
		eventBytes, err := marshalProtoMutationEventToJSON(msg)
		if err != nil {
			return 0, fmt.Errorf("marshalling message %q: %w", msg.EventID, err)
		}

		p.logger.Debug().
			RawJSON("sns_message", eventBytes).
			Str("correlation_id", msg.CorrelationID).
			Msg("Adding event to SNS batch")

		entries[i] = snstypes.PublishBatchRequestEntry{
			Id:      aws.String(msg.EventID),
			Message: aws.String(string(eventBytes)),
		}
	}

	resp, err := p.snsClient.PublishBatch(ctx, &sns.PublishBatchInput{
		TopicArn:                   aws.String(p.topicArn),
		PublishBatchRequestEntries: entries,
	})
	if err != nil {
		return 0, fmt.Errorf("sns publish batch: %w", err)
	}
	if len(resp.Failed) > 0 {
		return len(resp.Successful), fmt.Errorf("partial sns batch failure: %v", resp.Failed)
	}

	p.logger.Info().
		Int("successful", len(resp.Successful)).
		Int("failed", len(resp.Failed)).
		Msg("SNS batch published successfully")

	return len(resp.Successful), nil
}
