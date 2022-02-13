package storage

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInvalidJobContentError_Error(t *testing.T) {
	a := assert.New(t)
	e := &InvalidJobContentError{
		Type: "error_type",
		Err:  errors.New("new error"),
	}

	a.Equal("error_type: new error", e.Error())
}

func TestIsInvalidJobContentError(t *testing.T) {
	cases := map[string]struct {
		err error
		res bool
	}{
		"invalid job content error": {
			err: &InvalidJobContentError{
				Type: "error_type",
				Err:  errors.New("new error"),
			},
			res: true,
		},
		"strings error": {
			err: errors.New("new error"),
		},
	}

	for n, tc := range cases {
		t.Run(n, func(t *testing.T) {
			a := assert.New(t)
			a.Equal(tc.res, IsInvalidJobContentError(tc.err))
		})
	}
}
