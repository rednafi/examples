package user

import (
	"context"
	"fmt"
)

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
