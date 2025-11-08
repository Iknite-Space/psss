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
// from an SQS queue, unmarshal them into strongly typed protobuf messages, and processes them
// using the provided event handler. 
// Note :- if includeSnsWrapper is true, the processor expects SQS messages to be wrapped in SNS JSON format.
// Otherwise, it expects direct SQS messages containing the SNS "Message" JSON.
func NewMutationEventSqsProcessor[T proto.Message](svc *sqs.Client, queueURL string, newMessage func() T, handler ProtoMutationEventHandlerFn[T], includeSnsWrapper bool) *SqsEventProcessor {
	snsStringHandler := MutationEventHandlerToStringHandler(handler, newMessage)

	var sqsHandler SqsHandlerFn
	if includeSnsWrapper {
		// handling wrapped sns messages in sns format, see example in 'SnsWrapper' struct.
		snsWrapperHandler := SnsWrapperToSqsWrapperHandler(snsStringHandler)
		sqsHandler = jsonEventHandlerToSqsHandlerFn(snsWrapperHandler)
	} else {
		// handling direct SQS messages containing the SNS "Message" Json field.
		sqsHandler = jsonEventHandlerToSqsHandlerFn(
			func(ctx context.Context, msg string) error {
				return snsStringHandler(ctx, msg)
			},
		)
	}

	return &SqsEventProcessor{
		svc:       svc,
		queueURL:  queueURL,
		logger:    zerolog.Nop(),
		handlerFn: sqsHandler,
	}
}

// MutationEventHandlerToStringHandler converts a strongly typed ProtoMutationEventHandlerFn
// into an SNS-compatible handler that processes the message field of an SNS JSON payload.
// It deserializes the incoming mutation SNS event message into a PublishedProtoMutationEvent,
// unmarshal the "Before" and "After" protobuf messages, and invokes the provided handler.
// Returns an error if JSON or protobuf unmarshaling fails.
func MutationEventHandlerToStringHandler[T proto.Message](handler ProtoMutationEventHandlerFn[T], newMessage func() T) SnsWrapperHandler {
	return func(ctx context.Context, s string) error {
		msg := &models.PublishedProtoMutationEvent{}
		err := json.Unmarshal([]byte(s), msg)
		if err != nil {
			return fmt.Errorf("error unmarshaling sns mutation event. why=%w", err)
		}

		before := newMessage()
		err = protojson.Unmarshal([]byte(msg.Before), before)
		if err != nil {
			return fmt.Errorf("error unmarshaling 'Before' field from sns mutation event. why=%w", err)
		}

		after := newMessage()
		err = protojson.Unmarshal([]byte(msg.After), after)
		if err != nil {
			return fmt.Errorf("error unmarshaling 'After' field from sns mutation event. why=%w", err)
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
