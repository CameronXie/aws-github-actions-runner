package main

import (
	"context"
	"fmt"
	"os"

	"github.com/CameronXie/aws-github-actions-runner/orchestrator/internal/handler"
	eksrunner "github.com/CameronXie/aws-github-actions-runner/orchestrator/pkg/runner/eks"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"go.uber.org/zap"
)

const (
	runnerNamePrefix         = "eks-runner"
	eksClusterEnv            = "EKS_CLUSTER"
	eksNamespaceEnv          = "EKS_NAMESPACE"
	githubTokenSecretEnv     = "GITHUB_TOKEN_SECRET"
	githubTokenSecretKeyEnv  = "GITHUB_TOKEN_SECRET_TOKEN"
	runnerContainerImageEnv  = "RUNNER_CONTAINER_IMAGE"
	runnerContainerCPUEnv    = "RUNNER_CONTAINER_CPU"
	runnerContainerMemoryEnv = "RUNNER_CONTAINER_MEMORY"
	dindContainerImageEnv    = "DIND_CONTAINER_IMAGE"
	dindContainerCPUEnv      = "DIND_CONTAINER_CPU"
	dindContainerMemoryEnv   = "DIND_CONTAINER_MEMORY"
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

	lambda.Start(handler.SetupLauncherHandler(
		eksrunner.NewLauncher(runnerNamePrefix, kubeClient, &eksrunner.LaunchConfig{
			Namespace: os.Getenv(eksNamespaceEnv),
			Runner: eksrunner.ContainerResource{
				Image:  os.Getenv(runnerContainerImageEnv),
				CPU:    os.Getenv(runnerContainerCPUEnv),
				Memory: os.Getenv(runnerContainerMemoryEnv),
			},
			DinD: eksrunner.ContainerResource{
				Image:  os.Getenv(dindContainerImageEnv),
				CPU:    os.Getenv(dindContainerCPUEnv),
				Memory: os.Getenv(dindContainerMemoryEnv),
			},
			GitHubSecret:    os.Getenv(githubTokenSecretEnv),
			GitHubSecretKey: os.Getenv(githubTokenSecretKeyEnv),
		}),
		logger,
	))
}
