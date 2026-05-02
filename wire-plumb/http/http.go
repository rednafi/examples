// Package http is the HTTP wiring for the greet domain.
//
// Package name http shadows net/http for outside importers; alias this
// package as ehttp at the call site and leave net/http plain.
package http

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/rednafi/examples/wire-plumb/greet"
)

type validator interface{ Validate() error }

// Wrap turns a domain function into an HTTP handler. The decode and encode
// callbacks are the only place HTTP-specific types appear.
func Wrap[In, Out any](
	decode func(*http.Request) (In, error),
	fn func(context.Context, In) (Out, error),
	encode func(http.ResponseWriter, Out) error,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		in, err := decode(r)
		if err != nil {
			writeErr(w, err)
			return
		}

		if v, ok := any(in).(validator); ok {
			if err := v.Validate(); err != nil {
				writeErr(w, err)
				return
			}
		}

		out, err := fn(r.Context(), in)
		if err != nil {
			writeErr(w, err)
			return
		}

		if err := encode(w, out); err != nil {
			log.Printf("encode response: %v", err)
		}
	})
}

func decodeGreet(r *http.Request) (greet.GreetIn, error) {
	var body struct {
		UserID    int64 `json:"user_id"`
		Formality int   `json:"formality"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return greet.GreetIn{}, greet.Invalid("malformed json")
	}
	return greet.GreetIn{UserID: body.UserID, Formality: body.Formality}, nil
}

func encodeGreet(w http.ResponseWriter, out greet.GreetOut) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(struct {
		Message string `json:"message"`
	}{out.Message})
}

func decodeFarewell(r *http.Request) (greet.FarewellIn, error) {
	var body struct {
		UserID int64 `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return greet.FarewellIn{}, greet.Invalid("malformed json")
	}
	return greet.FarewellIn{UserID: body.UserID}, nil
}

func encodeFarewell(w http.ResponseWriter, out greet.FarewellOut) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(struct {
		Message string `json:"message"`
	}{out.Message})
}

func Register(mux *http.ServeMux, svc *greet.Service) {
	mux.Handle("POST /greet", Wrap(decodeGreet, svc.Greet, encodeGreet))
	mux.Handle("POST /farewell", Wrap(decodeFarewell, svc.Farewell, encodeFarewell))
}

func writeErr(w http.ResponseWriter, err error) {
	var domainErr *greet.Error
	if !errors.As(err, &domainErr) {
		domainErr = greet.Internal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusFor(domainErr.Code))
	_ = json.NewEncoder(w).Encode(map[string]string{"message": domainErr.Message})
}

func statusFor(c greet.Code) int {
	switch c {
	case greet.CodeInvalidArgument:
		return http.StatusBadRequest
	case greet.CodeUnauthenticated:
		return http.StatusUnauthorized
	case greet.CodePermissionDenied:
		return http.StatusForbidden
	case greet.CodeNotFound:
		return http.StatusNotFound
	case greet.CodeAlreadyExists:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
