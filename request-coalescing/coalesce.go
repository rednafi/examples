// Package coalesce shows request coalescing with golang.org/x/sync/singleflight:
// many concurrent callers asking for the same key make the work happen once.
package coalesce

import (
	"context"
	"time"

	"golang.org/x/sync/singleflight"
)

// Loader fetches values by key and collapses concurrent requests for the same
// key into a single call to fetch.
type Loader struct {
	group singleflight.Group
	fetch func(ctx context.Context, key string) (string, error)
}

func NewLoader(fetch func(ctx context.Context, key string) (string, error)) *Loader {
	return &Loader{fetch: fetch}
}

// Load coalesces concurrent calls for the same key. Callers that arrive while a
// fetch is in flight wait for it and receive the same result.
func (l *Loader) Load(ctx context.Context, key string) (string, error) {
	v, err, _ := l.group.Do(key, func() (any, error) { // (1)
		return l.fetch(ctx, key)
	})
	if err != nil {
		return "", err
	}
	return v.(string), nil
}

// LoadChan is Load with per-caller cancellation. The shared fetch gets its own
// timeout and is detached from any single caller, so one caller giving up does
// not cancel the work the others are still waiting on.
func (l *Loader) LoadChan(ctx context.Context, key string, fetchTimeout time.Duration) (string, error) {
	ch := l.group.DoChan(key, func() (any, error) {
		detached := context.WithoutCancel(ctx) // (2)
		callCtx, cancel := context.WithTimeout(detached, fetchTimeout)
		defer cancel()
		return l.fetch(callCtx, key)
	})
	select {
	case <-ctx.Done(): // (3)
		return "", ctx.Err()
	case res := <-ch:
		if res.Err != nil {
			return "", res.Err
		}
		return res.Val.(string), nil
	}
}

// LoadMaxWait starts a fresh call for the next caller if this caller waits too
// long. Forget does not cancel the in-flight fetch; fetchTimeout still bounds it.
func (l *Loader) LoadMaxWait(ctx context.Context, key string, fetchTimeout, maxWait time.Duration) (string, error) {
	ch := l.group.DoChan(key, func() (any, error) {
		detached := context.WithoutCancel(ctx)
		callCtx, cancel := context.WithTimeout(detached, fetchTimeout)
		defer cancel()
		return l.fetch(callCtx, key)
	})
	select {
	case res := <-ch:
		if res.Err != nil {
			return "", res.Err
		}
		return res.Val.(string), nil
	case <-time.After(maxWait):
		l.group.Forget(key)
		return "", context.DeadlineExceeded
	}
}
