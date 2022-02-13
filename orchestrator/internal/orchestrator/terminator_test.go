package orchestrator

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTerminator_IsMatched(t *testing.T) {
	cases := map[string]struct {
		labels   []string
		expected bool
	}{
		"empty label": {
			labels:   make([]string, 0),
			expected: false,
		},
		"ubuntu ec2 label": {
			labels:   []string{SelfHostedLabel, UbuntuLabel, EC2Label},
			expected: true,
		},
		"ec2 label": {
			labels:   []string{EC2Label},
			expected: true,
		},
		"ubuntu eks label": {
			labels:   []string{SelfHostedLabel, UbuntuLabel, EKSLabel},
			expected: false,
		},
	}

	for n, tc := range cases {
		t.Run(n, func(t *testing.T) {
			a := assert.New(t)
			a.Equal(
				tc.expected,
				NewTerminator(
					EC2Label,
					new(mockedRunnerTerminator),
				).IsMatched(tc.labels),
			)
		})
	}
}

type mockedRunnerTerminator struct {
	id           int
	terminateErr error
}

func (m *mockedRunnerTerminator) Terminate(_ context.Context, id int) error {
	m.id = id
	return m.terminateErr
}
