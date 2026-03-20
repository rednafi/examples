package bookstore

import "context"

type Service struct {
	store BookStore
}

func NewService(s BookStore) *Service {
	return &Service{store: s}
}

func (s *Service) GetBook(ctx context.Context, id int64) (Book, error) {
	return s.store.Get(ctx, id)
}

// RegisterBook creates a book and writes an audit log in a single transaction.
func (s *Service) RegisterBook(ctx context.Context, title string) (Book, error) {
	var book Book

	err := s.store.Tx(ctx, func(tx BookStore) error {
		id, err := tx.Create(ctx, Book{Title: title})
		if err != nil {
			return err
		}
		book = Book{ID: id, Title: title}
		return tx.CreateAuditLog(ctx, AuditEntry{BookID: id, Action: "created"})
	})

	return book, err
}
