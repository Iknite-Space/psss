package sub

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/rs/zerolog"
)

// HTTPRequestHandlerFn is a function that handles an HTTP request. Note: this is not a "standard" HTTP handler, because
// it's intended to handle requests asynchronously. It is not intended to be used with an HTTP server.
type HTTPRequestHandlerFn func(ctx context.Context, request *http.Request) error

func NewHTTPRequestProcessor(svc *sqs.Client,
	queueURL string,
	handlerFn HTTPRequestHandlerFn,
	s3Client *s3.Client,
	logger zerolog.Logger,
) *SqsEventProcessor {
	return &SqsEventProcessor{
		svc:       svc,
		queueURL:  queueURL,
		handlerFn: httpRequestHandlerToSqsHandlerFn(handlerFn, s3Client, logger),
		logger:    logger,
	}
}

func httpRequestHandlerToSqsHandlerFn(handler HTTPRequestHandlerFn, s3Client *s3.Client, logger zerolog.Logger) SqsHandlerFn {
	return func(ctx context.Context, message awstypes.Message) error {
		if message.Body == nil {
			return nil
		}

		var snsWrapper struct {
			Message string `json:"Message"`
		}

		err := json.Unmarshal([]byte(*message.Body), &snsWrapper)
		if err != nil {
			logger.Err(err).Msg("failed to unmarshal message")
			return fmt.Errorf("failed to unmarshal message: %w", err)
		}

		var s3Message struct {
			Records []struct {
				S3 struct {
					Bucket struct {
						Name string `json:"name"`
					} `json:"bucket"`
					Object struct {
						Key string `json:"key"`
					} `json:"object"`
				} `json:"s3"`
			} `json:"Records"`
		}

		err = json.Unmarshal([]byte(snsWrapper.Message), &s3Message)
		if err != nil {
			logger.Err(err).Msg("failed to unmarshal message")
			return fmt.Errorf("failed to unmarshal message: %w", err)
		}

		for _, record := range s3Message.Records {
			bucketName := record.S3.Bucket.Name
			objectKey := record.S3.Object.Key

			result, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
				Bucket: aws.String(bucketName),
				Key:    aws.String(objectKey),
			})
			if err != nil {
				logger.Err(err).Msg("failed to get object from S3")
				return fmt.Errorf("failed to get object from S3: %w", err)
			}
			defer result.Body.Close()

			buf := bufio.NewReader(result.Body)
			req, err := http.ReadRequest(buf)
			if err != nil {
				logger.Err(err).Msg("failed to read request")
				return fmt.Errorf("failed to read request: %w", err)
			}
			err = handler(ctx, req)
			if err != nil {
				logger.Err(err).Msg("failed to handle request")
				return fmt.Errorf("failed to handle request: %w", err)
			}

		}

		return nil
	}
}
