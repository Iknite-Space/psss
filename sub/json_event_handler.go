package sub

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// JSONEventHandlerFn is a function type that processes messages of type T.
type JSONEventHandlerFn[T any] func(
	ctx context.Context,
	message T,
) error

// NewJSONSqsEventProcessor creates a new SqsEventProcessor that processes messages
// by unmarshaling the message body as JSON into the specified type T and then
// passing it to the provided JSONEventHandlerFn.
func NewJSONSqsEventProcessor[T any](svc *sqs.Client,
	queueURL string,
	handlerFn JSONEventHandlerFn[T],
) *SqsEventProcessor {
	return &SqsEventProcessor{
		svc:       svc,
		queueURL:  queueURL,
		handlerFn: jsonEventHandlerToSqsHandlerFn(handlerFn),
	}
}

// jsonEventHandlerToSqsHandlerFn converts a JSONEventHandlerFn to an SqsHandlerFn
// by unmarshaling the SQS message body into the appropriate type.
func jsonEventHandlerToSqsHandlerFn[T any](handler JSONEventHandlerFn[T],
) SqsHandlerFn {
	return func(ctx context.Context, message awstypes.Message) error {
		if message.Body == nil {
			return nil
		}

		var msgBody T
		err := json.Unmarshal([]byte(*message.Body), &msgBody)
		if err != nil {
			return fmt.Errorf("failed to unmarshal message body: %w", err)
		}

		err = handler(ctx, msgBody)
		if err != nil {
			return fmt.Errorf("failed to handle message: %w", err)
		}

		return nil
	}
}
