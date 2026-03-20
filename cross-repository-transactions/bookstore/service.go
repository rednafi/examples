package bookstore

import "context"

type Service struct {
	stores Stores
	uow    UnitOfWork
}

func NewService(s Stores, uow UnitOfWork) *Service {
	return &Service{stores: s, uow: uow}
}

func (s *Service) GetBook(ctx context.Context, id int64) (Book, error) {
	return s.stores.Books.Get(ctx, id)
}

func (s *Service) GetOrder(ctx context.Context, id int64) (Order, error) {
	return s.stores.Orders.Get(ctx, id)
}

// PlaceOrder decrements stock and creates an order atomically.
func (s *Service) PlaceOrder(ctx context.Context, bookID int64) (Order, error) {
	book, err := s.stores.Books.Get(ctx, bookID)
	if err != nil {
		return Order{}, err
	}

	var order Order
	err = s.uow.RunInTx(ctx, func(tx Stores) error {
		if err := tx.Books.DecrementStock(ctx, book.ID); err != nil {
			return err
		}
		id, err := tx.Orders.Create(ctx, Order{BookID: book.ID})
		if err != nil {
			return err
		}
		order = Order{ID: id, BookID: book.ID}
		return nil
	})

	return order, err
}
