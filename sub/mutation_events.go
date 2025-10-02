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

type ProtoMutationEventHandlerFn[T proto.Message] func(context.Context, models.ProtoMutationEvent[T]) error

// NewMutationEventSqsProcessor creates an SQS event processor that reads mutation events
// from an SQS queue, Unmarshal them into protocol buffer messages, and processes them
// using a strongly-typed event handler function.
func NewMutationEventSqsProcessor[T proto.Message](
	svc *sqs.Client,
	queueURL string,
	newMessage func() T,
	handler ProtoMutationEventHandlerFn[T],
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


func MutationEventHandlerToStringHandler[T proto.Message](handler ProtoMutationEventHandlerFn[T], newMessage func() T) StringHandler {
	return func(ctx context.Context, s string) error {
		msg := &models.PublishedProtoMutationEvent{}
		err := json.Unmarshal([]byte(s), msg)
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

		input := models.ProtoMutationEvent[T]{
			EventID:       msg.EventID,
			EventType:     msg.EventType,
			EventTime:     msg.EventTime,
			Source:        msg.Source,
			CorrelationID: msg.CorrelationID,
			ResourceType:  msg.ResourceType,
			ResourceID:    msg.ResourceID,
			UserID:        msg.UserID,
			Reason:        msg.Reason,
			Before:        before,
			After:         after,
			MetaData:      msg.MetaData,
		}

		return handler(ctx, input)
	}
}
