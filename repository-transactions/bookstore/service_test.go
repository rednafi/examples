package bookstore

import (
	"context"
	"fmt"
	"sync"
	"testing"
)

// Compile-time check that memStore satisfies BookStore.
var _ BookStore = (*memStore)(nil)

type memStore struct {
	mu       sync.Mutex
	books    map[int64]Book
	auditLog []AuditEntry
	next     int64
}

func newMemStore() *memStore {
	return &memStore{books: make(map[int64]Book)}
}

func (m *memStore) Get(ctx context.Context, id int64) (Book, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	b, ok := m.books[id]
	if !ok {
		return Book{}, fmt.Errorf("book %d not found", id)
	}
	return b, nil
}

func (m *memStore) Create(ctx context.Context, b Book) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.next++
	b.ID = m.next
	m.books[b.ID] = b
	return b.ID, nil
}

func (m *memStore) CreateAuditLog(ctx context.Context, e AuditEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.auditLog = append(m.auditLog, e)
	return nil
}

func (m *memStore) Tx(ctx context.Context, fn func(BookStore) error) error {
	return fn(m)
}

func TestRegisterBook(t *testing.T) {
	store := newMemStore()
	svc := NewService(store)

	b, err := svc.RegisterBook(context.Background(), "DDIA")
	if err != nil {
		t.Fatal(err)
	}
	if b.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if b.Title != "DDIA" {
		t.Fatalf("got title %q, want DDIA", b.Title)
	}

	store.mu.Lock()
	defer store.mu.Unlock()
	if len(store.auditLog) != 1 {
		t.Fatalf("expected 1 audit entry, got %d", len(store.auditLog))
	}
	if store.auditLog[0].BookID != b.ID {
		t.Fatalf("audit entry book ID = %d, want %d",
			store.auditLog[0].BookID, b.ID)
	}
}

type failingStore struct {
	*memStore
}

func (f *failingStore) CreateAuditLog(
	ctx context.Context, e AuditEntry) error {
	return fmt.Errorf("disk on fire")
}

func (f *failingStore) Tx(
	ctx context.Context, fn func(BookStore) error) error {
	return fn(f)
}

func TestRegisterBook_AuditFails(t *testing.T) {
	inner := newMemStore()
	store := &failingStore{memStore: inner}
	svc := NewService(store)

	_, err := svc.RegisterBook(context.Background(), "DDIA")
	if err == nil {
		t.Fatal("expected error when audit log fails")
	}
}
