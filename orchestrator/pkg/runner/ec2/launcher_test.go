package ec2

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"testing"
	"text/template"

	"github.com/CameronXie/aws-github-actions-runner/orchestrator/pkg/runner"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/stretchr/testify/assert"
)

func TestEc2Launcher_Launch(t *testing.T) {
	runnerPrefix := "prefix"
	cases := map[string]struct {
		config                        *LaunchConfig
		input                         *runner.LaunchInput
		numInstances                  int
		describeInstancesErr          error
		expectedRunInstanceInput      *ec2.RunInstancesInput
		expectedDescribeInstanceInput *ec2.DescribeInstancesInput
		err                           error
	}{
		"valid userdata template": {
			config: &LaunchConfig{
				TemplateID:      "template-id",
				TemplateVersion: "$Latest",
				SubnetID:        "subnet-id",
				GitHubToken:     "pat",
				RunnerVersion:   "1.0.0",
				UserDataTemplate: template.Must(template.New("tests").
					Parse(`{{.Owner}},{{.Repository}},{{.GitHubToken}},{{.RunnerName}},{{.RunnerVersion}},{{.RunnerLabels}}`)),
			},
			input: &runner.LaunchInput{
				ID:         1,
				Owner:      "owner",
				Repository: "repo",
				Labels:     []string{"ec2", "ubuntu"},
			},
			expectedDescribeInstanceInput: &ec2.DescribeInstancesInput{
				Filters: []types.Filter{
					{
						Name:   aws.String(fmt.Sprintf("tag:%s", idTag)),
						Values: []string{"1"},
					},
				},
			},
			expectedRunInstanceInput: &ec2.RunInstancesInput{
				MaxCount: aws.Int32(1),
				MinCount: aws.Int32(1),
				LaunchTemplate: &types.LaunchTemplateSpecification{
					LaunchTemplateId: aws.String("template-id"),
					Version:          aws.String("$Latest"),
				},
				SubnetId: aws.String("subnet-id"),
				TagSpecifications: []types.TagSpecification{
					{
						ResourceType: types.ResourceTypeInstance,
						Tags: []types.Tag{
							{
								Key:   aws.String(idTag),
								Value: aws.String("1"),
							},
						},
					},
				},
				UserData: aws.String(
					base64.StdEncoding.EncodeToString([]byte(`owner,repo,pat,prefix-1,1.0.0,ec2,ubuntu`)),
				),
			},
		},
		"invalid userdata template": {
			config: &LaunchConfig{
				TemplateID:      "template-id",
				TemplateVersion: "$Latest",
				SubnetID:        "subnet-id",
				GitHubToken:     "pat",
				RunnerVersion:   "1.0.0",
				UserDataTemplate: template.Must(template.New("tests").
					Parse(`{{.RandomValue}}`)),
			},
			input: &runner.LaunchInput{
				ID:         1,
				Owner:      "owner",
				Repository: "repo",
				Labels:     []string{"ec2", "ubuntu"},
			},
			expectedDescribeInstanceInput: &ec2.DescribeInstancesInput{
				Filters: []types.Filter{
					{
						Name:   aws.String(fmt.Sprintf("tag:%s", idTag)),
						Values: []string{"1"},
					},
				},
			},
			err: template.ExecError{
				Name: "tests",
				Err:  errors.New(`template: tests:1:2: executing "tests" at <.RandomValue>: can't evaluate field RandomValue in type ec2.templateData`),
			},
		},
		"runner with given tag already exists": {
			numInstances: 1,
			input: &runner.LaunchInput{
				ID:         1,
				Owner:      "owner",
				Repository: "repo",
				Labels:     []string{"ec2", "ubuntu"},
			},
			expectedDescribeInstanceInput: &ec2.DescribeInstancesInput{
				Filters: []types.Filter{
					{
						Name:   aws.String(fmt.Sprintf("tag:%s", idTag)),
						Values: []string{"1"},
					},
				},
			},
			err: &runner.AlreadyExistsError{
				ID:   1,
				Type: RunnerType,
			},
		},
		"describe instances error": {
			describeInstancesErr: errors.New("describe instances error"),
			input: &runner.LaunchInput{
				ID:         1,
				Owner:      "owner",
				Repository: "repo",
				Labels:     []string{"ec2", "ubuntu"},
			},
			expectedDescribeInstanceInput: &ec2.DescribeInstancesInput{
				Filters: []types.Filter{
					{
						Name:   aws.String(fmt.Sprintf("tag:%s", idTag)),
						Values: []string{"1"},
					},
				},
			},
			err: errors.New("describe instances error"),
		},
	}

	for n, tc := range cases {
		t.Run(n, func(t *testing.T) {
			a := assert.New(t)
			client := &mockedLauncherClient{
				describeInstancesErr: tc.describeInstancesErr,
				existsInstancesNum:   tc.numInstances,
			}

			a.Equal(tc.err, NewLauncher(runnerPrefix, client, tc.config).Launch(context.TODO(), tc.input))
			a.Equal(tc.expectedDescribeInstanceInput, client.describeInstancesInput)
			a.Equal(tc.expectedRunInstanceInput, client.instancesInput)
		})
	}
}

type mockedLauncherClient struct {
	instancesInput         *ec2.RunInstancesInput
	describeInstancesInput *ec2.DescribeInstancesInput
	describeInstancesErr   error
	existsInstancesNum     int
}

func (m *mockedLauncherClient) RunInstances(
	_ context.Context,
	input *ec2.RunInstancesInput,
	_ ...func(*ec2.Options),
) (*ec2.RunInstancesOutput, error) {
	m.instancesInput = input
	return nil, nil
}

func (m *mockedLauncherClient) DescribeInstances(
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
			{
				Instances: s,
			},
		},
	}, m.describeInstancesErr
}
