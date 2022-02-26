package handler

import (
	"encoding/json"

	"github.com/CameronXie/aws-github-actions-runner/orchestrator/pkg/runner"
)

type LaunchEvent struct {
	Message *runner.LaunchInput
}

func (l *LaunchEvent) UnmarshalJSON(data []byte) error {
	msg := new(runner.LaunchInput)
	if err := unmarshalEvent(data, msg); err != nil {
		return err
	}

	l.Message = msg
	return nil
}

type TerminationInput struct {
	ID uint64
}

type TerminationEvent struct {
	Message *TerminationInput
}

func (l *TerminationEvent) UnmarshalJSON(data []byte) error {
	msg := new(TerminationInput)
	if err := unmarshalEvent(data, msg); err != nil {
		return err
	}

	l.Message = msg
	return nil
}

func unmarshalEvent(data []byte, o interface{}) error {
	raw := new(struct {
		Message string
	})

	if err := json.Unmarshal(data, raw); err != nil {
		return err
	}

	return json.Unmarshal([]byte(raw.Message), o)
}
