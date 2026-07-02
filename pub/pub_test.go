package pub

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Iknite-Space/psss/models"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	snstypes "github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

type mockSNSBatchClient struct {
	publishBatchFn func(ctx context.Context, params *sns.PublishBatchInput, optFns ...func(*sns.Options)) (*sns.PublishBatchOutput, error)
}

func (m *mockSNSBatchClient) PublishBatch(ctx context.Context, params *sns.PublishBatchInput, optFns ...func(*sns.Options)) (*sns.PublishBatchOutput, error) {
	return m.publishBatchFn(ctx, params, optFns...)
}

func makeTestEvent(eventID, correlationID string) models.ProtoMutationEvent[proto.Message] {
	return models.ProtoMutationEvent[proto.Message]{
		EventID:       eventID,
		EventType:     models.EventTypeCreated,
		EventTime:     time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		Source:        "test-source",
		CorrelationID: correlationID,
		ResourceType:  "test-resource",
		ResourceID:    "res-1",
		UserID:        "user-1",
		Reason:        "testing",
		Before:        &structpb.Value{Kind: &structpb.Value_NullValue{}},
		After:         &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "after"}},
	}
}

func TestNewSNSBatchPublisher(t *testing.T) {
	client := &mockSNSBatchClient{}
	p := NewSNSBatchPublisher[proto.Message](client, "arn:aws:sns:us-east-1:123456789012:test-topic")
	if p == nil {
		t.Fatal("expected non-nil publisher")
	}
	if p.topicArn != "arn:aws:sns:us-east-1:123456789012:test-topic" {
		t.Fatalf("expected topic arn, got %q", p.topicArn)
	}
}

func TestSNSBatchPublisher_PublishBatch_Success(t *testing.T) {
	client := &mockSNSBatchClient{
		publishBatchFn: func(ctx context.Context, params *sns.PublishBatchInput, optFns ...func(*sns.Options)) (*sns.PublishBatchOutput, error) {
			if aws.ToString(params.TopicArn) != "arn:aws:sns:us-east-1:123456789012:test-topic" {
				t.Errorf("unexpected topic arn: %s", aws.ToString(params.TopicArn))
			}
			if len(params.PublishBatchRequestEntries) != 2 {
				t.Errorf("expected 2 entries, got %d", len(params.PublishBatchRequestEntries))
			}
			return &sns.PublishBatchOutput{
				Successful: []snstypes.PublishBatchResultEntry{
					{Id: aws.String("evt-1"), MessageId: aws.String("msg-1")},
					{Id: aws.String("evt-2"), MessageId: aws.String("msg-2")},
				},
			}, nil
		},
	}

	p := NewSNSBatchPublisher[proto.Message](client, "arn:aws:sns:us-east-1:123456789012:test-topic")

	messages := []models.ProtoMutationEvent[proto.Message]{
		makeTestEvent("evt-1", "corr-1"),
		makeTestEvent("evt-2", "corr-2"),
	}

	count, err := p.PublishBatch(context.Background(), messages)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 successful, got %d", count)
	}
}

func TestSNSBatchPublisher_PublishBatch_EmptyBatch(t *testing.T) {
	client := &mockSNSBatchClient{}
	p := NewSNSBatchPublisher[proto.Message](client, "arn:aws:sns:us-east-1:123456789012:test-topic")

	count, err := p.PublishBatch(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0, got %d", count)
	}
}

func TestSNSBatchPublisher_PublishBatch_TotalFailure(t *testing.T) {
	client := &mockSNSBatchClient{
		publishBatchFn: func(ctx context.Context, params *sns.PublishBatchInput, optFns ...func(*sns.Options)) (*sns.PublishBatchOutput, error) {
			return nil, errors.New("service unavailable")
		},
	}

	p := NewSNSBatchPublisher[proto.Message](client, "arn:aws:sns:us-east-1:123456789012:test-topic")

	messages := []models.ProtoMutationEvent[proto.Message]{
		makeTestEvent("evt-1", "corr-1"),
	}

	_, err := p.PublishBatch(context.Background(), messages)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSNSBatchPublisher_PublishBatch_PartialFailure(t *testing.T) {
	client := &mockSNSBatchClient{
		publishBatchFn: func(ctx context.Context, params *sns.PublishBatchInput, optFns ...func(*sns.Options)) (*sns.PublishBatchOutput, error) {
			return &sns.PublishBatchOutput{
				Successful: []snstypes.PublishBatchResultEntry{
					{Id: aws.String("evt-1"), MessageId: aws.String("msg-1")},
				},
				Failed: []snstypes.BatchResultErrorEntry{
					{Id: aws.String("evt-2"), Code: aws.String("InternalError"), SenderFault: false, Message: aws.String("internal error")},
				},
			}, nil
		},
	}

	p := NewSNSBatchPublisher[proto.Message](client, "arn:aws:sns:us-east-1:123456789012:test-topic")

	messages := []models.ProtoMutationEvent[proto.Message]{
		makeTestEvent("evt-1", "corr-1"),
		makeTestEvent("evt-2", "corr-2"),
	}

	count, err := p.PublishBatch(context.Background(), messages)
	if err == nil {
		t.Fatal("expected partial failure error, got nil")
	}
	if count != 1 {
		t.Fatalf("expected 1 successful, got %d", count)
	}
}

func TestSNSBatchPublisher_WithLogger(t *testing.T) {
	client := &mockSNSBatchClient{}
	p := NewSNSBatchPublisher[proto.Message](client, "arn:aws:sns:us-east-1:123456789012:test-topic")
	p2 := p.WithLogger(p.logger)
	if p != p2 {
		t.Fatal("WithLogger should return the same instance")
	}
}

func TestSNSBatchPublisher_PublishBatch_ExceedsMaxSize(t *testing.T) {
	client := &mockSNSBatchClient{}
	p := NewSNSBatchPublisher[proto.Message](client, "arn:aws:sns:us-east-1:123456789012:test-topic")

	messages := make([]models.ProtoMutationEvent[proto.Message], MaxBatchSize+1)
	for i := range messages {
		messages[i] = makeTestEvent("evt-1", "corr-1")
	}

	_, err := p.PublishBatch(context.Background(), messages)
	if err == nil {
		t.Fatal("expected error for exceeding batch size, got nil")
	}
}

func TestSNSBatchPublisher_PublishBatch_AtMaxSize(t *testing.T) {
	client := &mockSNSBatchClient{
		publishBatchFn: func(ctx context.Context, params *sns.PublishBatchInput, optFns ...func(*sns.Options)) (*sns.PublishBatchOutput, error) {
			successful := make([]snstypes.PublishBatchResultEntry, len(params.PublishBatchRequestEntries))
			for i := range successful {
				successful[i] = snstypes.PublishBatchResultEntry{Id: params.PublishBatchRequestEntries[i].Id, MessageId: aws.String("msg")}
			}
			return &sns.PublishBatchOutput{Successful: successful}, nil
		},
	}
	p := NewSNSBatchPublisher[proto.Message](client, "arn:aws:sns:us-east-1:123456789012:test-topic")

	messages := make([]models.ProtoMutationEvent[proto.Message], MaxBatchSize)
	for i := range messages {
		messages[i] = makeTestEvent("evt-1", "corr-1")
	}

	count, err := p.PublishBatch(context.Background(), messages)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != MaxBatchSize {
		t.Fatalf("expected %d, got %d", MaxBatchSize, count)
	}
}

func TestSQSBatchPublisher_SendBatch_ExceedsMaxSize(t *testing.T) {
	p := NewSQSBatchPublisher(nil)

	entries := make([]types.SendMessageBatchRequestEntry, MaxBatchSize+1)
	_, err := p.SendBatch(context.Background(), "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue", entries)
	if err == nil {
		t.Fatal("expected error for exceeding batch size, got nil")
	}
}
