package sub

import (
	"context"
	"fmt"
)

type SnsWrapper struct {
	Message string `json:"Message"`
}

type StringHandler func(context.Context, string) error

func StringHandlerToSnsWrapperhandler(handler StringHandler) func(context.Context, SnsWrapper) error {
	return func(ctx context.Context, sw SnsWrapper) error {
		msg := sw.Message
		err := handler(ctx, msg)
		if err != nil {
			return fmt.Errorf(`error calling this handler. here's why=%w`, err)
		}
		return nil
	}
}
