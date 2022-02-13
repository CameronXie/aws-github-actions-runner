package main

import (
	"context"
	"fmt"
	"os"

	"github.com/CameronXie/aws-github-actions-runner/orchestrator/internal/handler"
	"github.com/CameronXie/aws-github-actions-runner/orchestrator/internal/orchestrator"
	eksrunner "github.com/CameronXie/aws-github-actions-runner/orchestrator/pkg/runner/eks"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"go.uber.org/zap"
)

const (
	eksClusterEnv   = "EKS_CLUSTER"
	eksNamespaceEnv = "EKS_NAMESPACE"
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

	lambda.Start(handler.SetupTerminatorHandler(orchestrator.NewTerminationSvc(
		ec2.NewFromConfig(cfg),
		kubeClient,
		&eksrunner.RunnerTerminationConfig{
			Cluster:   os.Getenv(eksClusterEnv),
			Namespace: os.Getenv(eksNamespaceEnv),
		},
		logger,
	)))
}
