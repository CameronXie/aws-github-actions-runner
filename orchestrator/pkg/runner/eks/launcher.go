package eks

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/CameronXie/aws-github-actions-runner/orchestrator/pkg/runner"
	"github.com/aws/aws-sdk-go-v2/aws"
	appv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	runnerReplicas                = 1
	terminationGracePeriodSeconds = 10
)

type RunnerConfig struct {
	ID         int
	Owner      string
	Repository string
	Labels     string
}

type ContainerResource struct {
	Image  string
	CPU    string
	Memory string
}

type LaunchConfig struct {
	Namespace       string
	Runner          ContainerResource
	DinD            ContainerResource
	GitHubSecret    string
	GitHubSecretKey string
}

type eksLauncher struct {
	runnerNamePrefix string
	kubeClient       kubernetes.Interface
	config           *LaunchConfig
}

func (l *eksLauncher) Launch(ctx context.Context, input *runner.LaunchInput) error {
	_, err := l.kubeClient.AppsV1().Deployments(l.config.Namespace).
		Create(ctx, l.getRunnerDeployment(&RunnerConfig{
			ID:         input.ID,
			Owner:      input.Owner,
			Repository: input.Repository,
			Labels:     strings.Join(input.Labels, ","),
		}), metav1.CreateOptions{})

	if errors.IsAlreadyExists(err) {
		return &runner.AlreadyExistsError{
			Type: RunnerType,
			ID:   input.ID,
		}
	}

	return err
}

func (l *eksLauncher) getRunnerDeployment(config *RunnerConfig) *appv1.Deployment {
	return &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: strconv.Itoa(config.ID),
			Labels: map[string]string{
				"app": "actions-runner",
			},
		},
		Spec: appv1.DeploymentSpec{
			Replicas: aws.Int32(runnerReplicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "actions-runner",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "actions-runner",
					},
				},
				Spec: apiv1.PodSpec{
					TerminationGracePeriodSeconds: aws.Int64(terminationGracePeriodSeconds),
					Containers: []apiv1.Container{
						{
							Name:  "actions-runner",
							Image: l.config.Runner.Image,
							SecurityContext: &apiv1.SecurityContext{
								RunAsNonRoot: aws.Bool(true),
							},
							Resources: apiv1.ResourceRequirements{
								Requests: apiv1.ResourceList{
									apiv1.ResourceCPU:    resource.MustParse(l.config.Runner.CPU),
									apiv1.ResourceMemory: resource.MustParse(l.config.Runner.Memory),
								},
							},
							Env: []apiv1.EnvVar{
								{Name: "RUNNER_NAME", Value: fmt.Sprintf("%v-%v", l.runnerNamePrefix, config.ID)},
								{Name: "RUNNER_LABELS", Value: config.Labels},
								{Name: "RUNNER_ORG", Value: config.Owner},
								{Name: "RUNNER_REPO", Value: config.Repository},
								{Name: "GH_TOKEN", ValueFrom: &apiv1.EnvVarSource{
									SecretKeyRef: &apiv1.SecretKeySelector{
										LocalObjectReference: apiv1.LocalObjectReference{
											Name: l.config.GitHubSecret,
										},
										Key: l.config.GitHubSecretKey,
									},
								}},
								{Name: "DOCKER_HOST", Value: "tcp://localhost:2375"},
							},
						},
						{
							Name:  "dind",
							Image: l.config.DinD.Image,
							SecurityContext: &apiv1.SecurityContext{
								Privileged: aws.Bool(true),
							},
							Resources: apiv1.ResourceRequirements{
								Requests: apiv1.ResourceList{
									apiv1.ResourceCPU:    resource.MustParse(l.config.DinD.CPU),
									apiv1.ResourceMemory: resource.MustParse(l.config.DinD.Memory),
								},
							},
							Env: []apiv1.EnvVar{
								{Name: "DOCKER_TLS_CERTDIR", Value: ""},
							},
						},
					},
				},
			},
		},
		Status: appv1.DeploymentStatus{},
	}
}

func NewLauncher(
	runnerNamePrefix string,
	kubeClient kubernetes.Interface,
	config *LaunchConfig,
) runner.Launcher {
	return &eksLauncher{
		runnerNamePrefix: runnerNamePrefix,
		kubeClient:       kubeClient,
		config:           config,
	}
}
