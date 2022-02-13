package eks

import (
	"context"
	"strconv"

	"github.com/CameronXie/aws-github-actions-runner/orchestrator/pkg/runner"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type RunnerTerminationConfig struct {
	Cluster   string
	Namespace string
}

type eksTerminator struct {
	kubeClient kubernetes.Interface
	config     *RunnerTerminationConfig
}

func (t *eksTerminator) Terminate(ctx context.Context, id int) error {
	deletePolicy := metav1.DeletePropagationForeground
	err := t.kubeClient.AppsV1().
		Deployments(t.config.Namespace).
		Delete(
			ctx,
			strconv.Itoa(id),
			metav1.DeleteOptions{
				PropagationPolicy: &deletePolicy,
			},
		)

	if errors.IsNotFound(err) {
		return &runner.NotExistsError{
			Type: RunnerType,
			ID:   id,
		}
	}

	return err
}

func NewTerminator(
	kubeClient kubernetes.Interface,
	config *RunnerTerminationConfig,
) runner.Terminator {
	return &eksTerminator{
		kubeClient: kubeClient,
		config:     config,
	}
}
