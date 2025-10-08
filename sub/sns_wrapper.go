package sub

import (
	"context"
)

type SnsWrapper struct {
	Message string `json:"Message"`
}

type SnsWrapperHandler func(context.Context, string) error

// SnsWrapperToSqsWrapperHandler wraps an SNS message payload into a handler that extracts
// the 'Message' field and passes it to the provided SNS message handler.
func SnsWrapperToSqsWrapperHandler(handler SnsWrapperHandler) func(context.Context, SnsWrapper) error {
	return func(ctx context.Context, sw SnsWrapper) error {
		return handler(ctx, sw.Message)
	}
}