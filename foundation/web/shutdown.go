package web

import "errors"

type shutdownError struct {
	Message string
}

func (sh *shutdownError) Error() string {
	return sh.Message
}

func NewShutdownError(msg string) error {
	return &shutdownError{msg}
}

func IsShutDownError(err error) bool {
	var se *shutdownError
	return errors.As(err, &se)
}
