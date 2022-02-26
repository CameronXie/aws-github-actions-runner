package handler

import (
	"context"

	"github.com/CameronXie/aws-github-actions-runner/messenger/internal/messenger"
	"github.com/aws/aws-lambda-go/events"
)

type DynamoDBEventHandler = func(ctx context.Context, e events.DynamoDBEvent) error

func SetupHandler(svc messenger.Service) DynamoDBEventHandler {
	return func(ctx context.Context, _ events.DynamoDBEvent) error {
		return svc.NotifyPublisher(ctx)
	}
}
