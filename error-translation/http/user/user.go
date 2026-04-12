package user

import (
	"context"
	"errors"
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
