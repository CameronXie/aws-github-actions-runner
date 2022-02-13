package storage

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
)

func TestJob_UnmarshalDynamoDBAttributeValue(t *testing.T) {
	id, host, status := 123, "host", "status"
	cases := map[string]struct {
		av       types.AttributeValue
		expected *Job
		errType  string
	}{
		"dynamodb item to job": {
			av: getDynamoDBItem(
				id,
				host,
				status,
				getCompressedContent(JobContent{
					ID:         id,
					Owner:      "owner",
					Repository: "repo",
					Labels:     []string{"ec2", "eks"},
				}),
			),
			expected: &Job{
				ID:     id,
				Host:   host,
				Status: status,
				Content: JobContent{
					ID:         id,
					Owner:      "owner",
					Repository: "repo",
					Labels:     []string{"ec2", "eks"},
				},
			},
		},
		"item with invalid gzip content": {
			av: getDynamoDBItem(
				id,
				host,
				status,
				[]byte("random"),
			),
			errType: InvalidGZIPType,
		},
		"item with invalid json content": {
			av: getDynamoDBItem(
				id,
				host,
				status,
				getCompressedStr(`{`),
			),
			errType: InvalidJSONType,
		},
		"item with empty json content": {
			av: getDynamoDBItem(
				id,
				host,
				status,
				getCompressedStr(`{}`),
			),
			expected: &Job{
				ID:      id,
				Host:    host,
				Status:  status,
				Content: JobContent{},
			},
		},
		"invalid item": {
			av:       new(types.AttributeValueMemberL),
			expected: new(Job),
		},
	}

	for n, tc := range cases {
		t.Run(n, func(t *testing.T) {
			a := assert.New(t)
			res := new(Job)
			err := res.UnmarshalDynamoDBAttributeValue(tc.av)

			if tc.errType != "" {
				a.Equal(new(Job), res)
				a.Equal(tc.errType, err.(*InvalidJobContentError).Type)
				return
			}

			a.Nil(err)
			a.EqualValues(tc.expected, res)
		})
	}
}

//nolint:unparam
func getDynamoDBItem(
	id int,
	host string,
	status string,
	content []byte,
) types.AttributeValue {
	av, _ := attributevalue.MarshalMap(struct {
		ID      int
		Host    string
		Status  string
		Content []byte
	}{
		ID:      id,
		Host:    host,
		Status:  status,
		Content: content,
	})

	return &types.AttributeValueMemberM{Value: av}
}

func getCompressedContent(c JobContent) []byte {
	o := new(bytes.Buffer)
	w := gzip.NewWriter(o)
	_ = json.NewEncoder(w).Encode(c)
	_ = w.Close()

	return o.Bytes()
}

func getCompressedStr(str string) []byte {
	o := new(bytes.Buffer)
	w := gzip.NewWriter(o)
	_, _ = w.Write([]byte(str))
	_ = w.Close()

	return o.Bytes()
}
