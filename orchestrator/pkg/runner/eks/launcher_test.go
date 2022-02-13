package eks

import (
	"context"
	"strings"
	"testing"

	"github.com/CameronXie/aws-github-actions-runner/orchestrator/pkg/runner"
	"github.com/stretchr/testify/assert"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestEksLauncher_Launch(t *testing.T) {
	prefix := "prefix"
	cases := map[string]struct {
		config        *LaunchConfig
		input         *runner.LaunchInput
		deploymentErr error
		err           error
	}{
		"create a deployment": {
			config: &LaunchConfig{
				Namespace: "ns",
				Runner: ContainerResource{
					Image:  "runner",
					CPU:    "1",
					Memory: "1Gi",
				},
				DinD: ContainerResource{
					Image:  "dind",
					CPU:    "1",
					Memory: "1Gi",
				},
				GitHubSecret:    "secret",
				GitHubSecretKey: "secretKey",
			},
			input: &runner.LaunchInput{
				ID:         1,
				Owner:      "owner",
				Repository: "repo",
				Labels:     []string{"eks", "ubuntu"},
			},
		},
		"runner with given name already exists": {
			config: &LaunchConfig{
				Namespace: "ns",
				Runner: ContainerResource{
					Image:  "runner",
					CPU:    "1",
					Memory: "1Gi",
				},
				DinD: ContainerResource{
					Image:  "dind",
					CPU:    "1",
					Memory: "1Gi",
				},
				GitHubSecret:    "secret",
				GitHubSecretKey: "secretKey",
			},
			input: &runner.LaunchInput{
				ID:         1,
				Owner:      "owner",
				Repository: "repo",
				Labels:     []string{"eks", "ubuntu"},
			},
			deploymentErr: &k8serr.StatusError{
				ErrStatus: metav1.Status{Reason: metav1.StatusReasonAlreadyExists},
			},
			err: &runner.AlreadyExistsError{
				Type: RunnerType,
				ID:   1,
			},
		},
	}

	for n, tc := range cases {
		t.Run(n, func(t *testing.T) {
			a := assert.New(t)
			client := &mockedKubeClient{
				deploymentErr: tc.deploymentErr,
			}
			l := NewLauncher(prefix, client, tc.config).(*eksLauncher)
			a.Equal(tc.err, l.Launch(context.TODO(), tc.input))

			a.Equal(l.getRunnerDeployment(&RunnerConfig{
				ID:         tc.input.ID,
				Owner:      tc.input.Owner,
				Repository: tc.input.Repository,
				Labels:     strings.Join(tc.input.Labels, ","),
			}), client.deployment)
		})
	}
}
