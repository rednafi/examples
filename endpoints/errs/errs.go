// Package errs defines transport-agnostic error codes and a small error type.
// Business code returns these so handlers can map them to HTTP statuses or
// gRPC codes without coupling to either.
package errs

import "fmt"

type Code int

const (
	CodeInvalidArgument Code = iota
	CodeUnauthenticated
	CodePermissionDenied
	CodeNotFound
	CodeAlreadyExists
	CodeInternal
)

type Error struct {
	Code    Code
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *Error) Unwrap() error { return e.Err }

func Invalid(msg string) *Error          { return &Error{Code: CodeInvalidArgument, Message: msg} }
func NotFound(msg string) *Error         { return &Error{Code: CodeNotFound, Message: msg} }
func AlreadyExists(msg string) *Error    { return &Error{Code: CodeAlreadyExists, Message: msg} }
func PermissionDenied(msg string) *Error { return &Error{Code: CodePermissionDenied, Message: msg} }
func Unauthenticated(msg string) *Error  { return &Error{Code: CodeUnauthenticated, Message: msg} }
func Internal(err error) *Error          { return &Error{Code: CodeInternal, Message: "internal error", Err: err} }
