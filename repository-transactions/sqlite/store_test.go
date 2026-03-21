package sqlite

import (
	"context"
	"database/sql"
	"testing"

	"github.com/rednafi/examples/repository-transactions/bookstore"
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

func TestTx_CommitsOnSuccess(t *testing.T) {
	db := setupTestDB(t)
	store := NewStore(db)
	svc := bookstore.NewService(store)

	b, err := svc.RegisterBook(t.Context(), "DDIA")
	if err != nil {
		t.Fatal(err)
	}
	if b.ID == 0 {
		t.Fatal("expected non-zero ID")
	}

	got, err := store.Get(t.Context(), b.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "DDIA" {
		t.Fatalf("got title %q, want DDIA", got.Title)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM audit_log WHERE book_id = ?",
		b.ID).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("expected 1 audit entry, got %d", count)
	}
}

func TestTx_RollsBackOnError(t *testing.T) {
	db := setupTestDB(t)

	base := NewStore(db)
	failing := &failingStore{Store: base}

	svc := bookstore.NewService(failing)

	_, err := svc.RegisterBook(t.Context(), "DDIA")
	if err == nil {
		t.Fatal("expected error")
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM books").Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("expected 0 books after rollback, got %d", count)
	}
}

type failingStore struct {
	*Store
}

func (f *failingStore) CreateAuditLog(
	ctx context.Context, e bookstore.AuditEntry) error {
	return sql.ErrConnDone
}

func (f *failingStore) Tx(
	ctx context.Context, fn func(bookstore.BookStore) error) error {
	sqlDB, ok := f.db.(*sql.DB)
	if !ok {
		return nil
	}
	tx, err := sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	txStore := &failingStore{Store: NewStore(tx)}
	if err := fn(txStore); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
