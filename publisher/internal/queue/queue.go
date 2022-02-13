package queue

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"golang.org/x/sync/errgroup"
)

type SendMessageAPIClient interface {
	SendMessage(context.Context, *sqs.SendMessageInput, ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
}

type Message struct {
	URL             string
	Body            string
	DeduplicationID string
}

type Queue interface {
	Send(ctx context.Context, messages []Message) error
}

type queue struct {
	client SendMessageAPIClient
}

func (q *queue) Send(ctx context.Context, messages []Message) error {
	g, ctx := errgroup.WithContext(ctx)

	for _, i := range messages {
		msg := i
		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				_, err := q.client.SendMessage(ctx, &sqs.SendMessageInput{
					QueueUrl:    aws.String(msg.URL),
					MessageBody: aws.String(msg.Body),
				})

				return err
			}
		})
	}

	return g.Wait()
}

func New(client SendMessageAPIClient) Queue {
	return &queue{
		client: client,
	}
}
