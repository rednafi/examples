package order

import "context"

type Order struct {
	ID     int64
	BookID int64
}

type Store interface {
	Create(ctx context.Context, o Order) (int64, error)
	Get(ctx context.Context, id int64) (Order, error)
}
