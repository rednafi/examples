package book

import "context"

type Book struct {
	ID    int64
	Title string
	Stock int
}

type AuditEntry struct {
	BookID int64
	Action string
}

type Store interface {
	Get(ctx context.Context, id int64) (Book, error)
	Create(ctx context.Context, b Book) (int64, error)
	CreateAuditLog(ctx context.Context, e AuditEntry) error
	DecrementStock(ctx context.Context, id int64) error
}
