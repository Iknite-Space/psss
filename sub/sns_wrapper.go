package sub

import (
	"context"
)

type SnsWrapper struct {
	Message string `json:"Message"`
}

type StringHandler func(context.Context, string) error

func StringHandlerToSnsWrapperHandler(handler StringHandler) func(context.Context, SnsWrapper) error {
	return func(ctx context.Context, sw SnsWrapper) error {
		return handler(ctx, sw.Message)
	}
}
