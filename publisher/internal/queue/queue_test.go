package queue

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/stretchr/testify/assert"
)

func TestQueue_Send(t *testing.T) {
	cases := map[string]struct {
		messages      []Message
		msgErr        error
		expectedInput []sqs.SendMessageInput
		err           error
	}{
		"send messages": {
			messages: []Message{
				{
					URL:             "url_1",
					Body:            "msg_1",
					DeduplicationID: "1",
				},
				{
					URL:             "url_2",
					Body:            "msg_2",
					DeduplicationID: "2",
				},
			},
			expectedInput: []sqs.SendMessageInput{
				{
					QueueUrl:    aws.String("url_1"),
					MessageBody: aws.String("msg_1"),
				},
				{
					QueueUrl:    aws.String("url_2"),
					MessageBody: aws.String("msg_2"),
				},
			},
		},
		"failed to send message": {
			messages: []Message{
				{
					URL:             "url_1",
					Body:            "msg_1",
					DeduplicationID: "1",
				},
				{
					URL:             "url_2",
					Body:            "msg_2",
					DeduplicationID: "2",
				},
			},
			msgErr: errors.New("failed to send message"),
			err:    errors.New("failed to send message"),
		},
	}

	for n, tc := range cases {
		t.Run(n, func(t *testing.T) {
			a := assert.New(t)
			c := &mockedSendMessageAPIClient{msgErr: tc.msgErr}

			a.Equal(tc.err, New(c).Send(context.TODO(), tc.messages))
			if tc.err == nil {
				a.ElementsMatch(tc.expectedInput, c.input)
				return
			}

			a.GreaterOrEqual(len(c.input), 1)
			a.LessOrEqual(len(c.input), len(tc.messages))
		})
	}
}

type mockedSendMessageAPIClient struct {
	sync.RWMutex
	msgErr error
	input  []sqs.SendMessageInput
}

func (m *mockedSendMessageAPIClient) SendMessage(
	_ context.Context,
	input *sqs.SendMessageInput,
	_ ...func(*sqs.Options),
) (*sqs.SendMessageOutput, error) {
	m.Lock()
	defer m.Unlock()

	m.input = append(m.input, *input)
	return nil, m.msgErr
}
