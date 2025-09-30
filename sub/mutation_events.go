package sub

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type EventMutationHandlerFn func(context.Context, []types.Message, string) error

func MutationEventHandlerToStringHandlder(handler EventMutationHandlerFn) StringHandler {
	return func(ctx context.Context, s string) error {
		return nil
	}
}
