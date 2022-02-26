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
func TestSetupTerminatorHandler(t *testing.T) {
	cases := map[string]struct {
		event           events.SQSEvent
		terminationErr  error
		logs            []string
		expectedInputID uint64
		eventErrMsg     string
	}{
		"empty sqs messages": {
			event:       events.SQSEvent{Records: make([]events.SQSMessage, 0)},
			logs:        make([]string, 0),
			eventErrMsg: `sqs event sourcing configure issue, received 0 messages`,
		},
		"more than one sqs messages": {
			event:       events.SQSEvent{Records: make([]events.SQSMessage, 10)},
			logs:        make([]string, 0),
			eventErrMsg: `sqs event sourcing configure issue, received 10 messages`,
		},
		"invalid sqs message": {
			event: events.SQSEvent{Records: []events.SQSMessage{
				{Body: `{`},
			}},
			eventErrMsg: `unexpected end of JSON input`,
		},
		"one sqs message": {
			event: events.SQSEvent{Records: []events.SQSMessage{
				{Body: `{"Message":"{\"ID\":1,\"Owner\":\"owner\",\"Repository\":\"repo\",\"Labels\":[\"ec2\",\"ubuntu\"]}"}`},
			}},
			expectedInputID: 1,
			logs:            []string{"terminating runner with ID (1)"},
		},
		"runner not exists": {
			event: events.SQSEvent{Records: []events.SQSMessage{
				{Body: `{"Message":"{\"ID\":1,\"Owner\":\"owner\",\"Repository\":\"repo\",\"Labels\":[\"ec2\",\"ubuntu\"]}"}`},
			}},
			terminationErr: &runner.NotExistsError{ID: 1},
			logs:           []string{"terminating runner with ID (1)", "runner with ID (1) not exists"},
		},
	}

	for n, tc := range cases {
		t.Run(n, func(t *testing.T) {
			a := assert.New(t)
			terminator := &mockedTerminator{terminationErr: tc.terminationErr}
			core, logs := observer.New(zap.DebugLevel)
			err := SetupTerminatorHandler(terminator, zap.New(core))(context.TODO(), tc.event)

			l := make([]string, 0)
			for _, i := range logs.All() {
				l = append(l, i.Message)
			}
			a.ElementsMatch(tc.logs, l)

			if tc.eventErrMsg == "" {
				a.Nil(err)
				return
			}

			a.Equal(tc.eventErrMsg, err.Error())
			a.Equal(tc.expectedInputID, terminator.inputID)
		})
	}
}

type mockedTerminator struct {
	inputID        uint64
	terminationErr error
}

func (m *mockedTerminator) Terminate(_ context.Context, id uint64) error {
	m.inputID = id
	return m.terminationErr
}
