package greet

import (
	"context"
	"errors"
	"sync"
)

// Callers translate ErrNotFound to a domain *greet.Error; raw store errors
// must not leak past greet.Service.
var ErrNotFound = errors.New("user not found")

type User struct {
	ID   int64
	Name string
}

type UserStore interface {
	GetUser(ctx context.Context, id int64) (User, error)
}

type MemoryStore struct {
	mu    sync.RWMutex
	users map[int64]User
}

func NewMemoryStore(users ...User) *MemoryStore {
	s := &MemoryStore{users: make(map[int64]User, len(users))}
	for _, u := range users {
		s.users[u.ID] = u
	}
	return s
}

func (s *MemoryStore) GetUser(_ context.Context, id int64) (User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[id]
	if !ok {
		return User{}, ErrNotFound
	}
	return u, nil
}
