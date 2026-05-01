// Package httpapi has the HTTP wiring: a decode and encode function per
// endpoint, plus a single RegisterRoutes call that mounts every endpoint
// onto an http.ServeMux through transport.WrapHTTP.
package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/rednafi/examples/endpoints/errs"
	"github.com/rednafi/examples/endpoints/greet"
	"github.com/rednafi/examples/endpoints/transport"
)

func decodeGreet(r *http.Request) (greet.GreetIn, error) {
	var body struct {
		Name      string `json:"name"`
		Formality int    `json:"formality"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return greet.GreetIn{}, errs.Invalid("malformed json")
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

func decodeCreateUser(r *http.Request) (greet.CreateUserIn, error) {
	var body struct {
		Email string `json:"email"`
		Age   int    `json:"age"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return greet.CreateUserIn{}, errs.Invalid("malformed json")
	}
	return greet.CreateUserIn{Email: body.Email, Age: body.Age}, nil
}

func encodeCreateUser(w http.ResponseWriter, out greet.CreateUserOut) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	return json.NewEncoder(w).Encode(struct {
		ID string `json:"id"`
	}{out.ID})
}

func RegisterRoutes(mux *http.ServeMux, svc *greet.Service) {
	mux.Handle("POST /greet", transport.WrapHTTP(decodeGreet, svc.Greet, encodeGreet))
	mux.Handle("POST /users", transport.WrapHTTP(decodeCreateUser, svc.CreateUser, encodeCreateUser))
}
