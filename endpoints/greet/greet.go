// Package greet is the domain. Types, validation rules, the service, and
// the error vocabulary all live here. Nothing in this package imports a
// transport (net/http, google.golang.org/grpc) or a wire format (encoding/json,
// google.golang.org/protobuf). That is the test of whether the layering is
// right; grep this directory for those imports and the result should be empty.
package greet

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// GreetIn is the input to Greet. It is a plain Go struct with no
// JSON tags or protobuf annotations.
type GreetIn struct {
	Name      string
	Formality int
}

func (in GreetIn) Validate() error {
	if in.Name == "" {
		return Invalid("name is required")
	}
	if in.Formality < 0 || in.Formality > 1 {
		return Invalid("formality must be 0 or 1")
	}
	return nil
}

type GreetOut struct {
	Message string
}

type SubscribeIn struct {
	Email     string
	Formality int
}

func (in SubscribeIn) Validate() error {
	if !strings.Contains(in.Email, "@") {
		return Invalid("email must contain @")
	}
	if in.Formality < 0 || in.Formality > 1 {
		return Invalid("formality must be 0 or 1")
	}
	return nil
}

type SubscribeOut struct {
	ID string
}

type Service struct {
	mu   sync.Mutex
	subs map[string]string
}

func NewService() *Service {
	return &Service{subs: make(map[string]string)}
}

func (s *Service) Greet(_ context.Context, in GreetIn) (GreetOut, error) {
	if in.Formality == 1 {
		return GreetOut{Message: "Good day, " + in.Name + "."}, nil
	}
	return GreetOut{Message: "hey " + in.Name + "!"}, nil
}

func (s *Service) Subscribe(_ context.Context, in SubscribeIn) (SubscribeOut, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.subs[in.Email]; ok {
		return SubscribeOut{}, AlreadyExists("already subscribed")
	}
	id := fmt.Sprintf("sub_%d", len(s.subs)+1)
	s.subs[in.Email] = id
	return SubscribeOut{ID: id}, nil
}
