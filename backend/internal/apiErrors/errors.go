package apiErrors

import (
	"errors"
	"fmt"
	"net/http"
	"os"
)

type Error struct {
	err        error
	StatusCode int
}

func (this *Error) Error() string {
	return this.err.Error()
}

func New(msg string, code int) *Error {
	return &Error{
		err: errors.New(msg), StatusCode: code,
	}
}

func FromError(err error, statusCode int) *Error {
	fmt.Fprintln(os.Stderr, err)
	return &Error{err, statusCode}
}

func Generic(err error) *Error {
	fmt.Fprintln(os.Stderr, err)
	return &Error{errors.New("Algo malió sal."), http.StatusInternalServerError}
}
