package runner

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAlreadyExistsError_Error(t *testing.T) {
	a := assert.New(t)
	a.Equal("runner id: 1 type: resource already exists", (&AlreadyExistsError{
		ID:   1,
		Type: "resource",
	}).Error())
}

func TestIsAlreadyExistsError(t *testing.T) {
	cases := map[string]struct {
		err      error
		expected bool
	}{
		"AlreadyExistsError": {
			err: &AlreadyExistsError{
				ID:   1,
				Type: "resource",
			},
			expected: true,
		},
		"errorString": {
			err:      errors.New("new errorString"),
			expected: false,
		},
	}

	for n, tc := range cases {
		t.Run(n, func(t *testing.T) {
			a := assert.New(t)
			a.Equal(tc.expected, IsAlreadyExistsError(tc.err))
		})
	}
}

func TestNotExistsError_Error(t *testing.T) {
	a := assert.New(t)
	a.Equal("runner id: 1 type: resource not exists", (&NotExistsError{
		ID:   1,
		Type: "resource",
	}).Error())
}

func TestIsNotExistsError(t *testing.T) {
	cases := map[string]struct {
		err      error
		expected bool
	}{
		"NotExistsError": {
			err: &NotExistsError{
				ID:   1,
				Type: "resource",
			},
			expected: true,
		},
		"errorString": {
			err:      errors.New("new errorString"),
			expected: false,
		},
	}

	for n, tc := range cases {
		t.Run(n, func(t *testing.T) {
			a := assert.New(t)
			a.Equal(tc.expected, IsNotExistsError(tc.err))
		})
	}
}
