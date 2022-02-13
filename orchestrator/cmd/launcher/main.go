package main

import (
	"context"
	"fmt"
	"os"
	"path"
	"text/template"

	"github.com/CameronXie/aws-github-actions-runner/orchestrator/internal/handler"
	"github.com/CameronXie/aws-github-actions-runner/orchestrator/internal/orchestrator"
	ec2runner "github.com/CameronXie/aws-github-actions-runner/orchestrator/pkg/runner/ec2"
	eksrunner "github.com/CameronXie/aws-github-actions-runner/orchestrator/pkg/runner/eks"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"go.uber.org/zap"
)

const (
	userDataDir                    = "userdata"
	githubTokenEnv                 = "GITHUB_TOKEN"
	runnerVersionEnv               = "GITHUB_RUNNER_VERSION"
	ec2RunnerNamePrefix            = "ec2-runner"
	eksRunnerNamePrefix            = "eks-runner"
	subnetEnv                      = "SUBNET_ID"
	ubuntuLaunchTemplateEnv        = "UBUNTU_LAUNCH_TEMPLATE_ID"
	eksClusterEnv                  = "EKS_CLUSTER"
	eksNamespaceEnv                = "EKS_NAMESPACE"
	githubTokenSecretEnv           = "GITHUB_TOKEN_SECRET"
	githubTokenSecretKeyEnv        = "GITHUB_TOKEN_SECRET_TOKEN"
	ubuntuRunnerContainerImageEnv  = "UBUNTU_RUNNER_CONTAINER_IMAGE"
	ubuntuRunnerContainerCPUEnv    = "UBUNTU_RUNNER_CONTAINER_CPU"
	ubuntuRunnerContainerMemoryEnv = "UBUNTU_RUNNER_CONTAINER_MEMORY"
	dindContainerImageEnv          = "DIND_CONTAINER_IMAGE"
	dindContainerCPUEnv            = "DIND_CONTAINER_CPU"
	dindContainerMemoryEnv         = "DIND_CONTAINER_MEMORY"
	ubuntuLaunchTemplateVersion    = "$Latest"
	ubuntuUserData                 = "ubuntu.tmpl"
)

func main() {
	logger, _ := zap.NewProduction()
	defer func() { _ = logger.Sync() }()

	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithDefaultRegion(os.Getenv("DEFAULT_REGION")),
	)

	if err != nil {
		logger.Fatal(fmt.Sprintf("aws sdk error: %v", err.Error()))
	}

	kubeClient, kubeErr := eksrunner.GetKubeClient(
		context.TODO(),
		os.Getenv(eksClusterEnv),
		eks.NewFromConfig(cfg), nil, nil)
	if kubeErr != nil {
		logger.Fatal(fmt.Sprintf("kube client error: %v", err.Error()))
	}

	lambda.Start(handler.SetupLauncherHandler(orchestrator.NewLaunchSvc(
		ec2RunnerNamePrefix,
		eksRunnerNamePrefix,
		ec2.NewFromConfig(cfg),
		kubeClient,
		&ec2runner.LaunchConfig{
			TemplateID:       os.Getenv(ubuntuLaunchTemplateEnv),
			TemplateVersion:  ubuntuLaunchTemplateVersion,
			SubnetID:         os.Getenv(subnetEnv),
			GitHubToken:      os.Getenv(githubTokenEnv),
			RunnerVersion:    os.Getenv(runnerVersionEnv),
			UserDataTemplate: template.Must(template.ParseFiles(path.Join(userDataDir, ubuntuUserData))),
		},
		&eksrunner.LaunchConfig{
			Namespace: os.Getenv(eksNamespaceEnv),
			Runner: eksrunner.ContainerResource{
				Image:  os.Getenv(ubuntuRunnerContainerImageEnv),
				CPU:    os.Getenv(ubuntuRunnerContainerCPUEnv),
				Memory: os.Getenv(ubuntuRunnerContainerMemoryEnv),
			},
			DinD: eksrunner.ContainerResource{
				Image:  os.Getenv(dindContainerImageEnv),
				CPU:    os.Getenv(dindContainerCPUEnv),
				Memory: os.Getenv(dindContainerMemoryEnv),
			},
			GitHubSecret:    os.Getenv(githubTokenSecretEnv),
			GitHubSecretKey: os.Getenv(githubTokenSecretKeyEnv),
		},
		logger,
	)))
}
