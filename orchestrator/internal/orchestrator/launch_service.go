package orchestrator

import (
	"context"
	"fmt"
	"strings"

	"github.com/CameronXie/aws-github-actions-runner/orchestrator/pkg/runner"
	ec2runner "github.com/CameronXie/aws-github-actions-runner/orchestrator/pkg/runner/ec2"
	eksrunner "github.com/CameronXie/aws-github-actions-runner/orchestrator/pkg/runner/eks"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
)

type LaunchService interface {
	Launch(ctx context.Context, input *runner.LaunchInput) error
}

type launchService struct {
	launchers []*Launcher
	logger    *zap.Logger
}

func (svc *launchService) Launch(ctx context.Context, input *runner.LaunchInput) error {
	var launcher runner.Launcher
	for _, l := range svc.launchers {
		if l.IsMatched(input.Labels) {
			svc.logger.Debug(fmt.Sprintf(
				`id: %v, labels: %v, target: %v\n`,
				input.ID,
				strings.Join(input.Labels, ","),
				strings.Join(l.Target, ","),
			))
			launcher = l.Launcher
		}
	}

	if launcher == nil {
		return fmt.Errorf(`labels (%v) launch currently is not supported`, input.Labels)
	}

	err := launcher.Launch(ctx, input)
	if err != nil && runner.IsAlreadyExistsError(err) {
		svc.logger.Info(fmt.Sprintf(`runner with ID (%v) already exists`, input.ID))
		return nil
	}

	return err
}

func NewLaunchSvc(
	ec2RunnerNamePrefix string,
	eksRunnerNamePrefix string,
	ec2Client *ec2.Client,
	kubeClient kubernetes.Interface,
	ubuntuEC2Config *ec2runner.LaunchConfig,
	ubuntuEKSConfig *eksrunner.LaunchConfig,
	logger *zap.Logger,
) LaunchService {
	return &launchService{
		launchers: []*Launcher{
			NewLauncher([]string{EC2Label, UbuntuLabel}, ec2runner.NewLauncher(ec2RunnerNamePrefix, ec2Client, ubuntuEC2Config)),
			NewLauncher([]string{EKSLabel, UbuntuLabel}, eksrunner.NewLauncher(eksRunnerNamePrefix, kubeClient, ubuntuEKSConfig)),
		},
		logger: logger,
	}
}
