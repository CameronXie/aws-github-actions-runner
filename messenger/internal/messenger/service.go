package messenger

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

const (
	messageSource = "DynamoDBStream"
)

type Service interface {
	NotifyPublisher(ctx context.Context) error
}

type PublishAPIClient interface {
	Publish(ctx context.Context, params *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error)
}

type service struct {
	client         PublishAPIClient
	publisherTopic string
}

func (svc service) NotifyPublisher(ctx context.Context) error {
	_, err := svc.client.Publish(ctx, &sns.PublishInput{
		TopicArn: aws.String(svc.publisherTopic),
		Message:  aws.String(getMessage(messageSource)),
	})

	return err
}

func getMessage(source string) string {
	b, _ := json.Marshal(struct {
		Source string
	}{
		Source: source,
	})

	return string(b)
}

func New(client PublishAPIClient, publisherTopic string) Service {
	return &service{client: client, publisherTopic: publisherTopic}
}
