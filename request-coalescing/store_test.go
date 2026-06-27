package coalesce

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// A cold cache hit by many readers fetches once; a warm cache serves the rest.
func TestStoreCacheAside(t *testing.T) {
	var calls atomic.Int64
	start := make(chan struct{})
	s := NewStore(func(ctx context.Context, key string) (string, error) {
		calls.Add(1)
		<-start // hold the upstream call open so the readers pile up
		return "value:" + key, nil
	})

	const n = 100
	var wg sync.WaitGroup
	for range n {
		wg.Go(func() {
			v, err := s.Get(context.Background(), "k")
			if err != nil || v != "value:k" {
				t.Errorf("Get = %q, %v", v, err)
			}
		})
	}

	time.Sleep(10 * time.Millisecond)
	close(start)
	wg.Wait()

	if got := calls.Load(); got != 1 {
		t.Fatalf("cold miss: upstream ran %d times, want 1", got)
	}

	for range 50 { // warm cache: further reads never reach the upstream
		if _, err := s.Get(context.Background(), "k"); err != nil {
			t.Fatal(err)
		}
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("warm cache: upstream ran %d times total, want 1", got)
	}
}

func TestStoreSharedMetricsCountRecipients(t *testing.T) {
	var calls atomic.Int64
	var sharedOK atomic.Int64
	var sharedErr atomic.Int64
	start := make(chan struct{})
	s := NewStore(func(ctx context.Context, key string) (string, error) {
		calls.Add(1)
		<-start // hold the upstream call open so every caller joins the group
		return "value:" + key, nil
	})

	const n = 4
	var wg sync.WaitGroup
	for range n {
		wg.Go(func() {
			v, err := s.GetWithMetrics(context.Background(), "k", &sharedOK, &sharedErr)
			if err != nil || v != "value:k" {
				t.Errorf("GetWithMetrics = %q, %v", v, err)
			}
		})
	}

	time.Sleep(10 * time.Millisecond)
	close(start)
	wg.Wait()

	if got := calls.Load(); got != 1 {
		t.Fatalf("upstream ran %d times, want 1", got)
	}
	if got := sharedOK.Load(); got != n {
		t.Fatalf("sharedOK = %d, want %d", got, n)
	}
	if got := sharedErr.Load(); got != 0 {
		t.Fatalf("sharedErr = %d, want 0", got)
	}
}

func TestStoreSharedErrorMetricsCountRecipients(t *testing.T) {
	var calls atomic.Int64
	var sharedOK atomic.Int64
	var sharedErr atomic.Int64
	start := make(chan struct{})
	wantErr := errors.New("upstream failed")
	s := NewStore(func(ctx context.Context, key string) (string, error) {
		calls.Add(1)
		<-start // hold the upstream call open so every caller joins the group
		return "", wantErr
	})

	const n = 4
	var wg sync.WaitGroup
	for range n {
		wg.Go(func() {
			if _, err := s.GetWithMetrics(context.Background(), "k", &sharedOK, &sharedErr); !errors.Is(err, wantErr) {
				t.Errorf("GetWithMetrics err = %v, want %v", err, wantErr)
			}
		})
	}

	time.Sleep(10 * time.Millisecond)
	close(start)
	wg.Wait()

	if got := calls.Load(); got != 1 {
		t.Fatalf("upstream ran %d times, want 1", got)
	}
	if got := sharedOK.Load(); got != 0 {
		t.Fatalf("sharedOK = %d, want 0", got)
	}
	if got := sharedErr.Load(); got != n {
		t.Fatalf("sharedErr = %d, want %d", got, n)
	}
}
