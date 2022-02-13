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

type TerminationInput struct {
	ID     int
	Labels []string
}

type TerminationService interface {
	Terminate(ctx context.Context, input *TerminationInput) error
}

type terminationService struct {
	terminators []*Terminator
	logger      *zap.Logger
}

func (svc *terminationService) Terminate(ctx context.Context, input *TerminationInput) error {
	var terminator runner.Terminator
	for _, t := range svc.terminators {
		if t.IsMatched(input.Labels) {
			svc.logger.Debug(fmt.Sprintf(
				`id: %v, labels: %v, target: %v\n`,
				input.ID,
				strings.Join(input.Labels, ","),
				t.Target,
			))
			terminator = t.Terminator
		}
	}

	if terminator == nil {
		return fmt.Errorf(`labels (%v) termination currently is not supported`, input.Labels)
	}

	err := terminator.Terminate(ctx, input.ID)
	if err != nil && runner.IsNotExistsError(err) {
		svc.logger.Info(fmt.Sprintf(`runner with ID (%v) not exists`, input.ID))
		return nil
	}

	return err
}

func NewTerminationSvc(
	ec2Client *ec2.Client,
	kubeClient kubernetes.Interface,
	eksTerminationConfig *eksrunner.RunnerTerminationConfig,
	logger *zap.Logger,
) TerminationService {
	return &terminationService{
		terminators: []*Terminator{
			NewTerminator(EC2Label, ec2runner.NewTerminator(ec2Client)),
			NewTerminator(EKSLabel, eksrunner.NewTerminator(kubeClient, eksTerminationConfig)),
		},
		logger: logger,
	}
}
