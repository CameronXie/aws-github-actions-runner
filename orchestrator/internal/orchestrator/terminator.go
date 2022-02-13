package orchestrator

import "github.com/CameronXie/aws-github-actions-runner/orchestrator/pkg/runner"

type Terminator struct {
	Target     string
	Terminator runner.Terminator
}

func (t *Terminator) IsMatched(labels []string) bool {
	if len(labels) < 1 {
		return false
	}

	for _, w := range labels {
		if t.Target == w {
			return true
		}
	}

	return false
}

func NewTerminator(target string, terminator runner.Terminator) *Terminator {
	return &Terminator{
		Target:     target,
		Terminator: terminator,
	}
}
