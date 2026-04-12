package user

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type User struct {
	ID        int64      `json:"id"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
)

type Store interface {
	Get(ctx context.Context, id int64) (User, error)
	Create(ctx context.Context, u User) (int64, error)
}

type Service struct {
	store Store
}

func NewService(s Store) *Service {
	return &Service{store: s}
}

func (s *Service) GetUser(ctx context.Context, id int64) (User, error) {
	u, err := s.store.Get(ctx, id)
	if err != nil {
		return User{}, err
	}
	if u.DeletedAt != nil {
		return User{}, fmt.Errorf("user %d soft-deleted: %w", id, ErrNotFound)
	}
	return u, nil
}

func (s *Service) CreateUser(
	ctx context.Context, name, email string,
) (User, error) {
	id, err := s.store.Create(ctx, User{Name: name, Email: email})
	if err != nil {
		return User{}, err
	}
	return User{ID: id, Name: name, Email: email}, nil
}
