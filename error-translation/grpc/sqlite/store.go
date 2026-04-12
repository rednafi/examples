package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/mattn/go-sqlite3"
	"github.com/rednafi/examples/error-translation/grpc/user"
)

type UserStore struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db: db}
}

func (s *UserStore) Get(
	ctx context.Context, id int64,
) (user.User, error) {
	row := s.db.QueryRowContext(ctx,
		"SELECT id, name, email FROM users WHERE id = ?", id)

	var u user.User
	if err := row.Scan(&u.ID, &u.Name, &u.Email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return user.User{}, fmt.Errorf(
				"user %d not in db: %w", id, user.ErrNotFound,
			)
		}
		return user.User{}, fmt.Errorf(
			"querying user %d: %v", id, err,
		)
	}
	return u, nil
}

func (s *UserStore) Create(
	ctx context.Context, u user.User,
) (int64, error) {
	res, err := s.db.ExecContext(ctx,
		"INSERT INTO users (name, email) VALUES (?, ?)",
		u.Name, u.Email,
	)
	if err != nil {
		if sqliteErr, ok := errors.AsType[sqlite3.Error](err); ok &&
			sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf(
				"user %s already exists: %w", u.Email, user.ErrConflict,
			)
		}
		return 0, fmt.Errorf("inserting user: %v", err)
	}
	return res.LastInsertId()
}
