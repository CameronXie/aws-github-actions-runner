package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/CameronXie/aws-github-actions-runner/orchestrator/pkg/runner"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"go.uber.org/zap/zaptest/observer"
)

func TestTerminationService_Terminate(t *testing.T) {
	cases := map[string]struct {
		terminationInput *TerminationInput
		matchErr         error
		terminateErr     error
	}{
		"terminate runner": {
			terminationInput: &TerminationInput{
				ID:     1,
				Labels: []string{SelfHostedLabel, EC2Label},
			},
		},
		"not terminator matches input labels": {
			terminationInput: &TerminationInput{
				ID:     1,
				Labels: []string{SelfHostedLabel, EKSLabel},
			},
			matchErr: fmt.Errorf(
				`labels (%v) termination currently is not supported`,
				[]string{SelfHostedLabel, EKSLabel},
			),
		},
		"runner not exists": {
			terminationInput: &TerminationInput{
				ID:     1,
				Labels: []string{SelfHostedLabel, EC2Label},
			},
			terminateErr: new(runner.NotExistsError),
		},
		"failed to terminate runner": {
			terminationInput: &TerminationInput{
				ID:     1,
				Labels: []string{SelfHostedLabel, EC2Label},
			},
			terminateErr: errors.New("failed to terminate runner"),
		},
	}

	for n, tc := range cases {
		t.Run(n, func(t *testing.T) {
			a := assert.New(t)
			terminator := &mockedRunnerTerminator{terminateErr: tc.terminateErr}
			core, logs := observer.New(zap.DebugLevel)
			svc := &terminationService{
				terminators: []*Terminator{
					{
						Target:     EC2Label,
						Terminator: terminator,
					},
				},
				logger: zap.New(core),
			}

			err := svc.Terminate(context.TODO(), tc.terminationInput)

			if tc.matchErr != nil {
				a.Equal(tc.matchErr, err)
				a.Equal(0, len(logs.All()))
				return
			}

			a.Equal(fmt.Sprintf(
				`id: %v, labels: %v, target: %v\n`,
				tc.terminationInput.ID,
				strings.Join(tc.terminationInput.Labels, ","),
				EC2Label,
			), logs.All()[0].Message)

			if tc.terminateErr != nil && runner.IsNotExistsError(tc.terminateErr) {
				a.Equal(fmt.Sprintf(
					`runner with ID (%v) not exists`, tc.terminationInput.ID,
				), logs.All()[1].Message)
				a.Nil(err)
				return
			}

			a.Equal(tc.terminateErr, err)
		})
	}
}

func TestNewTerminationSvc(t *testing.T) {
	a := assert.New(t)
	logger := zaptest.NewLogger(t)
	expectedTargets := []string{EC2Label, EKSLabel}
	svc := NewTerminationSvc(
		nil,
		nil,
		nil,
		logger,
	).(*terminationService)

	targets := make([]string, 0)
	for i := range svc.terminators {
		targets = append(targets, svc.terminators[i].Target)
	}

	a.Equal(expectedTargets, targets)
	a.Equal(logger, svc.logger)
}
