package handler

import (
	"context"
	"testing"

	"github.com/CameronXie/aws-github-actions-runner/orchestrator/pkg/runner"
	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

// nolint:dupl
func TestSetupLauncherHandler(t *testing.T) {
	cases := map[string]struct {
		event         events.SQSEvent
		expectedInput *runner.LaunchInput
		logs          []string
		launchErr     error
		errMsg        string
	}{
		"empty sqs messages": {
			event:  events.SQSEvent{Records: make([]events.SQSMessage, 0)},
			logs:   make([]string, 0),
			errMsg: `sqs event sourcing configure issue, received 0 messages`,
		},
		"more than one sqs messages": {
			event:  events.SQSEvent{Records: make([]events.SQSMessage, 10)},
			logs:   make([]string, 0),
			errMsg: `sqs event sourcing configure issue, received 10 messages`,
		},
		"invalid sqs message": {
			event: events.SQSEvent{Records: []events.SQSMessage{
				{Body: `{`},
			}},
			errMsg: `unexpected end of JSON input`,
		},
		"one sqs message": {
			event: events.SQSEvent{Records: []events.SQSMessage{
				{Body: `{"Message":"{\"ID\":1,\"Owner\":\"owner\",\"Repository\":\"repo\",\"Labels\":[\"ec2\",\"ubuntu\"]}"}`},
			}},
			expectedInput: &runner.LaunchInput{
				ID:         1,
				Owner:      "Owner",
				Repository: "Repo",
				Labels:     []string{"ec2", "ubuntu"},
			},
			logs: []string{"launching runner with ID (1)"},
		},
		"runner exists": {
			event: events.SQSEvent{Records: []events.SQSMessage{
				{Body: `{"Message":"{\"ID\":1,\"Owner\":\"owner\",\"Repository\":\"repo\",\"Labels\":[\"ec2\",\"ubuntu\"]}"}`},
			}},
			launchErr: &runner.AlreadyExistsError{ID: 1},
			logs:      []string{"launching runner with ID (1)", "runner with ID (1) already exists"},
		},
	}

	for n, tc := range cases {
		t.Run(n, func(t *testing.T) {
			a := assert.New(t)
			launcher := &mockedLauncher{launchErr: tc.launchErr}
			core, logs := observer.New(zap.DebugLevel)
			err := SetupLauncherHandler(launcher, zap.New(core))(context.TODO(), tc.event)

			l := make([]string, 0)
			for _, i := range logs.All() {
				l = append(l, i.Message)
			}
			a.ElementsMatch(tc.logs, l)

			if tc.errMsg == "" {
				a.Nil(err)
				return
			}

			a.Equal(tc.errMsg, err.Error())
			a.Equal(tc.expectedInput, launcher.input)
		})
	}
}

type mockedLauncher struct {
	input     *runner.LaunchInput
	launchErr error
}

func (m *mockedLauncher) Launch(_ context.Context, input *runner.LaunchInput) error {
	m.input = input
	return m.launchErr
}
