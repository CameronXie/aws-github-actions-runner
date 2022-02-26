package storage

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/smithy-go"
	"github.com/stretchr/testify/suite"
)

const (
	mockedDynamodbEndpoint = "http://dynamodb:8000"
	testTable              = "publisher-storage-test"
	hostIndex              = "HostIndex"
	testJobsNum            = 10
)

type storageSuite struct {
	suite.Suite
	client    *dynamodb.Client
	table     string
	endpoint  string
	hostIndex string
	testJobs  []Job
}

func (s *storageSuite) TestStorage_GetJobs() {
	cases := map[string]struct {
		getJobsInput *GetJobsInput
		expected     []Job
		errType      interface{}
	}{
		"get jobs": {
			getJobsInput: &GetJobsInput{
				Host:     "ec2",
				Statuses: []string{"queued", "completed"},
				Limit:    5,
			},
			// nolint: dupl
			expected: []Job{
				{
					ID:     1,
					Host:   "ec2",
					OS:     "ubuntu",
					Status: "queued",
					Content: JobContent{
						ID:         1,
						Owner:      "owner-1",
						Repository: "repo-1",
						Labels:     []string{"ec2", "ubuntu"},
					},
				},
				{
					ID:     2,
					Host:   "ec2",
					OS:     "ubuntu",
					Status: "completed",
					Content: JobContent{
						ID:         2,
						Owner:      "owner-2",
						Repository: "repo-2",
						Labels:     []string{"ec2", "ubuntu"},
					},
				},
				{
					ID:     4,
					Host:   "ec2",
					OS:     "ubuntu",
					Status: "completed",
					Content: JobContent{
						ID:         4,
						Owner:      "owner-4",
						Repository: "repo-4",
						Labels:     []string{"ec2", "ubuntu"},
					},
				},
			},
		},
		"invalid input": {
			getJobsInput: &GetJobsInput{
				Host:     "ec2",
				Statuses: []string{"queued", "completed"},
				Limit:    0,
			},
			errType: new(smithy.OperationError),
		},
	}

	for n, tc := range cases {
		s.T().Run(n, func(t *testing.T) {
			jobs, err := New(s.client, s.table, s.hostIndex).GetJobs(
				context.TODO(),
				tc.getJobsInput,
			)

			if tc.errType == nil {
				s.Nil(err)
				s.EqualValues(tc.expected, jobs)
				return
			}

			s.ErrorAs(err, &tc.errType)
		})
	}
}

func (s *storageSuite) TestStorage_UpdateJobs() {
	db := New(s.client, s.table, s.hostIndex)
	cases := map[string]struct {
		update   *UpdateJobsInput
		expected []Job
		err      error
	}{
		"update and delete jobs": {
			update: &UpdateJobsInput{
				Update: []UpdateJob{
					{ID: 3, Status: "updated"},
					{ID: 4, Status: "updated"},
				},
				Delete: []uint64{0, 1, 2},
			},
			// nolint: dupl
			expected: []Job{
				{
					ID:     3,
					Host:   "ec2",
					OS:     "ubuntu",
					Status: "updated",
					Content: JobContent{
						ID:         3,
						Owner:      "owner-3",
						Repository: "repo-3",
						Labels:     []string{"ec2", "ubuntu"},
					},
				},
				{
					ID:     4,
					Host:   "ec2",
					OS:     "ubuntu",
					Status: "updated",
					Content: JobContent{
						ID:         4,
						Owner:      "owner-4",
						Repository: "repo-4",
						Labels:     []string{"ec2", "ubuntu"},
					},
				},
				{
					ID:     5,
					Host:   "ec2",
					OS:     "ubuntu",
					Status: "queued",
					Content: JobContent{
						ID:         5,
						Owner:      "owner-5",
						Repository: "repo-5",
						Labels:     []string{"ec2", "ubuntu"},
					},
				},
			},
		},
	}

	for n, tc := range cases {
		s.T().Run(n, func(t *testing.T) {
			err := db.UpdateJobs(context.TODO(), tc.update)
			if err != nil {
				s.Equal(tc.err, err)
				return
			}

			jobs, err := db.GetJobs(context.TODO(), &GetJobsInput{
				Host:     "ec2",
				Statuses: []string{"queued", "completed", "updated"},
				Limit:    3,
			})

			s.Equal(tc.err, err)
			s.EqualValues(tc.expected, jobs)
		})
	}
}

func (s *storageSuite) SetupSuite() {
	cfg, _ := config.LoadDefaultConfig(
		context.TODO(),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{URL: s.endpoint}, nil
			},
		)),
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{AccessKeyID: "AccessKeyID", SecretAccessKey: "SecretAccessKey"},
		}),
	)

	s.client = dynamodb.NewFromConfig(cfg)
	s.tearDownTestTable()
	s.setupTestTable()
}

func (s *storageSuite) SetupTest() {
	requests := make([]types.WriteRequest, 0)

	for _, j := range s.testJobs {
		tmp := j
		requests = append(requests, types.WriteRequest{
			PutRequest: getPutRequestFromJob(&tmp),
		})
	}

	s.handleError(s.client.BatchWriteItem(context.TODO(), &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			s.table: requests,
		},
	}))
}

func (s *storageSuite) TearDownTest() {
	requests := make([]types.WriteRequest, 0)
	for _, j := range s.testJobs {
		requests = append(requests, types.WriteRequest{
			DeleteRequest: &types.DeleteRequest{
				Key: map[string]types.AttributeValue{
					"ID": &types.AttributeValueMemberN{
						Value: uint64ToString(j.ID),
					},
				},
			},
		})
	}

	s.handleError(s.client.BatchWriteItem(context.TODO(), &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			s.table: requests,
		},
	}))
}

func (s *storageSuite) TearDownSuite() {
	s.tearDownTestTable()
}

func (s *storageSuite) setupTestTable() {
	s.handleError(s.client.CreateTable(context.TODO(), &dynamodb.CreateTableInput{
		TableName: aws.String(s.table),
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("ID"),
				AttributeType: "N",
			},
			{
				AttributeName: aws.String("Host"),
				AttributeType: "S",
			},
			{
				AttributeName: aws.String("CreatedAt"),
				AttributeType: "N",
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("ID"),
				KeyType:       "HASH",
			},
		},
		GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String(hostIndex),
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("Host"),
						KeyType:       "HASH",
					},
					{
						AttributeName: aws.String("CreatedAt"),
						KeyType:       "RANGE",
					},
				},
				Projection: &types.Projection{
					NonKeyAttributes: []string{"OS", "Content", "Status"},
					ProjectionType:   types.ProjectionTypeInclude,
				},
			},
		},
		BillingMode: types.BillingModePayPerRequest,
	}))
}

func (s *storageSuite) tearDownTestTable() {
	_, _ = s.client.DeleteTable(
		context.TODO(),
		&dynamodb.DeleteTableInput{TableName: aws.String(s.table)},
	)
}

func (s *storageSuite) handleError(_ interface{}, err error) {
	if err != nil {
		s.T().Error(err)
	}
}

func getTestJobs(num int) []Job {
	jobs := make([]Job, num)

	for i := range jobs {
		status := "queued"
		if i%2 == 0 {
			status = "completed"
		}

		if i%3 == 0 {
			status = "in_progress"
		}

		jobs[i] = Job{
			ID:     uint64(i),
			Host:   "ec2",
			OS:     "ubuntu",
			Status: status,
			Content: JobContent{
				ID:         uint64(i),
				Owner:      fmt.Sprintf("owner-%v", i),
				Repository: fmt.Sprintf("repo-%v", i),
				Labels:     []string{"ec2", "ubuntu"},
			},
		}
	}

	return jobs
}

func getPutRequestFromJob(job *Job) *types.PutRequest {
	return &types.PutRequest{
		Item: map[string]types.AttributeValue{
			"ID": &types.AttributeValueMemberN{
				Value: uint64ToString(job.ID),
			},
			"Host": &types.AttributeValueMemberS{
				Value: job.Host,
			},
			"OS": &types.AttributeValueMemberS{
				Value: job.OS,
			},
			"Content": &types.AttributeValueMemberB{
				Value: getCompressedContent(job.Content),
			},
			"Status": &types.AttributeValueMemberS{
				Value: job.Status,
			},
			"CreatedAt": &types.AttributeValueMemberN{
				Value: strconv.Itoa(int(time.Now().UnixMilli())),
			},
		},
	}
}

func TestStorageSuite(t *testing.T) {
	suite.Run(t, &storageSuite{
		table:     testTable,
		endpoint:  mockedDynamodbEndpoint,
		hostIndex: hostIndex,
		testJobs:  getTestJobs(testJobsNum),
	})
}
