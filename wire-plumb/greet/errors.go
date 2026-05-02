package greet

import "fmt"

// Code names a failure mode in domain terms. The HTTP and gRPC packages
// each map it onto their own status enum.
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

func newErr(c Code, format string, args ...any) *Error {
	return &Error{Code: c, Message: fmt.Sprintf(format, args...)}
}

func Invalid(format string, args ...any) *Error {
	return newErr(CodeInvalidArgument, format, args...)
}

func NotFound(format string, args ...any) *Error {
	return newErr(CodeNotFound, format, args...)
}

func AlreadyExists(format string, args ...any) *Error {
	return newErr(CodeAlreadyExists, format, args...)
}

func PermissionDenied(format string, args ...any) *Error {
	return newErr(CodePermissionDenied, format, args...)
}

func Unauthenticated(format string, args ...any) *Error {
	return newErr(CodeUnauthenticated, format, args...)
}

// Internal wraps an underlying error and never originates from a domain rule.
func Internal(err error) *Error {
	return &Error{Code: CodeInternal, Message: "internal error", Err: err}
}
