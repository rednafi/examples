// Package coalesce shows request coalescing with golang.org/x/sync/singleflight:
// many concurrent callers asking for the same key make the work happen once.
package coalesce

import (
	"context"

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

// LoadChan is Load with per-caller cancellation. The shared fetch runs on a
// context detached from any single caller, so one caller giving up does not
// cancel the work the others are still waiting on.
func (l *Loader) LoadChan(ctx context.Context, key string) (string, error) {
	ch := l.group.DoChan(key, func() (any, error) {
		return l.fetch(context.WithoutCancel(ctx), key) // (2)
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
