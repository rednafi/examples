// Package http is the HTTP wiring for the greet domain. It contains:
//
//  1. Wrap, the generic adapter that turns a (ctx, In) -> (Out, error)
//     function into an http.Handler.
//  2. Per-endpoint decode and encode functions.
//  3. Register, which mounts every endpoint onto an *http.ServeMux.
//  4. Error mapping from greet error codes to HTTP statuses.
//
// Importers must alias this package because it shadows net/http; see cmd/server.
package http

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	nethttp "net/http"

	"github.com/rednafi/examples/endpoints/greet"
)

type validator interface{ Validate() error }

// Wrap turns a domain function into an HTTP handler. The decode and encode
// callbacks are the only place HTTP-specific types appear at the call site.
func Wrap[In, Out any](
	decode func(*nethttp.Request) (In, error),
	fn func(context.Context, In) (Out, error),
	encode func(nethttp.ResponseWriter, Out) error,
) nethttp.Handler {
	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
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

func decodeGreet(r *nethttp.Request) (greet.GreetIn, error) {
	var body struct {
		Name      string `json:"name"`
		Formality int    `json:"formality"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return greet.GreetIn{}, greet.Invalid("malformed json")
	}
	return greet.GreetIn{Name: body.Name, Formality: body.Formality}, nil
}

func encodeGreet(w nethttp.ResponseWriter, out greet.GreetOut) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(nethttp.StatusOK)
	return json.NewEncoder(w).Encode(struct {
		Message string `json:"message"`
	}{out.Message})
}

func decodeSubscribe(r *nethttp.Request) (greet.SubscribeIn, error) {
	var body struct {
		Email     string `json:"email"`
		Formality int    `json:"formality"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return greet.SubscribeIn{}, greet.Invalid("malformed json")
	}
	return greet.SubscribeIn{Email: body.Email, Formality: body.Formality}, nil
}

func encodeSubscribe(w nethttp.ResponseWriter, out greet.SubscribeOut) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(nethttp.StatusCreated)
	return json.NewEncoder(w).Encode(struct {
		ID string `json:"id"`
	}{out.ID})
}

// Register mounts every greet endpoint on mux.
func Register(mux *nethttp.ServeMux, svc *greet.Service) {
	mux.Handle("POST /greet", Wrap(decodeGreet, svc.Greet, encodeGreet))
	mux.Handle("POST /subscribe", Wrap(decodeSubscribe, svc.Subscribe, encodeSubscribe))
}

func writeErr(w nethttp.ResponseWriter, err error) {
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
		return nethttp.StatusBadRequest
	case greet.CodeUnauthenticated:
		return nethttp.StatusUnauthorized
	case greet.CodePermissionDenied:
		return nethttp.StatusForbidden
	case greet.CodeNotFound:
		return nethttp.StatusNotFound
	case greet.CodeAlreadyExists:
		return nethttp.StatusConflict
	default:
		return nethttp.StatusInternalServerError
	}
}
