// Package greet holds the domain types and the business functions.
//
// Nothing here imports net/http, google.golang.org/grpc, encoding/json,
// or anything from api. That is the test of whether the layering is right.
package greet

import (
	"context"
	"strings"

	"github.com/rednafi/examples/endpoints/errs"
)

type GreetIn struct {
	Name      string
	Formality int
}

func (in GreetIn) Validate() error {
	if in.Name == "" {
		return errs.Invalid("name is required")
	}
	if in.Formality < 0 || in.Formality > 1 {
		return errs.Invalid("formality must be 0 or 1")
	}
	return nil
}

type GreetOut struct {
	Message string
}

type CreateUserIn struct {
	Email string
	Age   int
}

func (in CreateUserIn) Validate() error {
	if !strings.Contains(in.Email, "@") {
		return errs.Invalid("email must contain @")
	}
	if in.Age < 0 {
		return errs.Invalid("age must be >= 0")
	}
	if in.Age > 150 {
		return errs.Invalid("age must be <= 150")
	}
	return nil
}

type CreateUserOut struct {
	ID string
}

type Service struct {
	users map[string]string
}

func NewService() *Service {
	return &Service{users: make(map[string]string)}
}

func (s *Service) Greet(_ context.Context, in GreetIn) (GreetOut, error) {
	if in.Formality == 1 {
		return GreetOut{Message: "Good day, " + in.Name + "."}, nil
	}
	return GreetOut{Message: "hey " + in.Name + "!"}, nil
}

func (s *Service) CreateUser(_ context.Context, in CreateUserIn) (CreateUserOut, error) {
	if _, ok := s.users[in.Email]; ok {
		return CreateUserOut{}, errs.AlreadyExists("user already exists")
	}
	id := "u_" + in.Email
	s.users[in.Email] = id
	return CreateUserOut{ID: id}, nil
}
