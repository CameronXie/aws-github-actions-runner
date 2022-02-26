package handler

import (
	"context"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

func TestSetupHandler(t *testing.T) {
	a := assert.New(t)
	svc := new(mockedMessenger)
	ctx := context.TODO()

	a.Nil(SetupHandler(svc)(ctx, events.DynamoDBEvent{}))
	a.Equal(ctx, svc.ctx)
}

type mockedMessenger struct {
	ctx context.Context
}

func (m *mockedMessenger) NotifyPublisher(ctx context.Context) error {
	m.ctx = ctx
	return nil
}
