package sub

import "context"

type SnsWrapper struct {
	Message string `json:"Message"`
}

type StringHandler func(context.Context, string) error

func handleToJSON(handler StringHandler) func(context.Context, SnsWrapper) error {
	return nil
}
