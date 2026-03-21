package checkout

import (
	"context"

	"github.com/rednafi/examples/cross-repository-transactions/book"
	"github.com/rednafi/examples/cross-repository-transactions/order"
)

// Stores groups every repository the service layer needs. It lives here
// because checkout is the package that coordinates across both stores.
// Neither book nor order imports the other.
type Stores struct {
	Books  book.Store
	Orders order.Store
}

// UnitOfWork runs fn inside a single transaction. Every store in the
// Stores value passed to fn executes against that transaction.
type UnitOfWork interface {
	RunInTx(ctx context.Context, fn func(Stores) error) error
}

type Service struct {
	stores Stores
	uow    UnitOfWork
}

func NewService(s Stores, uow UnitOfWork) *Service {
	return &Service{stores: s, uow: uow}
}

func (s *Service) GetBook(
	ctx context.Context, id int64) (book.Book, error) {
	return s.stores.Books.Get(ctx, id)
}

func (s *Service) GetOrder(
	ctx context.Context, id int64) (order.Order, error) {
	return s.stores.Orders.Get(ctx, id)
}

// PlaceOrder decrements stock and creates an order atomically.
func (s *Service) PlaceOrder(
	ctx context.Context, bookID int64) (order.Order, error) {

	book, err := s.stores.Books.Get(ctx, bookID)
	if err != nil {
		return order.Order{}, err
	}

	var ord order.Order
	err = s.uow.RunInTx(ctx, func(tx Stores) error {
		if err := tx.Books.DecrementStock(ctx, book.ID); err != nil {
			return err
		}
		id, err := tx.Orders.Create(ctx, order.Order{BookID: book.ID})
		if err != nil {
			return err
		}
		ord = order.Order{ID: id, BookID: book.ID}
		return nil
	})

	return ord, err
}
