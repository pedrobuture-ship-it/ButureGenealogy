// Package apperr defines application errors with stable codes that the
// frontend can map to user-friendly messages.
package apperr

import (
	"errors"
	"fmt"
)

type Code string

const (
	CodeInternal   Code = "internal"
	CodeNotFound   Code = "not_found"
	CodeValidation Code = "validation"
	CodeConflict   Code = "conflict"
	CodeDatabase   Code = "database"
)

// Error is an error carrying a stable code and a user-facing message.
type Error struct {
	Code    Code
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error { return e.Err }

func New(code Code, msg string) *Error { return &Error{Code: code, Message: msg} }

func Wrap(code Code, msg string, err error) *Error {
	return &Error{Code: code, Message: msg, Err: err}
}

// CodeOf returns the code of err, or CodeInternal for foreign errors.
func CodeOf(err error) Code {
	var e *Error
	if errors.As(err, &e) {
		return e.Code
	}
	return CodeInternal
}

// UserMessage returns a message safe to show in the UI.
func UserMessage(err error) string {
	var e *Error
	if errors.As(err, &e) {
		return e.Message
	}
	return "Ocorreu um erro inesperado."
}
