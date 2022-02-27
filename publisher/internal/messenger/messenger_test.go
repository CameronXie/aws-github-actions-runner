package messenger

import (
	"context"
	"errors"
	"math"
	"strconv"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/stretchr/testify/assert"
)

func TestMessenger_PublishJobs(t *testing.T) {
	jobsTopic := "jobs"
	cases := map[string]struct {
		messages       []Message
		publishJobsErr error
		expectedJobs   []sns.PublishBatchInput
		err            error
	}{
		"publish jobs": {
			messages: getTestMessages(12),
			expectedJobs: []sns.PublishBatchInput{
				{
					PublishBatchRequestEntries: toPublishBatchRequestEntry(getTestMessages(10)),
					TopicArn:                   aws.String(jobsTopic),
				},
				{
					PublishBatchRequestEntries: toPublishBatchRequestEntry(getTestMessages(2)),
					TopicArn:                   aws.String(jobsTopic),
				},
			},
		},
		"failed to publish jobs": {
			messages:       getTestMessages(12),
			publishJobsErr: errors.New("failed to publish jobs"),
			err:            errors.New("failed to publish jobs"),
		},
	}

	for n, tc := range cases {
		t.Run(n, func(t *testing.T) {
			a := assert.New(t)
			mockedClient := &mockedPublishAPIClient{
				publishJobsErr: tc.publishJobsErr,
			}
			m := New(mockedClient, jobsTopic, "")

			a.Equal(tc.err, m.PublishJobs(context.TODO(), tc.messages))

			if tc.err == nil {
				a.ElementsMatch(tc.expectedJobs, mockedClient.jobs)
				return
			}

			a.GreaterOrEqual(len(mockedClient.jobs), 1)
			a.LessOrEqual(float64(len(mockedClient.jobs)), math.Ceil(float64(len(tc.messages))/snsBatchSize))
		})
	}
}

func TestMessenger_toPublishBatchRequestEntry(t *testing.T) {
	cases := map[string]struct {
		messages      []Message
		expectedEntry []types.PublishBatchRequestEntry
	}{
		"Message to PublishBatchRequestEntry": {
			messages: getTestMessages(2),
			expectedEntry: []types.PublishBatchRequestEntry{
				{
					Id:      aws.String(strconv.Itoa(0)),
					Message: aws.String("msg"),
					MessageAttributes: map[string]types.MessageAttributeValue{
						hostAttribute: {
							DataType:    aws.String("String"),
							StringValue: aws.String("ec2"),
						},
						osAttribute: {
							DataType:    aws.String("String"),
							StringValue: aws.String("ubuntu"),
						},
						statusAttribute: {
							DataType:    aws.String("String"),
							StringValue: aws.String("completed"),
						},
					},
				},
				{
					Id:      aws.String(strconv.Itoa(1)),
					Message: aws.String("msg"),
					MessageAttributes: map[string]types.MessageAttributeValue{
						hostAttribute: {
							DataType:    aws.String("String"),
							StringValue: aws.String("ec2"),
						},
						osAttribute: {
							DataType:    aws.String("String"),
							StringValue: aws.String("ubuntu"),
						},
						statusAttribute: {
							DataType:    aws.String("String"),
							StringValue: aws.String("completed"),
						},
					},
				},
			},
		},
	}

	for n, tc := range cases {
		t.Run(n, func(t *testing.T) {
			a := assert.New(t)
			a.Equal(tc.expectedEntry, toPublishBatchRequestEntry(tc.messages))
		})
	}
}

func TestMessenger_NotifyPublisher(t *testing.T) {
	publisherTopic := "publisher"
	cases := map[string]struct {
		notifyPublisherErr error
		expectedInput      *sns.PublishInput
		err                error
	}{
		"notify publisher": {
			expectedInput: &sns.PublishInput{
				TopicArn: aws.String(publisherTopic),
				Message:  aws.String(`{"Source":"Publisher"}`),
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
			m := New(mockedClient, "", publisherTopic)

			a.Equal(tc.err, m.NotifyPublisher(context.TODO()))
			if tc.err == nil {
				a.EqualValues(tc.expectedInput, mockedClient.notification)
			}
		})
	}
}

func getTestMessages(n int) []Message {
	res := make([]Message, 0)
	for i := 0; i < n; i++ {
		res = append(res, Message{
			Host:   "ec2",
			OS:     "ubuntu",
			Status: "completed",
			Body:   "msg",
		})
	}

	return res
}

type mockedPublishAPIClient struct {
	sync.RWMutex
	jobs               []sns.PublishBatchInput
	notification       *sns.PublishInput
	publishJobsErr     error
	notifyPublisherErr error
}

func (m *mockedPublishAPIClient) Publish(
	_ context.Context,
	params *sns.PublishInput,
	_ ...func(*sns.Options),
) (*sns.PublishOutput, error) {
	m.Lock()
	defer m.Unlock()

	m.notification = params
	return nil, m.notifyPublisherErr
}

func (m *mockedPublishAPIClient) PublishBatch(
	_ context.Context,
	params *sns.PublishBatchInput,
	_ ...func(*sns.Options),
) (*sns.PublishBatchOutput, error) {
	m.Lock()
	defer m.Unlock()

	m.jobs = append(m.jobs, *params)
	return nil, m.publishJobsErr
}
