package bookstore

import "context"

type Book struct {
	ID    int64
	Title string
}

type AuditEntry struct {
	BookID int64
	Action string
}

type BookStore interface {
	Get(ctx context.Context, id int64) (Book, error)
	Create(ctx context.Context, b Book) (int64, error)
	CreateAuditLog(ctx context.Context, e AuditEntry) error

	// Tx runs fn inside a transaction. The BookStore passed to fn
	// executes against that transaction.
	Tx(ctx context.Context, fn func(BookStore) error) error
}
