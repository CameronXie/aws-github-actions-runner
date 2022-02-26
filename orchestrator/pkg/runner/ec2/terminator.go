package ec2

import (
	"context"

	"github.com/CameronXie/aws-github-actions-runner/orchestrator/pkg/runner"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

type TerminateInstancesAPIClient interface {
	ec2.DescribeInstancesAPIClient
	TerminateInstances(
		ctx context.Context,
		params *ec2.TerminateInstancesInput,
		optFns ...func(*ec2.Options),
	) (*ec2.TerminateInstancesOutput, error)
}

type ec2Terminator struct {
	client TerminateInstancesAPIClient
}

func (t *ec2Terminator) Terminate(ctx context.Context, id uint64) error {
	ids, rErr := getInstanceIDByTag(t.client, ctx, idTag, []string{uint64ToString(id)})

	if rErr != nil {
		return rErr
	}

	if len(ids) == 0 {
		return &runner.NotExistsError{
			ID:   id,
			Type: RunnerType,
		}
	}

	_, dErr := t.client.TerminateInstances(ctx, &ec2.TerminateInstancesInput{
		InstanceIds: ids,
	})

	return dErr
}

func NewTerminator(client TerminateInstancesAPIClient) runner.Terminator {
	return &ec2Terminator{client: client}
}
