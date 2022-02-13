package runner

import (
	"errors"
	"fmt"
)

type AlreadyExistsError struct {
	ID   int
	Type string
}

func (e *AlreadyExistsError) Error() string {
	return fmt.Sprintf(`runner id: %v type: %v already exists`, e.ID, e.Type)
}

func IsAlreadyExistsError(err error) bool {
	e := new(AlreadyExistsError)
	return errors.As(err, &e)
}

type NotExistsError struct {
	ID   int
	Type string
}

func (e *NotExistsError) Error() string {
	return fmt.Sprintf(`runner id: %v type: %v not exists`, e.ID, e.Type)
}

func IsNotExistsError(err error) bool {
	e := new(NotExistsError)
	return errors.As(err, &e)
}
