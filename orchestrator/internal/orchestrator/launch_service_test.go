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

func TestLaunchService_Launch(t *testing.T) {
	cases := map[string]struct {
		launchInput *runner.LaunchInput
		matchErr    error
		launchErr   error
	}{
		"launch runner": {
			launchInput: &runner.LaunchInput{
				ID:     1,
				Labels: []string{SelfHostedLabel, EC2Label},
			},
		},
		"not launcher matches input labels": {
			launchInput: &runner.LaunchInput{
				ID:     1,
				Labels: []string{SelfHostedLabel, EKSLabel},
			},
			matchErr: fmt.Errorf(
				`labels (%v) launch currently is not supported`,
				[]string{SelfHostedLabel, EKSLabel},
			),
		},
		"runner already exists": {
			launchInput: &runner.LaunchInput{
				ID:     1,
				Labels: []string{SelfHostedLabel, EC2Label},
			},
			launchErr: new(runner.AlreadyExistsError),
		},
		"failed to launch runner": {
			launchInput: &runner.LaunchInput{
				ID:     1,
				Labels: []string{SelfHostedLabel, EC2Label},
			},
			launchErr: errors.New("failed to launch runner"),
		},
	}

	for n, tc := range cases {
		t.Run(n, func(t *testing.T) {
			a := assert.New(t)
			launcher := &mockedRunnerLauncher{launchErr: tc.launchErr}
			core, logs := observer.New(zap.DebugLevel)
			svc := &launchService{
				launchers: []*Launcher{
					{
						Target:   []string{EC2Label},
						Launcher: launcher,
					},
				},
				logger: zap.New(core),
			}

			err := svc.Launch(context.TODO(), tc.launchInput)

			if tc.matchErr != nil {
				a.Equal(tc.matchErr, err)
				a.Equal(0, len(logs.All()))
				return
			}

			a.Equal(fmt.Sprintf(
				`id: %v, labels: %v, target: %v\n`,
				tc.launchInput.ID,
				strings.Join(tc.launchInput.Labels, ","),
				EC2Label,
			), logs.All()[0].Message)

			if tc.launchErr != nil && runner.IsAlreadyExistsError(tc.launchErr) {
				a.Equal(fmt.Sprintf(
					`runner with ID (%v) already exists`, tc.launchInput.ID,
				), logs.All()[1].Message)
				a.Nil(err)
				return
			}

			a.Equal(tc.launchErr, err)
		})
	}
}

func TestNewLaunchSvc(t *testing.T) {
	a := assert.New(t)
	logger := zaptest.NewLogger(t)
	svc := NewLaunchSvc(
		"ec2",
		"eks",
		nil,
		nil,
		nil,
		nil,
		logger,
	).(*launchService)
	expectedTargets := [][]string{{EC2Label, UbuntuLabel}, {EKSLabel, UbuntuLabel}}

	targets := make([][]string, 0)
	for i := range svc.launchers {
		targets = append(targets, svc.launchers[i].Target)
	}

	a.Equal(expectedTargets, targets)
	a.Equal(logger, svc.logger)
}
