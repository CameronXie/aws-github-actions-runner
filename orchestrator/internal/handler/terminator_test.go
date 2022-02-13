package handler

import (
	"context"
	"testing"

	"github.com/CameronXie/aws-github-actions-runner/orchestrator/internal/orchestrator"
	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

func TestSetupTerminatorHandler(t *testing.T) {
	cases := map[string]struct {
		event         events.SQSEvent
		expectedInput *orchestrator.TerminationInput
		eventErrMsg   string
	}{
		"empty sqs messages": {
			event:       events.SQSEvent{Records: make([]events.SQSMessage, 0)},
			eventErrMsg: `sqs event sourcing configure issue, received 0 messages`,
		},
		"more than one sqs messages": {
			event:       events.SQSEvent{Records: make([]events.SQSMessage, 10)},
			eventErrMsg: `sqs event sourcing configure issue, received 10 messages`,
		},
		"one sqs message": {
			event: events.SQSEvent{Records: []events.SQSMessage{
				{Body: `{"ID":1,"Owner":"Owner","Repository":"Repo","Labels":["ec2","ubuntu"]}`},
			}},
			expectedInput: &orchestrator.TerminationInput{
				ID:     1,
				Labels: []string{"ec2", "ubuntu"},
			},
		},
		"invalid sqs message": {
			event: events.SQSEvent{Records: []events.SQSMessage{
				{Body: `{`},
			}},
			eventErrMsg: `unexpected end of JSON input`,
		},
	}

	for n, tc := range cases {
		t.Run(n, func(t *testing.T) {
			a := assert.New(t)
			svc := new(mockedTerminationSVC)
			err := SetupTerminatorHandler(svc)(context.TODO(), tc.event)

			if tc.eventErrMsg == "" {
				a.Nil(err)
				return
			}

			a.Equal(tc.eventErrMsg, err.Error())
			a.Equal(tc.expectedInput, svc.input)
		})
	}
}

type mockedTerminationSVC struct {
	input *orchestrator.TerminationInput
}

func (m *mockedTerminationSVC) Terminate(_ context.Context, input *orchestrator.TerminationInput) error {
	m.input = input
	return nil
}
