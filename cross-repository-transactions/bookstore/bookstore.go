package bookstore

import "context"

type Book struct {
	ID    int64
	Title string
	Stock int
}

type Order struct {
	ID     int64
	BookID int64
}

type BookStore interface {
	Get(ctx context.Context, id int64) (Book, error)
	Create(ctx context.Context, b Book) (int64, error)
	DecrementStock(ctx context.Context, id int64) error
}

type OrderStore interface {
	Create(ctx context.Context, o Order) (int64, error)
	Get(ctx context.Context, id int64) (Order, error)
}

// Stores groups every repository the service layer needs.
type Stores struct {
	Books  BookStore
	Orders OrderStore
}

// UnitOfWork runs fn inside a single transaction. Every store in the
// Stores value passed to fn executes against that transaction.
type UnitOfWork interface {
	RunInTx(ctx context.Context, fn func(Stores) error) error
}
