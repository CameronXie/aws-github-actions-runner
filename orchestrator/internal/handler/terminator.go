package handler

import (
	"context"
	"fmt"

	"github.com/CameronXie/aws-github-actions-runner/orchestrator/pkg/runner"
	"github.com/aws/aws-lambda-go/events"
	"go.uber.org/zap"
)

func SetupTerminatorHandler(terminator runner.Terminator, logger *zap.Logger) SQSEventHandler {
	return func(ctx context.Context, event events.SQSEvent) error {
		input := new(TerminationEvent)
		if inputErr := getInput(event, input); inputErr != nil {
			return inputErr
		}

		logger.Info(fmt.Sprintf("terminating runner with ID (%v)", input.Message.ID))

		err := terminator.Terminate(ctx, input.Message.ID)
		if err != nil && runner.IsNotExistsError(err) {
			logger.Info(fmt.Sprintf(`runner with ID (%v) not exists`, input.Message.ID))
			return nil
		}

		return err
	}
}
