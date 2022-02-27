package messenger

import (
	"context"
	"encoding/json"
	"math"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"golang.org/x/sync/errgroup"
)

const (
	hostAttribute   = "Host"
	osAttribute     = "OS"
	statusAttribute = "Status"
	messageSource   = "Publisher"
	snsBatchSize    = 10
)

type Message struct {
	Host   string
	OS     string
	Status string
	Body   string
}

type Messenger interface {
	PublishJobs(ctx context.Context, messages []Message) error
	NotifyPublisher(ctx context.Context) error
}

type PublishAPIClient interface {
	Publish(ctx context.Context, params *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error)
	PublishBatch(ctx context.Context, params *sns.PublishBatchInput, optFns ...func(*sns.Options)) (*sns.PublishBatchOutput, error)
}

type messenger struct {
	client         PublishAPIClient
	jobsTopic      string
	publisherTopic string
}

func (n *messenger) PublishJobs(ctx context.Context, messages []Message) error {
	entriesList := partitionJobs(messages, snsBatchSize)
	g, ctx := errgroup.WithContext(ctx)

	for i := range entriesList {
		entries := entriesList[i]
		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				_, err := n.client.PublishBatch(ctx, &sns.PublishBatchInput{
					PublishBatchRequestEntries: entries,
					TopicArn:                   aws.String(n.jobsTopic),
				})

				return err
			}
		})
	}

	return g.Wait()
}

func (n *messenger) NotifyPublisher(ctx context.Context) error {
	_, err := n.client.Publish(ctx, &sns.PublishInput{
		TopicArn: aws.String(n.publisherTopic),
		Message:  aws.String(getMessage(messageSource)),
	})

	return err
}

func partitionJobs(messages []Message, batchSize int) [][]types.PublishBatchRequestEntry {
	jobsNum := len(messages)
	entries := make([][]types.PublishBatchRequestEntry, 0)

	for i := 0; i < jobsNum; i += batchSize {
		minNum := int(math.Min(float64(i+batchSize), float64(jobsNum)))
		entries = append(entries, toPublishBatchRequestEntry(messages[i:minNum]))
	}

	return entries
}

func toPublishBatchRequestEntry(messages []Message) []types.PublishBatchRequestEntry {
	entries := make([]types.PublishBatchRequestEntry, 0)
	for i, m := range messages {
		entries = append(entries, types.PublishBatchRequestEntry{
			Id:      aws.String(strconv.Itoa(i)),
			Message: aws.String(m.Body),
			MessageAttributes: map[string]types.MessageAttributeValue{
				hostAttribute: {
					DataType:    aws.String("String"),
					StringValue: aws.String(m.Host),
				},
				osAttribute: {
					DataType:    aws.String("String"),
					StringValue: aws.String(m.OS),
				},
				statusAttribute: {
					DataType:    aws.String("String"),
					StringValue: aws.String(m.Status),
				},
			},
		})
	}

	return entries
}

func getMessage(source string) string {
	b, _ := json.Marshal(struct {
		Source string
	}{
		Source: source,
	})

	return string(b)
}

func New(
	client PublishAPIClient,
	jobsTopic string,
	publisherTopic string,
) Messenger {
	return &messenger{
		client:         client,
		jobsTopic:      jobsTopic,
		publisherTopic: publisherTopic,
	}
}
