package handler

import (
	"context"
	"testing"

	"github.com/CameronXie/aws-github-actions-runner/orchestrator/pkg/runner"
	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

func TestSetupLauncherHandler(t *testing.T) {
	cases := map[string]struct {
		event         events.SQSEvent
		expectedInput *runner.LaunchInput
		errMsg        string
	}{
		"empty sqs messages": {
			event:  events.SQSEvent{Records: make([]events.SQSMessage, 0)},
			errMsg: `sqs event sourcing configure issue, received 0 messages`,
		},
		"more than one sqs messages": {
			event:  events.SQSEvent{Records: make([]events.SQSMessage, 10)},
			errMsg: `sqs event sourcing configure issue, received 10 messages`,
		},
		"one sqs message": {
			event: events.SQSEvent{Records: []events.SQSMessage{
				{Body: `{"ID":1,"Owner":"Owner","Repository":"Repo","Labels":["ec2","ubuntu"]}`},
			}},
			expectedInput: &runner.LaunchInput{
				ID:         1,
				Owner:      "Owner",
				Repository: "Repo",
				Labels:     []string{"ec2", "ubuntu"},
			},
		},
		"invalid sqs message": {
			event: events.SQSEvent{Records: []events.SQSMessage{
				{Body: `{`},
			}},
			errMsg: `unexpected end of JSON input`,
		},
	}

	for n, tc := range cases {
		t.Run(n, func(t *testing.T) {
			a := assert.New(t)
			svc := new(mockedLaunchSVC)
			err := SetupLauncherHandler(svc)(context.TODO(), tc.event)

			if tc.errMsg == "" {
				a.Nil(err)
				return
			}

			a.Equal(tc.errMsg, err.Error())
			a.Equal(tc.expectedInput, svc.input)
		})
	}
}

type mockedLaunchSVC struct {
	input *runner.LaunchInput
}

func (m *mockedLaunchSVC) Launch(_ context.Context, input *runner.LaunchInput) error {
	m.input = input
	return nil
}
