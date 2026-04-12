package user

import (
	"context"
	"errors"
	"time"
)

type User struct {
	ID        int64
	Name      string
	Email     string
	DeletedAt *time.Time
}

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
)

type Store interface {
	Get(ctx context.Context, id int64) (User, error)
	Create(ctx context.Context, u User) (int64, error)
}
