package handler

import (
	"context"

	"github.com/CameronXie/aws-github-actions-runner/publisher/internal/publisher"
)

type DynamoDBEventHandler func(ctx context.Context) error

func SetupPublisherHandler(p publisher.Publisher) DynamoDBEventHandler {
	return func(ctx context.Context) error {
		return p.Publish(ctx)
	}
}
