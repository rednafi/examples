package coalesce

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Many concurrent Loads for the same key trigger exactly one fetch.
func TestLoadCoalesces(t *testing.T) {
	var calls atomic.Int64
	start := make(chan struct{})
	l := NewLoader(func(ctx context.Context, key string) (string, error) {
		calls.Add(1)
		<-start // hold the fetch open so every caller piles up behind it
		return "value:" + key, nil
	})

	const n = 100
	var wg sync.WaitGroup
	got := make([]string, n)
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, err := l.Load(context.Background(), "k")
			if err != nil {
				t.Errorf("Load: %v", err)
			}
			got[i] = v
		}()
	}

	time.Sleep(10 * time.Millisecond) // let the callers queue, then release the one fetch
	close(start)
	wg.Wait()

	if n := calls.Load(); n != 1 {
		t.Fatalf("fetch ran %d times, want 1", n)
	}
	for i, v := range got {
		if v != "value:k" {
			t.Fatalf("caller %d got %q, want value:k", i, v)
		}
	}
}

// A caller that cancels gets its own context error; the shared fetch keeps running.
func TestLoadChanPerCallerCancel(t *testing.T) {
	release := make(chan struct{})
	l := NewLoader(func(ctx context.Context, key string) (string, error) {
		<-release
		return "v", nil
	})

	ctxA, cancelA := context.WithCancel(context.Background())
	errA := make(chan error, 1)
	go func() {
		_, err := l.LoadChan(ctxA, "k")
		errA <- err
	}()

	time.Sleep(10 * time.Millisecond) // caller A is now waiting on the fetch
	cancelA()

	if err := <-errA; err != context.Canceled {
		t.Fatalf("caller A err = %v, want context.Canceled", err)
	}

	close(release) // shared fetch finishes; a fresh caller still gets the value
	v, err := l.LoadChan(context.Background(), "k")
	if err != nil {
		t.Fatalf("caller B: %v", err)
	}
	if v != "v" {
		t.Fatalf("caller B got %q, want v", v)
	}
}
