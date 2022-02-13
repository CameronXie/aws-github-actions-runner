package orchestrator

import (
	"sort"

	"github.com/CameronXie/aws-github-actions-runner/orchestrator/pkg/runner"
)

const (
	SelfHostedLabel = "self-hosted"
)

type Launcher struct {
	Target   []string
	Launcher runner.Launcher
}

func (l *Launcher) IsMatched(labels []string) bool {
	if len(labels) != len(l.Target)+1 {
		return false
	}

	for _, w := range labels {
		if w == SelfHostedLabel {
			continue
		}

		index := sort.SearchStrings(l.Target, w)
		if index < len(l.Target) && l.Target[index] == w {
			continue
		}

		return false
	}

	return true
}

func NewLauncher(targets []string, launcher runner.Launcher) *Launcher {
	sort.Strings(targets)

	return &Launcher{
		Target:   targets,
		Launcher: launcher,
	}
}
