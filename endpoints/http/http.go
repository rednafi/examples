// Package http is the HTTP wiring for the greet domain. It contains:
//
//  1. Wrap, the generic adapter that turns a (ctx, In) -> (Out, error)
//     function into an http.Handler.
//  2. Per-endpoint decode and encode functions.
//  3. Register, which mounts every endpoint onto an *http.ServeMux.
//  4. Error mapping from greet error codes to HTTP statuses.
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

	"github.com/rednafi/examples/endpoints/greet"
)

type validator interface{ Validate() error }

// Wrap turns a domain function into an HTTP handler. The decode and encode
// callbacks are the only place HTTP-specific types appear at the call site.
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
		Name      string `json:"name"`
		Formality int    `json:"formality"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return greet.GreetIn{}, greet.Invalid("malformed json")
	}
	return greet.GreetIn{Name: body.Name, Formality: body.Formality}, nil
}

func encodeGreet(w http.ResponseWriter, out greet.GreetOut) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(struct {
		Message string `json:"message"`
	}{out.Message})
}

func decodeSubscribe(r *http.Request) (greet.SubscribeIn, error) {
	var body struct {
		Email     string `json:"email"`
		Formality int    `json:"formality"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return greet.SubscribeIn{}, greet.Invalid("malformed json")
	}
	return greet.SubscribeIn{Email: body.Email, Formality: body.Formality}, nil
}

func encodeSubscribe(w http.ResponseWriter, out greet.SubscribeOut) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	return json.NewEncoder(w).Encode(struct {
		ID string `json:"id"`
	}{out.ID})
}

// Register mounts every greet endpoint on mux.
func Register(mux *http.ServeMux, svc *greet.Service) {
	mux.Handle("POST /greet", Wrap(decodeGreet, svc.Greet, encodeGreet))
	mux.Handle("POST /subscribe", Wrap(decodeSubscribe, svc.Subscribe, encodeSubscribe))
}

func writeErr(w http.ResponseWriter, err error) {
	var de *greet.Error
	if !errors.As(err, &de) {
		de = greet.Internal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusFor(de.Code))
	_ = json.NewEncoder(w).Encode(map[string]string{"message": de.Message})
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
