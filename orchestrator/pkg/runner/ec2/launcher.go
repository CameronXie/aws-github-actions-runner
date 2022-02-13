package ec2

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"text/template"

	"github.com/CameronXie/aws-github-actions-runner/orchestrator/pkg/runner"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type LaunchConfig struct {
	TemplateID       string
	TemplateVersion  string
	SubnetID         string
	GitHubToken      string
	RunnerVersion    string
	UserDataTemplate *template.Template
}

type templateData struct {
	ID            int
	Owner         string
	Repository    string
	GitHubToken   string
	RunnerName    string
	RunnerVersion string
	RunnerLabels  string
}

type RunInstancesAPIClient interface {
	ec2.DescribeInstancesAPIClient
	RunInstances(ctx context.Context, params *ec2.RunInstancesInput, optFns ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error)
}

type ec2Launcher struct {
	runnerNamePrefix string
	client           RunInstancesAPIClient
	config           *LaunchConfig
}

func (l *ec2Launcher) Launch(ctx context.Context, input *runner.LaunchInput) error {
	ids, rErr := getInstanceIDByTag(l.client, ctx, idTag, []string{strconv.Itoa(input.ID)})

	if rErr != nil {
		return rErr
	}

	if len(ids) != 0 {
		return &runner.AlreadyExistsError{
			ID:   input.ID,
			Type: RunnerType,
		}
	}

	i := &ec2.RunInstancesInput{
		MaxCount: aws.Int32(1),
		MinCount: aws.Int32(1),
		LaunchTemplate: &types.LaunchTemplateSpecification{
			LaunchTemplateId: aws.String(l.config.TemplateID),
			Version:          aws.String(l.config.TemplateVersion),
		},
		SubnetId: aws.String(l.config.SubnetID),
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceTypeInstance,
				Tags: []types.Tag{
					{
						Key:   aws.String(idTag),
						Value: aws.String(strconv.Itoa(input.ID)),
					},
				},
			},
		},
	}

	if l.config.UserDataTemplate != nil {
		userData := new(bytes.Buffer)
		if err := l.config.UserDataTemplate.Execute(userData, templateData{
			ID:            input.ID,
			Owner:         input.Owner,
			Repository:    input.Repository,
			GitHubToken:   l.config.GitHubToken,
			RunnerName:    fmt.Sprintf("%v-%v", l.runnerNamePrefix, input.ID),
			RunnerVersion: l.config.RunnerVersion,
			RunnerLabels:  strings.Join(input.Labels, ","),
		}); err != nil {
			return err
		}

		i.UserData = aws.String(base64.StdEncoding.EncodeToString(userData.Bytes()))
	}

	_, launchErr := l.client.RunInstances(ctx, i)

	return launchErr
}

func NewLauncher(
	prefix string,
	client RunInstancesAPIClient,
	config *LaunchConfig,
) runner.Launcher {
	return &ec2Launcher{
		runnerNamePrefix: prefix,
		client:           client,
		config:           config,
	}
}
