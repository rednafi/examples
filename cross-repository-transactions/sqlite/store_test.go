package sqlite

import (
	"context"
	"database/sql"
	"testing"

	"github.com/rednafi/examples/cross-repository-transactions/bookstore"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := SetupDB(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func seedBook(t *testing.T, db *sql.DB, title string, stock int) int64 {
	t.Helper()
	res, err := db.Exec(
		"INSERT INTO books (title, stock) VALUES (?, ?)", title, stock)
	if err != nil {
		t.Fatal(err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func TestRunInTx_CommitsOnSuccess(t *testing.T) {
	db := setupTestDB(t)
	bookID := seedBook(t, db, "DDIA", 5)

	stores := bookstore.Stores{
		Books:  NewBookStore(db),
		Orders: NewOrderStore(db),
	}
	uow := NewUoW(db)
	svc := bookstore.NewService(stores, uow)

	order, err := svc.PlaceOrder(context.Background(), bookID)
	if err != nil {
		t.Fatal(err)
	}
	if order.BookID != bookID {
		t.Fatalf("order book ID = %d, want %d", order.BookID, bookID)
	}

	// Verify stock was decremented.
	var stock int
	err = db.QueryRow("SELECT stock FROM books WHERE id = ?", bookID).Scan(&stock)
	if err != nil {
		t.Fatal(err)
	}
	if stock != 4 {
		t.Fatalf("stock = %d, want 4", stock)
	}

	// Verify order was created.
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM orders WHERE book_id = ?",
		bookID).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("expected 1 order, got %d", count)
	}
}

func TestRunInTx_RollsBackOnError(t *testing.T) {
	db := setupTestDB(t)
	bookID := seedBook(t, db, "DDIA", 5)

	// Use a failing order store so the transaction rolls back after
	// the stock decrement but before the order insert commits.
	stores := bookstore.Stores{
		Books:  NewBookStore(db),
		Orders: NewOrderStore(db),
	}
	failUoW := &failingOrderUoW{db: db}
	svc := bookstore.NewService(stores, failUoW)

	_, err := svc.PlaceOrder(context.Background(), bookID)
	if err == nil {
		t.Fatal("expected error")
	}

	// Stock should be unchanged because the transaction rolled back.
	var stock int
	err = db.QueryRow("SELECT stock FROM books WHERE id = ?", bookID).Scan(&stock)
	if err != nil {
		t.Fatal(err)
	}
	if stock != 5 {
		t.Fatalf("stock = %d, want 5 (rollback should have restored it)", stock)
	}

	// No orders should exist.
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM orders").Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("expected 0 orders after rollback, got %d", count)
	}
}

// failingOrderUoW is a UnitOfWork whose OrderStore always fails on Create.
type failingOrderUoW struct{ db *sql.DB }

func (u *failingOrderUoW) RunInTx(
	ctx context.Context, fn func(bookstore.Stores) error) error {

	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	stores := bookstore.Stores{
		Books:  NewBookStore(tx),
		Orders: &failingOrderStore{},
	}

	if err := fn(stores); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

type failingOrderStore struct{}

func (f *failingOrderStore) Create(
	_ context.Context, _ bookstore.Order) (int64, error) {
	return 0, sql.ErrConnDone
}

func (f *failingOrderStore) Get(
	_ context.Context, _ int64) (bookstore.Order, error) {
	return bookstore.Order{}, sql.ErrConnDone
}
