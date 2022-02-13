package storage

import (
	"bytes"
	"compress/gzip"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const (
	InvalidGZIPType = "invalid_gzip"
	InvalidJSONType = "invalid_json"
)

type JobContent struct {
	ID         int
	Owner      string
	Repository string
	Labels     []string
}

type Job struct {
	ID      int
	Host    string
	Status  string
	Content JobContent
}

func (j *Job) UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error {
	m, ok := av.(*types.AttributeValueMemberM)
	if !ok {
		return nil
	}

	raw := new(struct {
		ID      int
		Host    string
		Status  string
		Content []byte
	})

	_ = attributevalue.UnmarshalMap(m.Value, raw)

	var content JobContent
	r, zErr := gzip.NewReader(bytes.NewReader(raw.Content))
	if zErr != nil {
		return &InvalidJobContentError{Type: InvalidGZIPType, Err: zErr}
	}

	if err := json.NewDecoder(r).Decode(&content); err != nil {
		return &InvalidJobContentError{Type: InvalidJSONType, Err: err}
	}

	j.ID = raw.ID
	j.Host = raw.Host
	j.Status = raw.Status
	j.Content = content
	return nil
}
