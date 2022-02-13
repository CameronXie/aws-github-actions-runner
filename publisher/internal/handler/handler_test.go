package handler

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetupPublisherHandler(t *testing.T) {
	a := assert.New(t)
	ctx := context.TODO()
	err := errors.New("error")
	p := &mockedPublisher{err: err}

	a.Equal(err, SetupPublisherHandler(p)(ctx))
	a.Equal(ctx, p.ctx)
}

type mockedPublisher struct {
	ctx context.Context
	err error
}

func (m *mockedPublisher) Publish(ctx context.Context) error {
	m.ctx = ctx
	return m.err
}
