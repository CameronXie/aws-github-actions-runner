package handler

import (
	"context"

	"github.com/CameronXie/aws-github-actions-runner/orchestrator/internal/orchestrator"
	"github.com/CameronXie/aws-github-actions-runner/orchestrator/pkg/runner"
	"github.com/aws/aws-lambda-go/events"
)

func SetupLauncherHandler(svc orchestrator.LaunchService) SQSEventHandler {
	return func(ctx context.Context, event events.SQSEvent) error {
		input := new(runner.LaunchInput)
		if inputErr := getInput(event, input); inputErr != nil {
			return inputErr
		}

		return svc.Launch(ctx, input)
	}
}
