package sub

import (
	"context"
)

type EventMutationHandlerFn func(context.Context, MutationEvents) error

func MutationEventHandlerToStringHandlder(handler EventMutationHandlerFn) StringHandler {
	return func(ctx context.Context, s string) error {
		return nil
	}
}
