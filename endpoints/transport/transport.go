// Package transport contains the two generic adapters that turn a
// (ctx, In) -> (Out, error) function into something http.ServeMux or a
// gRPC server can register. WrapHTTP and WrapGRPC do the same five steps
// in the same order; only the wire-level mechanics differ.
package transport

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/rednafi/examples/endpoints/errs"
)

type validator interface{ Validate() error }

// WrapHTTP turns a domain function into an http.Handler. The decode and
// encode callbacks are the only places that touch HTTP-specific types.
func WrapHTTP[In, Out any](
	decode func(*http.Request) (In, error),
	fn func(context.Context, In) (Out, error),
	encode func(http.ResponseWriter, Out) error,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		in, err := decode(r)
		if err != nil {
			writeHTTPErr(w, err)
			return
		}

		if v, ok := any(in).(validator); ok {
			if err := v.Validate(); err != nil {
				writeHTTPErr(w, err)
				return
			}
		}

		out, err := fn(r.Context(), in)
		if err != nil {
			writeHTTPErr(w, err)
			return
		}

		if err := encode(w, out); err != nil {
			log.Printf("encode response: %v", err)
		}
	})
}

// WrapGRPC turns a domain function into a function with the shape
// google.golang.org/grpc generates for unary RPCs. Plug the result into
// the generated server method body.
func WrapGRPC[WireIn, In, Out, WireOut any](
	decode func(WireIn) (In, error),
	fn func(context.Context, In) (Out, error),
	encode func(Out) (WireOut, error),
) func(context.Context, WireIn) (WireOut, error) {
	return func(ctx context.Context, wireIn WireIn) (WireOut, error) {
		var zero WireOut

		in, err := decode(wireIn)
		if err != nil {
			return zero, toGRPCErr(err)
		}

		if v, ok := any(in).(validator); ok {
			if err := v.Validate(); err != nil {
				return zero, toGRPCErr(err)
			}
		}

		out, err := fn(ctx, in)
		if err != nil {
			return zero, toGRPCErr(err)
		}

		return encode(out)
	}
}

func writeHTTPErr(w http.ResponseWriter, err error) {
	var de *errs.Error
	if !errors.As(err, &de) {
		de = errs.Internal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus(de.Code))
	_ = json.NewEncoder(w).Encode(map[string]string{"message": de.Message})
}

func httpStatus(c errs.Code) int {
	switch c {
	case errs.CodeInvalidArgument:
		return http.StatusBadRequest
	case errs.CodeUnauthenticated:
		return http.StatusUnauthorized
	case errs.CodePermissionDenied:
		return http.StatusForbidden
	case errs.CodeNotFound:
		return http.StatusNotFound
	case errs.CodeAlreadyExists:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

func toGRPCErr(err error) error {
	var de *errs.Error
	if !errors.As(err, &de) {
		return status.Error(codes.Internal, "internal error")
	}
	return status.Error(grpcCode(de.Code), de.Message)
}

func grpcCode(c errs.Code) codes.Code {
	switch c {
	case errs.CodeInvalidArgument:
		return codes.InvalidArgument
	case errs.CodeUnauthenticated:
		return codes.Unauthenticated
	case errs.CodePermissionDenied:
		return codes.PermissionDenied
	case errs.CodeNotFound:
		return codes.NotFound
	case errs.CodeAlreadyExists:
		return codes.AlreadyExists
	default:
		return codes.Internal
	}
}
