package storage

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type GetJobsInput struct {
	Host     string
	Statuses []string
	Limit    int32
}

type UpdateJob struct {
	ID     uint64
	Status string
}

type UpdateJobsInput struct {
	Update []UpdateJob
	Delete []uint64
}

type Storage interface {
	GetJobs(ctx context.Context, input *GetJobsInput) ([]Job, error)
	UpdateJobs(ctx context.Context, input *UpdateJobsInput) error
}

type storage struct {
	client    *dynamodb.Client
	table     string
	hostIndex string
}

func (s *storage) GetJobs(ctx context.Context, input *GetJobsInput) ([]Job, error) {
	keys := make([]string, 0)
	values := map[string]types.AttributeValue{
		":h": &types.AttributeValueMemberS{Value: input.Host},
	}

	for i := range input.Statuses {
		key := fmt.Sprintf(`:s%v`, i)
		values[key] = &types.AttributeValueMemberS{
			Value: input.Statuses[i],
		}
		keys = append(keys, key)
	}

	o, err := s.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(s.table),
		IndexName:              aws.String(s.hostIndex),
		Limit:                  aws.Int32(input.Limit),
		KeyConditionExpression: aws.String("#h = :h"),
		FilterExpression:       aws.String(fmt.Sprintf("#s IN (%v)", strings.Join(keys, ","))),
		ProjectionExpression:   aws.String("ID,OS,Content,#s,#h"),
		ExpressionAttributeNames: map[string]string{
			"#h": "Host",
			"#s": "Status",
		},
		ExpressionAttributeValues: values,
	})

	if err != nil {
		return nil, err
	}

	jobs := make([]Job, 0)
	if err := attributevalue.UnmarshalListOfMaps(o.Items, &jobs); err != nil {
		return nil, err
	}

	return jobs, nil
}

func (s *storage) UpdateJobs(ctx context.Context, input *UpdateJobsInput) error {
	u := make([]types.TransactWriteItem, 0)
	for _, v := range input.Update {
		u = append(u, types.TransactWriteItem{
			Update: &types.Update{
				TableName: aws.String(s.table),
				Key: map[string]types.AttributeValue{
					"ID": &types.AttributeValueMemberN{
						Value: uint64ToString(v.ID),
					},
				},
				UpdateExpression:    aws.String("SET #s = :s"),
				ConditionExpression: aws.String("attribute_exists(ID)"),
				ExpressionAttributeNames: map[string]string{
					"#s": "Status",
				},
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":s": &types.AttributeValueMemberS{Value: v.Status},
				},
			},
		})
	}

	d := make([]types.TransactWriteItem, 0)
	for _, v := range input.Delete {
		u = append(u, types.TransactWriteItem{
			Delete: &types.Delete{
				TableName: aws.String(s.table),
				Key: map[string]types.AttributeValue{
					"ID": &types.AttributeValueMemberN{
						Value: uint64ToString(v),
					},
				},
			},
		})
	}

	_, err := s.client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: append(u, d...),
	})

	return err
}

func uint64ToString(n uint64) string {
	base := 10
	return strconv.FormatUint(n, base)
}

func New(client *dynamodb.Client, table, index string) Storage {
	return &storage{
		client:    client,
		table:     table,
		hostIndex: index,
	}
}
