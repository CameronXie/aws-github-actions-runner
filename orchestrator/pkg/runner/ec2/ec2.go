package ec2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

const (
	idTag      = "GITHUB_WORKFLOW_JOB_ID"
	RunnerType = "ec2"
)

func getInstanceIDByTag(
	client ec2.DescribeInstancesAPIClient,
	ctx context.Context,
	tag string,
	values []string,
) ([]string, error) {
	resp, err := client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String(fmt.Sprintf("tag:%s", tag)),
				Values: values,
			},
		},
	})

	if err != nil {
		return nil, err
	}

	ids := make([]string, 0)
	for _, r := range resp.Reservations {
		for i := range r.Instances {
			ids = append(ids, aws.ToString(r.Instances[i].InstanceId))
		}
	}

	return ids, nil
}
