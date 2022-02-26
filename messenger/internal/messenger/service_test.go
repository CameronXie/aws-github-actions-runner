package messenger

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/stretchr/testify/assert"
)

func TestService_NotifyPublisher(t *testing.T) {
	publisherTopic := "publisher"
	cases := map[string]struct {
		notifyPublisherErr error
		expectedInput      *sns.PublishInput
		err                error
	}{
		"notify publisher": {
			expectedInput: &sns.PublishInput{
				TopicArn: aws.String(publisherTopic),
				Message:  aws.String(`{"Source":"DynamoDBStream"}`),
			},
		},
		"failed to notify publisher": {
			notifyPublisherErr: errors.New("failed to notify publisher"),
			err:                errors.New("failed to notify publisher"),
		},
	}

	for n, tc := range cases {
		t.Run(n, func(t *testing.T) {
			a := assert.New(t)
			mockedClient := &mockedPublishAPIClient{
				notifyPublisherErr: tc.notifyPublisherErr,
			}
			m := New(mockedClient, publisherTopic)

			a.Equal(tc.err, m.NotifyPublisher(context.TODO()))
			if tc.err == nil {
				a.EqualValues(tc.expectedInput, mockedClient.notification)
			}
		})
	}
}

type mockedPublishAPIClient struct {
	notification       *sns.PublishInput
	notifyPublisherErr error
}

func (m *mockedPublishAPIClient) Publish(_ context.Context,
	params *sns.PublishInput,
	_ ...func(*sns.Options),
) (*sns.PublishOutput, error) {
	m.notification = params

	return nil, m.notifyPublisherErr
}
