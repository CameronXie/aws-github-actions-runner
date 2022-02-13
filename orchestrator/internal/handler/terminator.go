package handler

import (
	"context"

	"github.com/CameronXie/aws-github-actions-runner/orchestrator/internal/orchestrator"
	"github.com/aws/aws-lambda-go/events"
)

func SetupTerminatorHandler(svc orchestrator.TerminationService) SQSEventHandler {
	return func(ctx context.Context, event events.SQSEvent) error {
		input := new(orchestrator.TerminationInput)
		if inputErr := getInput(event, input); inputErr != nil {
			return inputErr
		}

		return svc.Terminate(ctx, input)
	}
}
