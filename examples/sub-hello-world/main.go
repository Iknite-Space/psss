// Example: SQS Message Consumer using psss/sub SqsEventProcessor
//
// This example demonstrates how to consume messages from an AWS SQS queue using the
// psss/sub package's SqsEventProcessor. The script sets up an SQS client, subscribes
// to a specified queue, and processes incoming messages.
//
// The main business logic for handling each message is encapsulated in the `printMessage`
// function, which simply prints the message body to the console. The SqsEventProcessor
// is responsible for polling messages from the queue, invoking the handler for each message,
// and deleting messages from the queue if they are handled successfully.
//
// Usage:
//
//	go run main.go <sqs-queue-url>
//
// AWS region must be specified via the AWS_REGION environment variable.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Iknite-Space/psss/sub"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// printMessage is a callback function that processes a single SQS message.
// It prints the message body to the console.
func printMessage(ctx context.Context, message awstypes.Message) error {
	if message.Body == nil {
		return fmt.Errorf("message body is nil")
	}

	fmt.Println("Received message:", *message.Body)
	return nil
}

// run sets up the AWS SQS client and starts the message processor.
func run() error {
	ctx := context.Background()

	// Ensure the SQS queue URL is provided as a command-line argument.
	if len(os.Args) < 2 {
		return fmt.Errorf("usage: %s <sqs-queue-url>", os.Args[0])
	}

	// Get the AWS region from the environment variable.
	awsRegion := os.Getenv("AWS_REGION")
	sqsURL := os.Args[1]

	// Load the AWS SDK configuration with the specified region.
	awscfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(awsRegion))
	if err != nil {
		return fmt.Errorf("failed to load SDK config, %w", err)
	}

	// Create a new SQS service client.
	sqsSvc := sqs.NewFromConfig(awscfg)

	// Create a new SQS event processor with the SQS client, queue URL, and message handler.
	processor := sub.NewSqsEventProcessor(sqsSvc, sqsURL, printMessage)

	// Run the processor to start receiving and handling messages.
	err = processor.Run(ctx)
	if err != nil {
		return fmt.Errorf("failed to run processor, %w", err)
	}

	return nil
}

// main is the entry point of the application.
// It calls run() and handles any errors by printing them and exiting with a non-zero status.
func main() {
	err := run()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
