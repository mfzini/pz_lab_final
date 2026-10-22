package api

import (
	"errors"
	"fmt"
	"net/http"
	"os"
)

// Error es el error de negocio que viaja en la respuesta JSON: siempre como
// {"error": "<mensaje>"}. El detalle interno queda en err (sin serializar).
type Error struct {
	err        error
	Message    string `json:"error"`
	StatusCode int    `json:"-"`
}

func (this *Error) Error() string {
	return this.err.Error()
}

func NewError(msg string, code int) *Error {
	return &Error{
		err: errors.New(msg), Message: msg, StatusCode: code,
	}
}

// NewGenericError es para fallas de infraestructura: loguea el detalle real y
// devuelve un mensaje genérico (nada de errores de DB en la respuesta).
func NewGenericError(err error) *Error {
	fmt.Fprintln(os.Stderr, err)
	return &Error{
		err: errors.New("Algo malió sal."), Message: "Algo malió sal.", StatusCode: http.StatusInternalServerError,
	}
}
