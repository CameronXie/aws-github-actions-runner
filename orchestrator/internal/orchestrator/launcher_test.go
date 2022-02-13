package orchestrator

import (
	"context"
	"testing"

	"github.com/CameronXie/aws-github-actions-runner/orchestrator/pkg/runner"
	"github.com/stretchr/testify/assert"
)

func TestLauncher_IsMatched(t *testing.T) {
	cases := map[string]struct {
		labels   []string
		expected bool
	}{
		"empty label": {
			labels:   make([]string, 0),
			expected: false,
		},
		"ubuntu ec2 self-hosted label": {
			labels:   []string{SelfHostedLabel, UbuntuLabel, EC2Label},
			expected: true,
		},
		"ubuntu ec2 label": {
			labels:   []string{UbuntuLabel, EC2Label},
			expected: false,
		},
		"ubuntu eks label": {
			labels:   []string{SelfHostedLabel, UbuntuLabel, EKSLabel},
			expected: false,
		},
		"ubuntu label": {
			labels:   []string{SelfHostedLabel, UbuntuLabel},
			expected: false,
		},
	}

	for n, tc := range cases {
		t.Run(n, func(t *testing.T) {
			a := assert.New(t)
			a.Equal(
				tc.expected,
				NewLauncher(
					[]string{EC2Label, UbuntuLabel},
					new(mockedRunnerLauncher),
				).IsMatched(tc.labels),
			)
		})
	}
}

type mockedRunnerLauncher struct {
	input     *runner.LaunchInput
	launchErr error
}

func (m *mockedRunnerLauncher) Launch(_ context.Context, input *runner.LaunchInput) error {
	m.input = input
	return m.launchErr
}
