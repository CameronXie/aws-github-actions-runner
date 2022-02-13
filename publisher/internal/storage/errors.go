package storage

import (
	"errors"
	"fmt"
)

type InvalidJobContentError struct {
	Type string
	Err  error
}

func (e *InvalidJobContentError) Error() string {
	return fmt.Sprintf(`%v: %v`, e.Type, e.Err.Error())
}

func IsInvalidJobContentError(err error) bool {
	var e *InvalidJobContentError
	return errors.As(err, &e)
}
