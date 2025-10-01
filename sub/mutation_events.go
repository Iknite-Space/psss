package sub

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Iknite-Space/psss/models"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// NewMutationEventSqsProcessor creates an SQS event processor that reads mutation events
// from an SQS queue, Unmarshal them into protocol buffer messages, and processes them
// using a strongly-typed event handler function.
func NewMutationEventSqsProcessor[T proto.Message](
	svc *sqs.Client,
	queueURL string,
	newMessage func() T,
	handler EventMutationProtoHandlerFn[T],
) *SqsEventProcessor {

	snsString := MutationEventHandlerToStringHandler(handler, newMessage)
	snsWrapper := StringHandlerToSnsWrapperHandler(snsString)

	return &SqsEventProcessor{
		svc:       svc,
		queueURL:  queueURL,
		logger:    zerolog.Nop(),
		handlerFn: jsonEventHandlerToSqsHandlerFn(snsWrapper),
	}
}

type EventMutationProtoHandlerFn[T proto.Message] func(context.Context, models.ProtoMutationEvent[T]) error

func MutationEventHandlerToStringHandler[T proto.Message](handler EventMutationProtoHandlerFn[T], newMessage func() T) StringHandler {
	return func(ctx context.Context, s string) error {
		msg := models.MutationEvent{}
		err := json.Unmarshal([]byte(s), &msg)
		if err != nil {
			return fmt.Errorf("error unmarshaling JSON mutation event: %w", err)
		}

		before := newMessage()
		err = protojson.Unmarshal([]byte(msg.Before), before)
		if err != nil {
			return fmt.Errorf("error unmarshaling 'Before' field from protobuf: %w", err)
		}

		after := newMessage()
		err = protojson.Unmarshal([]byte(msg.After), after)
		if err != nil {
			return fmt.Errorf("error unmarshaling 'After' field from protobuf: %w", err)
		}

		meta := newMessage()
		err = protojson.Unmarshal([]byte(msg.MetaData), meta)
		if err != nil {
			return fmt.Errorf("error unmarshaling 'MetaData' field from protobuf: %w", err)
		}

		input := models.ProtoMutationEvent[T]{
			EventID:       msg.EventID,
			EventType:     msg.EventType,
			EventTime:     msg.EventTime,
			Source:        msg.Source,
			CorrelationID: msg.CorrelationID,
			ResourceType:  msg.ResourceType,
			ResourceID:    msg.ResourceID,
			PerformedBy:   msg.PerformedBy,
			Reason:        msg.Reason,
			Before:        before,
			After:         after,
			MetaData:      meta,
		}

		return handler(ctx, input)
	}
}
