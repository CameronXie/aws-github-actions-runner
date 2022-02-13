package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
)

type SQSEventHandler func(ctx context.Context, event events.SQSEvent) error

func getSQSEventMessage(event events.SQSEvent) (string, error) {
	if n := len(event.Records); n != 1 {
		return "", fmt.Errorf("sqs event sourcing configure issue, received %v messages", n)
	}

	return event.Records[0].Body, nil
}

func getInput(event events.SQSEvent, input interface{}) error {
	msg, err := getSQSEventMessage(event)
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(msg), input)
}
