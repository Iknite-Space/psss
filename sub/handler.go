package sub

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/rs/zerolog"
)

const (
	defaultWaitTimeSeconds   = 20
	defaultVisibilityTimeout = 60
)

// SqsHandlerFn is a function that handles an SQS message.
type SqsHandlerFn func(ctx context.Context, message awstypes.Message) error

// SqsEventProcessor is a processor that reads events from an SQS queue and processes them using the
// provided handler function. It uses the AWS SDK for Go v2 to interact with SQS.
type SqsEventProcessor struct {
	svc       *sqs.Client
	queueURL  string
	handlerFn SqsHandlerFn
	logger    zerolog.Logger
}

// NewEventSqsProcessor creates a new SqsEventProcessor.
// The processor reads messages from the provided SQS queue URL and processes them using the provided handler function.
func NewSqsEventProcessor(svc *sqs.Client,
	queueURL string,
	handlerFn SqsHandlerFn,
) *SqsEventProcessor {
	return &SqsEventProcessor{
		svc:       svc,
		queueURL:  queueURL,
		handlerFn: handlerFn,
		logger:    zerolog.Nop(),
	}
}

// WithLogger sets the logger for the SqsEventProcessor.
func (s *SqsEventProcessor) WithLogger(logger zerolog.Logger) *SqsEventProcessor {
	s.logger = logger
	return s
}

// Run starts the processor and begins reading messages from the SQS queue. For each message, it passes the message
// to the handler function. If the handler function returns an error, the message will be released back to the queue.
// If the handler function returns nil, then the message will be deleted from the queue. The processor will continue
// running until the context is cancelled. Retries for failed messages are handled by SQS and the deadletter
// configuration of the queue.
func (s *SqsEventProcessor) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			s.logger.Info().Msg("Stopping SQS mutation event processor...")
			return nil
		default:
		}

		out, err := s.svc.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			MaxNumberOfMessages: 1,
			QueueUrl:            aws.String(s.queueURL),
			WaitTimeSeconds:     defaultWaitTimeSeconds,
			VisibilityTimeout:   defaultVisibilityTimeout,
		})
		if err != nil {
			return fmt.Errorf("failed to receive message: %w", err)
		}

		// no messages, lets wait a bit before retrying for more messages.
		if len(out.Messages) < 1 {
			s.logger.Debug().Msgf("No messages received from SQS. Retrying now")
			continue
		}

		for _, message := range out.Messages {
			messageID := ""
			if message.MessageId != nil {
				messageID = *message.MessageId
			}

			if message.ReceiptHandle == nil {
				s.logger.Error().Str("message_id", messageID).Msg("Message has no receipt handle, cannot delete")
				continue
			}
			messageHandle := *message.ReceiptHandle

			err = s.handlerFn(ctx, message)
			if err != nil {
				// Note: This is a debug message because "true" errors should be logged by the handling function.
				s.logger.Debug().Err(err).Str("message_id", messageID).Msg("Error processing message")
				continue
			}
			_, err = s.svc.DeleteMessage(ctx, &sqs.DeleteMessageInput{
				QueueUrl:      &s.queueURL,
				ReceiptHandle: aws.String(messageHandle),
			})
			if err != nil {
				s.logger.Error().Err(err).Str("message_id", messageID).Msg("Error deleting message. Warning this " +
					"message will likely get reprocessed")
			}
		}
	}
}
