package ec2

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/CameronXie/aws-github-actions-runner/orchestrator/pkg/runner"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/stretchr/testify/assert"
)

func TestEc2Terminator_Terminate(t *testing.T) {
	cases := map[string]struct {
		id                              int
		numInstances                    int
		describeInstancesErr            error
		expectedDescribeInstanceInput   *ec2.DescribeInstancesInput
		expectedTerminateInstancesInput *ec2.TerminateInstancesInput
		err                             error
	}{
		"terminates instances with given tag": {
			id:           1,
			numInstances: 2,
			expectedDescribeInstanceInput: &ec2.DescribeInstancesInput{
				Filters: []types.Filter{
					{
						Name:   aws.String(fmt.Sprintf("tag:%s", idTag)),
						Values: []string{"1"},
					},
				},
			},
			expectedTerminateInstancesInput: &ec2.TerminateInstancesInput{
				InstanceIds: []string{"0", "1"},
			},
		},
		"describe instance error": {
			id:                   1,
			describeInstancesErr: errors.New("describe instance error"),
			expectedDescribeInstanceInput: &ec2.DescribeInstancesInput{
				Filters: []types.Filter{
					{
						Name:   aws.String(fmt.Sprintf("tag:%s", idTag)),
						Values: []string{"1"},
					},
				},
			},
			err: errors.New("describe instance error"),
		},
		"no instances with given tag found": {
			id: 1,
			expectedDescribeInstanceInput: &ec2.DescribeInstancesInput{
				Filters: []types.Filter{
					{
						Name:   aws.String(fmt.Sprintf("tag:%s", idTag)),
						Values: []string{"1"},
					},
				},
			},
			err: &runner.NotExistsError{
				ID:   1,
				Type: RunnerType,
			},
		},
	}

	for n, tc := range cases {
		t.Run(n, func(t *testing.T) {
			a := assert.New(t)
			client := &mockedTerminatorClient{
				describeInstancesErr: tc.describeInstancesErr,
				existsInstancesNum:   tc.numInstances,
			}

			a.Equal(tc.err, NewTerminator(client).Terminate(context.TODO(), tc.id))
			a.Equal(tc.expectedDescribeInstanceInput, client.describeInstancesInput)
			a.Equal(tc.expectedTerminateInstancesInput, client.terminateInstancesInput)
		})
	}
}

type mockedTerminatorClient struct {
	terminateInstancesInput *ec2.TerminateInstancesInput
	describeInstancesInput  *ec2.DescribeInstancesInput
	describeInstancesErr    error
	existsInstancesNum      int
}

func (m *mockedTerminatorClient) DescribeInstances(
	_ context.Context,
	input *ec2.DescribeInstancesInput,
	_ ...func(*ec2.Options),
) (*ec2.DescribeInstancesOutput, error) {
	m.describeInstancesInput = input

	s := make([]types.Instance, m.existsInstancesNum)
	for i := range s {
		s[i].InstanceId = aws.String(strconv.Itoa(i))
	}

	return &ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{
			{Instances: s},
		},
	}, m.describeInstancesErr
}

func (m *mockedTerminatorClient) TerminateInstances(
	_ context.Context,
	input *ec2.TerminateInstancesInput,
	_ ...func(*ec2.Options),
) (*ec2.TerminateInstancesOutput, error) {
	m.terminateInstancesInput = input
	return nil, nil
}
