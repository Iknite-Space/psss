package sub

import (
	"context"
	"errors"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type StringHandlerFn func(ctx context.Context, msg string) error


type SnsWrapper struct {
	Message string `json:"Message"`
}

// StringHandlerToSnsWrapperHandler wraps an SNS message payload into a handler that extracts
// the 'Message' field and passes it to the provided SNS message handler.
func StringHandlerToSnsWrapperHandler(handler StringHandlerFn) func(context.Context, SnsWrapper) error {
	return func(ctx context.Context, sw SnsWrapper) error {
		return handler(ctx, sw.Message)
	}
}

// StringHandlerToSqsHandler wraps  a string handler fxn and returns an sqs handler fxn.
func StringHandlerToSqsHandler(handler StringHandlerFn) SqsHandlerFn {
	return func(ctx context.Context, message awstypes.Message) error {
		if message.Body == nil {
			return errors.New("body is nil.")
		}
		return handler(ctx, *message.Body)
	}
}
