package main

import (
	"context"
	"log"
	"os"
	"strconv"

	"github.com/CameronXie/aws-github-actions-runner/publisher/internal/handler"
	"github.com/CameronXie/aws-github-actions-runner/publisher/internal/publisher"
	"github.com/CameronXie/aws-github-actions-runner/publisher/internal/queue"
	"github.com/CameronXie/aws-github-actions-runner/publisher/internal/storage"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"go.uber.org/zap"
)

const (
	ec2Host                = "ec2"
	eksHost                = "eks"
	tableNameEnv           = "JOBS_TABLE"
	tableHostIndexEnv      = "JOBS_TABLE_HOST_INDEX"
	ec2CurrencyLimitEnv    = "EC2_CURRENCY_LIMIT"
	eksCurrencyLimitEnv    = "EKS_CURRENCY_LIMIT"
	launchQueueURLEnv      = "LAUNCH_QUEUE_URL"
	terminationQueueURLEnv = "TERMINATION_QUEUE_URL"
)

func main() {
	logger, _ := zap.NewProduction()
	defer func() { _ = logger.Sync() }()

	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithDefaultRegion(os.Getenv("DEFAULT_REGION")),
	)
	handleError(err)

	ec2Limits, ec2Err := strconv.ParseInt(os.Getenv(ec2CurrencyLimitEnv), 10, 32)
	handleError(ec2Err)

	eksLimits, eksErr := strconv.ParseInt(os.Getenv(eksCurrencyLimitEnv), 10, 32)
	handleError(eksErr)

	lambda.Start(handler.SetupPublisherHandler(publisher.New(
		os.Getenv(launchQueueURLEnv),
		os.Getenv(terminationQueueURLEnv),
		[]publisher.HostOption{
			{
				Host:  ec2Host,
				Limit: int32(ec2Limits),
			},
			{
				Host:  eksHost,
				Limit: int32(eksLimits),
			},
		},
		storage.New(
			dynamodb.NewFromConfig(cfg),
			os.Getenv(tableNameEnv),
			os.Getenv(tableHostIndexEnv),
		),
		queue.New(sqs.NewFromConfig(cfg)),
		logger,
	)))
}

func handleError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
