package api

import (
	"errors"
	"fmt"
	"net/http"
	"os"
)

type Error struct {
	err        error
	StatusCode int `json:"-"`
}

func (this *Error) Error() string {
	return this.err.Error()
}

func NewError(msg string, code int) *Error {
	return &Error{
		err: errors.New(msg), StatusCode: code,
	}
}

func NewGenericError(err error) *Error {
	fmt.Fprintln(os.Stderr, err)
	return &Error{errors.New("Algo malió sal."), http.StatusInternalServerError}
}
