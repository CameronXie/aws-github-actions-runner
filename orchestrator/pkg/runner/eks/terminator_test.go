package eks

import (
	"context"
	"testing"

	"github.com/CameronXie/aws-github-actions-runner/orchestrator/pkg/runner"
	"github.com/stretchr/testify/assert"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestEksTerminator_Terminate(t *testing.T) {
	config := &RunnerTerminationConfig{
		Cluster:   "cluster",
		Namespace: "ns",
	}
	cases := map[string]struct {
		id                 int
		deploymentErr      error
		expectedDeleteName string
		err                error
	}{
		"terminate deployment": {
			id:                 1,
			expectedDeleteName: "1",
		},
		"deployment not found": {
			id: 1,
			deploymentErr: &k8serr.StatusError{
				ErrStatus: metav1.Status{Reason: metav1.StatusReasonNotFound},
			},
			expectedDeleteName: "1",
			err: &runner.NotExistsError{
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
			deletePolicy := metav1.DeletePropagationForeground

			err := NewTerminator(client, config).Terminate(context.TODO(), tc.id)

			a.Equal(tc.err, err)
			a.Equal(tc.expectedDeleteName, client.deleteDeploymentName)
			a.EqualValues(metav1.DeleteOptions{
				PropagationPolicy: &deletePolicy,
			}, client.deleteOpt)
		})
	}
}
