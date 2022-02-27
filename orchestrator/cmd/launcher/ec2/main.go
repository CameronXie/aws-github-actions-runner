package main

import (
	"context"
	"fmt"
	"os"
	"text/template"

	"github.com/CameronXie/aws-github-actions-runner/orchestrator/internal/handler"
	ec2runner "github.com/CameronXie/aws-github-actions-runner/orchestrator/pkg/runner/ec2"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"go.uber.org/zap"
)

const (
	runnerNamePrefix            = "ec2-runner"
	githubTokenEnv              = "GITHUB_TOKEN"
	runnerVersionEnv            = "GITHUB_RUNNER_VERSION"
	subnetEnv                   = "SUBNET_ID"
	launchTemplateEnv           = "LAUNCH_TEMPLATE_ID"
	ubuntuLaunchTemplateVersion = "$Latest"
	userData                    = "userdata.tmpl"
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

	lambda.Start(handler.SetupLauncherHandler(
		ec2runner.NewLauncher(runnerNamePrefix, ec2.NewFromConfig(cfg), &ec2runner.LaunchConfig{
			TemplateID:       os.Getenv(launchTemplateEnv),
			TemplateVersion:  ubuntuLaunchTemplateVersion,
			SubnetID:         os.Getenv(subnetEnv),
			GitHubToken:      os.Getenv(githubTokenEnv),
			RunnerVersion:    os.Getenv(runnerVersionEnv),
			UserDataTemplate: template.Must(template.ParseFiles(userData)),
		}),
		logger,
	))
}
