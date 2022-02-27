package main

import (
	"context"
	"fmt"
	"os"

	"github.com/CameronXie/aws-github-actions-runner/orchestrator/internal/handler"
	ec2runner "github.com/CameronXie/aws-github-actions-runner/orchestrator/pkg/runner/ec2"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"go.uber.org/zap"
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

	lambda.Start(handler.SetupTerminatorHandler(
		ec2runner.NewTerminator(ec2.NewFromConfig(cfg)),
		logger,
	))
}
