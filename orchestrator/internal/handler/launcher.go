package handler

import (
	"context"
	"fmt"

	"github.com/CameronXie/aws-github-actions-runner/orchestrator/pkg/runner"
	"github.com/aws/aws-lambda-go/events"
	"go.uber.org/zap"
)

func SetupLauncherHandler(launcher runner.Launcher, logger *zap.Logger) SQSEventHandler {
	return func(ctx context.Context, event events.SQSEvent) error {
		input := new(LaunchEvent)
		if inputErr := getInput(event, input); inputErr != nil {
			return inputErr
		}

		logger.Info(fmt.Sprintf("launching runner with ID (%v)", input.Message.ID))

		err := launcher.Launch(ctx, input.Message)
		if err != nil && runner.IsAlreadyExistsError(err) {
			logger.Info(fmt.Sprintf(`runner with ID (%v) already exists`, input.Message.ID))
			return nil
		}

		return err
	}
}
