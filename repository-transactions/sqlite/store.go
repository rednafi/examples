package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/rednafi/examples/repository-transactions/book"

	_ "github.com/mattn/go-sqlite3"
)

type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type Store struct{ db DBTX }

func NewStore(db DBTX) *Store { return &Store{db: db} }

func (s *Store) Get(ctx context.Context, id int64) (book.Book, error) {
	row := s.db.QueryRowContext(ctx,
		"SELECT id, title FROM books WHERE id = ?", id)
	var b book.Book
	err := row.Scan(&b.ID, &b.Title)
	return b, err
}

func (s *Store) Create(ctx context.Context, b book.Book) (int64, error) {
	res, err := s.db.ExecContext(ctx,
		"INSERT INTO books (title) VALUES (?)", b.Title)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) CreateAuditLog(ctx context.Context, e book.AuditEntry) error {
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO audit_log (book_id, action) VALUES (?, ?)",
		e.BookID, e.Action)
	return err
}

func (s *Store) Tx(ctx context.Context, fn func(book.Store) error) error {
	sqlDB, ok := s.db.(*sql.DB)
	if !ok {
		return errors.New("cannot start tx: already inside a transaction")
	}

	tx, err := sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(NewStore(tx)); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func SetupDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS books (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS audit_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			book_id INTEGER NOT NULL,
			action TEXT NOT NULL
		);
	`)
	if err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
