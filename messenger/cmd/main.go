package main

import (
	"context"
	"log"
	"os"

	"github.com/CameronXie/aws-github-actions-runner/messenger/internal/handler"
	"github.com/CameronXie/aws-github-actions-runner/messenger/internal/messenger"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

const (
	regionEnv         = "DEFAULT_REGION"
	publisherTopicEnv = "PUBLISHER_TOPIC"
)

func main() {
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithDefaultRegion(os.Getenv(regionEnv)),
	)
	handleError(err)

	lambda.Start(handler.SetupHandler(messenger.New(
		sns.NewFromConfig(cfg),
		os.Getenv(publisherTopicEnv),
	)))
}

func handleError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
