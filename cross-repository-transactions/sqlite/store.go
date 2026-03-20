package sqlite

import (
	"context"
	"database/sql"

	"github.com/rednafi/examples/cross-repository-transactions/bookstore"

	_ "github.com/mattn/go-sqlite3"
)

type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// BookStore implements bookstore.BookStore against SQLite.
type BookStore struct{ db DBTX }

func NewBookStore(db DBTX) *BookStore { return &BookStore{db: db} }

func (s *BookStore) Get(ctx context.Context, id int64) (bookstore.Book, error) {
	row := s.db.QueryRowContext(ctx,
		"SELECT id, title, stock FROM books WHERE id = ?", id)
	var b bookstore.Book
	err := row.Scan(&b.ID, &b.Title, &b.Stock)
	return b, err
}

func (s *BookStore) Create(ctx context.Context, b bookstore.Book) (int64, error) {
	res, err := s.db.ExecContext(ctx,
		"INSERT INTO books (title, stock) VALUES (?, ?)", b.Title, b.Stock)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *BookStore) DecrementStock(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx,
		"UPDATE books SET stock = stock - 1 WHERE id = ? AND stock > 0", id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// OrderStore implements bookstore.OrderStore against SQLite.
type OrderStore struct{ db DBTX }

func NewOrderStore(db DBTX) *OrderStore { return &OrderStore{db: db} }

func (s *OrderStore) Create(ctx context.Context, o bookstore.Order) (int64, error) {
	res, err := s.db.ExecContext(ctx,
		"INSERT INTO orders (book_id) VALUES (?)", o.BookID)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *OrderStore) Get(ctx context.Context, id int64) (bookstore.Order, error) {
	row := s.db.QueryRowContext(ctx,
		"SELECT id, book_id FROM orders WHERE id = ?", id)
	var o bookstore.Order
	err := row.Scan(&o.ID, &o.BookID)
	return o, err
}

// UoW implements bookstore.UnitOfWork using a real SQL transaction.
type UoW struct{ db *sql.DB }

func NewUoW(db *sql.DB) *UoW { return &UoW{db: db} }

func (u *UoW) RunInTx(ctx context.Context, fn func(bookstore.Stores) error) error {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	stores := bookstore.Stores{
		Books:  NewBookStore(tx),
		Orders: NewOrderStore(tx),
	}

	if err := fn(stores); err != nil {
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
			title TEXT NOT NULL,
			stock INTEGER NOT NULL DEFAULT 0
		);
		CREATE TABLE IF NOT EXISTS orders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			book_id INTEGER NOT NULL
		);
	`)
	if err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
