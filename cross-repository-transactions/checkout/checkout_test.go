package checkout

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/rednafi/examples/cross-repository-transactions/bookstore"
	"github.com/rednafi/examples/cross-repository-transactions/orderstore"
)

var _ bookstore.BookStore = (*memBookStore)(nil)
var _ orderstore.OrderStore = (*memOrderStore)(nil)
var _ UnitOfWork = (*memUoW)(nil)

type memBookStore struct {
	mu    sync.Mutex
	books map[int64]bookstore.Book
	next  int64
}

func newMemBookStore() *memBookStore {
	return &memBookStore{books: make(map[int64]bookstore.Book)}
}

func (m *memBookStore) Get(_ context.Context, id int64) (bookstore.Book, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	b, ok := m.books[id]
	if !ok {
		return bookstore.Book{}, fmt.Errorf("book %d not found", id)
	}
	return b, nil
}

func (m *memBookStore) Create(_ context.Context, b bookstore.Book) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.next++
	b.ID = m.next
	m.books[b.ID] = b
	return b.ID, nil
}

func (m *memBookStore) CreateAuditLog(_ context.Context, _ bookstore.AuditEntry) error {
	return nil
}

func (m *memBookStore) DecrementStock(_ context.Context, id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	b, ok := m.books[id]
	if !ok {
		return fmt.Errorf("book %d not found", id)
	}
	if b.Stock <= 0 {
		return fmt.Errorf("book %d out of stock", id)
	}
	b.Stock--
	m.books[id] = b
	return nil
}

type memOrderStore struct {
	mu     sync.Mutex
	orders map[int64]orderstore.Order
	next   int64
}

func newMemOrderStore() *memOrderStore {
	return &memOrderStore{orders: make(map[int64]orderstore.Order)}
}

func (m *memOrderStore) Create(_ context.Context, o orderstore.Order) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.next++
	o.ID = m.next
	m.orders[o.ID] = o
	return o.ID, nil
}

func (m *memOrderStore) Get(_ context.Context, id int64) (orderstore.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	o, ok := m.orders[id]
	if !ok {
		return orderstore.Order{}, fmt.Errorf("order %d not found", id)
	}
	return o, nil
}

type memUoW struct {
	stores Stores
}

func (m *memUoW) RunInTx(_ context.Context, fn func(Stores) error) error {
	return fn(m.stores)
}

func TestPlaceOrder(t *testing.T) {
	books := newMemBookStore()
	orders := newMemOrderStore()

	books.books[1] = bookstore.Book{ID: 1, Title: "DDIA", Stock: 5}

	stores := Stores{Books: books, Orders: orders}
	svc := NewService(stores, &memUoW{stores: stores})

	order, err := svc.PlaceOrder(t.Context(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if order.BookID != 1 {
		t.Fatalf("order book ID = %d, want 1", order.BookID)
	}

	books.mu.Lock()
	got := books.books[1]
	books.mu.Unlock()
	if got.Stock != 4 {
		t.Fatalf("stock = %d, want 4", got.Stock)
	}
}

func TestPlaceOrder_OutOfStock(t *testing.T) {
	books := newMemBookStore()
	orders := newMemOrderStore()

	books.books[1] = bookstore.Book{ID: 1, Title: "DDIA", Stock: 0}

	stores := Stores{Books: books, Orders: orders}
	svc := NewService(stores, &memUoW{stores: stores})

	_, err := svc.PlaceOrder(t.Context(), 1)
	if err == nil {
		t.Fatal("expected error for out-of-stock book")
	}

	orders.mu.Lock()
	count := len(orders.orders)
	orders.mu.Unlock()
	if count != 0 {
		t.Fatalf("expected 0 orders after failure, got %d", count)
	}
}
