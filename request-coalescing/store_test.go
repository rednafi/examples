package coalesce

import (
	"context"
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
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, err := s.Get(context.Background(), "k")
			if err != nil || v != "value:k" {
				t.Errorf("Get = %q, %v", v, err)
			}
		}()
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
