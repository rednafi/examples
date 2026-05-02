// Package greet is the domain. Types, validation rules, the service, and
// the error vocabulary all live here. Nothing in this package imports a
// transport (net/http, google.golang.org/grpc) or a wire format
// (encoding/json, google.golang.org/protobuf).
package greet

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
)

type GreetIn struct {
	UserID    int64
	Formality int
}

func (in GreetIn) Validate() error {
	if in.UserID == 0 {
		return Invalid("user_id is required")
	}
	if in.Formality < 0 || in.Formality > 1 {
		return Invalid("formality must be 0 or 1")
	}
	return nil
}

type GreetOut struct {
	Message string
}

type FarewellIn struct {
	UserID int64
}

func (in FarewellIn) Validate() error {
	if in.UserID == 0 {
		return Invalid("user_id is required")
	}
	return nil
}

type FarewellOut struct {
	Message string
}

type Service struct {
	users  UserStore
	logger *slog.Logger
}

func NewService(users UserStore, logger *slog.Logger) *Service {
	return &Service{users: users, logger: logger}
}

func (s *Service) Greet(ctx context.Context, in GreetIn) (GreetOut, error) {
	u, err := s.users.GetUser(ctx, in.UserID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return GreetOut{}, NotFound("user %d", in.UserID)
		}
		return GreetOut{}, fmt.Errorf("getting user: %w", err)
	}

	s.logger.Info("greeted", "user_id", u.ID, "formality", in.Formality)

	msg := "hey " + u.Name + "!"
	if in.Formality == 1 {
		msg = "Good day, " + u.Name + "."
	}
	return GreetOut{Message: msg}, nil
}

func (s *Service) Farewell(ctx context.Context, in FarewellIn) (FarewellOut, error) {
	u, err := s.users.GetUser(ctx, in.UserID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return FarewellOut{}, NotFound("user %d", in.UserID)
		}
		return FarewellOut{}, fmt.Errorf("getting user: %w", err)
	}

	s.logger.Info("farewelled", "user_id", u.ID)

	return FarewellOut{Message: "bye " + u.Name + "!"}, nil
}
