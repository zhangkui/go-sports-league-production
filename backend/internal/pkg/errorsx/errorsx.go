package errorsx

import (
	"errors"
	"fmt"
)

// AppError carries a business error code plus an HTTP status hint.
type AppError struct {
	Code     int
	HTTPStatus int
	Message  string
	Wrapped  error
}

func (e *AppError) Error() string {
	if e.Wrapped != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Wrapped)
	}
	return e.Message
}

func (e *AppError) Unwrap() error { return e.Wrapped }

func New(httpStatus, code int, msg string) *AppError {
	return &AppError{Code: code, HTTPStatus: httpStatus, Message: msg}
}

func Wrap(httpStatus, code int, msg string, err error) *AppError {
	return &AppError{Code: code, HTTPStatus: httpStatus, Message: msg, Wrapped: err}
}

// Convenience constructors.
func BadRequest(msg string) *AppError       { return New(400, 40000, msg) }
func Unauthorized(msg string) *AppError     { return New(401, 40100, msg) }
func Forbidden(msg string) *AppError        { return New(403, 40300, msg) }
func NotFound(msg string) *AppError         { return New(404, 40400, msg) }
func NotFoundID(resource string, id any) *AppError {
	return New(404, 40400, fmt.Sprintf("%s %v not found", resource, id))
}
func Conflict(msg string) *AppError        { return New(409, 40900, msg) }
func TooMany(msg string) *AppError          { return New(429, 42900, msg) }
func Internal(msg string) *AppError        { return New(500, 50000, msg) }

// IsNotFound reports whether err is a not-found app error.
func IsNotFound(err error) bool {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae.HTTPStatus == 404
	}
	return errors.Is(err, ErrNotFound)
}

// IsConflict reports a 409-level conflict.
func IsConflict(err error) bool {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae.HTTPStatus == 409
	}
	return false
}

var ErrNotFound = errors.New("not found")
var ErrConflict = errors.New("conflict")
