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

func Invalid(msg string) *Error          { return &Error{Code: CodeInvalidArgument, Message: msg} }
func NotFound(msg string) *Error         { return &Error{Code: CodeNotFound, Message: msg} }
func AlreadyExists(msg string) *Error    { return &Error{Code: CodeAlreadyExists, Message: msg} }
func PermissionDenied(msg string) *Error { return &Error{Code: CodePermissionDenied, Message: msg} }
func Unauthenticated(msg string) *Error  { return &Error{Code: CodeUnauthenticated, Message: msg} }
func Internal(err error) *Error          { return &Error{Code: CodeInternal, Message: "internal error", Err: err} }
