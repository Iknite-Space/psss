// Example usage:
//	go run main.go <sns-topic-arn>
// Ensure AWS_REGION is set in your environment.

// AWS region must be specified via the AWS_REGION environment variable.
// The SNS topic ARN must be provided as a command-line argument.
// Example usage:
//
//	go run main.go <sns-topic-arn>
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Iknite-Space/psss/models"
	"github.com/Iknite-Space/psss/pub"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"google.golang.org/protobuf/proto"
)

func run() error {
	ctx := context.Background()

	if len(os.Args) < 2 {
		return fmt.Errorf("usage: %s <sns-topic-arn>", os.Args[0])
	}

	awsRegion := os.Getenv("AWS_REGION")
	topicArn := os.Args[1]

	awscfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(awsRegion))
	if err != nil {
		return fmt.Errorf("failed to load SDK config, %w", err)
	}

	snsClient := sns.NewFromConfig(awscfg)
	
	publisher := pub.NewPubService(snsClient, topicArn)

	// Publishing a "created" event
	event := models.ProtoMutationEvent[proto.Message]{
		EventID:       "392f8b1e-4f1c-4d2a-9c3e-1a2b3c4d5e6f",
		EventType:     models.EventTypeCreated,
		EventTime:     time.Now(),
		Source:        "pub-hello-world",
		ResourceType:  "example-resource",
		ResourceID:    "resource-1",
		UserID:        "432f8b1e-4f1c-4d2a-9c3e-1a2b3c4d5e6f",
		CorrelationID: "432f8b1e-4f1c-4d2a-9c3e-1a2b3c4d5e6f",
		Reason:        "testing pub-hello-world example",
	}

	err = publisher.Publish(ctx, event)
	if err != nil {
		return fmt.Errorf("failed to publish message, %w", err)
	}

	fmt.Println("Message published successfully")

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
