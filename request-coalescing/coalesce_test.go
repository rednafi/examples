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
		wg.Go(func() {
			v, err := l.Load(context.Background(), "k")
			if err != nil {
				t.Errorf("Load: %v", err)
			}
			got[i] = v
		})
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
	started := make(chan struct{})
	var once sync.Once
	l := NewLoader(func(ctx context.Context, key string) (string, error) {
		once.Do(func() { close(started) })
		select {
		case <-release:
			return "v", nil
		case <-ctx.Done():
			return "", ctx.Err()
		}
	})

	ctxA, cancelA := context.WithCancel(context.Background())
	errA := make(chan error, 1)
	go func() {
		_, err := l.LoadChan(ctxA, "k", time.Second)
		errA <- err
	}()

	<-started
	gotB := make(chan string, 1)
	errB := make(chan error, 1)
	go func() {
		v, err := l.LoadChan(context.Background(), "k", time.Second)
		gotB <- v
		errB <- err
	}()

	time.Sleep(10 * time.Millisecond) // caller B is now waiting on the same fetch
	cancelA()

	if err := <-errA; err != context.Canceled {
		t.Fatalf("caller A err = %v, want context.Canceled", err)
	}

	close(release) // shared fetch keeps running for caller B
	if err := <-errB; err != nil {
		t.Fatalf("caller B err = %v", err)
	}
	if v := <-gotB; v != "v" {
		t.Fatalf("caller B got %q, want v", v)
	}
}

func TestLoadChanSharedFetchTimeout(t *testing.T) {
	l := NewLoader(func(ctx context.Context, key string) (string, error) {
		<-ctx.Done()
		return "", ctx.Err()
	})

	_, err := l.LoadChan(context.Background(), "k", 10*time.Millisecond)
	if err != context.DeadlineExceeded {
		t.Fatalf("LoadChan err = %v, want context.DeadlineExceeded", err)
	}
}

func TestLoadMaxWaitForgetsKey(t *testing.T) {
	releaseFirst := make(chan struct{})
	started := make(chan struct{})
	var calls atomic.Int64
	l := NewLoader(func(ctx context.Context, key string) (string, error) {
		call := calls.Add(1)
		if call == 1 {
			close(started)
			select {
			case <-releaseFirst:
				return "first:" + key, nil
			case <-ctx.Done():
				return "", ctx.Err()
			}
		}
		return "second:" + key, nil
	})

	errA := make(chan error, 1)
	go func() {
		_, err := l.LoadMaxWait(context.Background(), "k", time.Second, 10*time.Millisecond)
		errA <- err
	}()
	<-started
	if err := <-errA; err != context.DeadlineExceeded {
		t.Fatalf("first caller err = %v, want context.DeadlineExceeded", err)
	}

	v, err := l.LoadChan(context.Background(), "k", time.Second)
	if err != nil {
		t.Fatalf("fresh caller: %v", err)
	}
	if v != "second:k" {
		t.Fatalf("fresh caller got %q, want second:k", v)
	}
	close(releaseFirst)
	if got := calls.Load(); got != 2 {
		t.Fatalf("fetch ran %d times, want 2", got)
	}
}
