package handler

import (
	"encoding/json"
	"testing"

	"github.com/CameronXie/aws-github-actions-runner/orchestrator/pkg/runner"
	"github.com/stretchr/testify/assert"
)

func TestLaunchEvent_UnmarshalJSON(t *testing.T) {
	cases := map[string]struct {
		input    []byte
		expected *LaunchEvent
		errType  interface{}
	}{
		"unmarshal launch event": {
			input: []byte(`{"Message":"{\"ID\":1,\"Owner\":\"owner\",\"Repository\":\"repo\",\"Labels\":[\"ubuntu\"]}"}`),
			expected: &LaunchEvent{Message: &runner.LaunchInput{
				ID:         1,
				Owner:      "owner",
				Repository: "repo",
				Labels:     []string{"ubuntu"},
			}},
		},
		"invalid json input": {
			input:   []byte(`{`),
			errType: new(json.SyntaxError),
		},
		"invalid message input": {
			input:   []byte(`{"Message":"{"}`),
			errType: new(json.SyntaxError),
		},
	}

	for n, tc := range cases {
		t.Run(n, func(t *testing.T) {
			a := assert.New(t)
			res := new(LaunchEvent)
			err := json.Unmarshal(tc.input, res)

			if tc.errType == nil {
				a.Nil(err)
				a.Equal(tc.expected, res)
				return
			}

			a.ErrorAs(err, &tc.errType)
		})
	}
}

func TestTerminationEvent_UnmarshalJSON(t *testing.T) {
	cases := map[string]struct {
		input    []byte
		expected *TerminationEvent
		errType  interface{}
	}{
		"unmarshal launch event": {
			input:    []byte(`{"Message":"{\"ID\":1,\"Owner\":\"owner\",\"Repository\":\"repo\",\"Labels\":[\"ubuntu\"]}"}`),
			expected: &TerminationEvent{Message: &TerminationInput{ID: 1}},
		},
		"invalid json input": {
			input:   []byte(`{`),
			errType: new(json.SyntaxError),
		},
		"invalid message input": {
			input:   []byte(`{"Message":"{"}`),
			errType: new(json.SyntaxError),
		},
	}

	for n, tc := range cases {
		t.Run(n, func(t *testing.T) {
			a := assert.New(t)
			res := new(TerminationEvent)
			err := json.Unmarshal(tc.input, res)

			if tc.errType == nil {
				a.Nil(err)
				a.Equal(tc.expected, res)
				return
			}

			a.ErrorAs(err, &tc.errType)
		})
	}
}
