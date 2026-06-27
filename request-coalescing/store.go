package coalesce

import (
	"context"
	"sync"
	"sync/atomic"

	"golang.org/x/sync/singleflight"
)

// cache is a tiny concurrency-safe map standing in for a real cache.
type cache struct {
	mu sync.RWMutex
	m  map[string]string
}

func newCache() *cache { return &cache{m: make(map[string]string)} }

func (c *cache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.m[key]
	return v, ok
}

func (c *cache) Set(key, val string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[key] = val
}

// Store is a cache-aside reader: serve from the cache, and on a miss fetch the
// value once, coalescing concurrent misses, before caching it.
type Store struct {
	cache *cache
	group singleflight.Group
	// fetch is the upstream call made on a miss: an HTTP or gRPC request to
	// another service, or a database query.
	fetch func(ctx context.Context, key string) (string, error)
}

func NewStore(fetch func(ctx context.Context, key string) (string, error)) *Store {
	return &Store{cache: newCache(), fetch: fetch}
}

// Get returns the cached value, or fetches it once on a miss and caches it.
func (s *Store) Get(ctx context.Context, key string) (string, error) {
	if v, ok := s.cache.Get(key); ok { // (1)
		return v, nil
	}
	v, err, _ := s.group.Do(key, func() (any, error) { // (2)
		val, err := s.fetch(ctx, key)
		if err != nil {
			return "", err
		}
		s.cache.Set(key, val) // (3)
		return val, nil
	})
	if err != nil {
		return "", err
	}
	return v.(string), nil
}

// GetWithMetrics is Get with counters for coalesced successes and failures.
func (s *Store) GetWithMetrics(
	ctx context.Context,
	key string,
	sharedOK *atomic.Int64,
	sharedErr *atomic.Int64,
) (string, error) {
	if v, ok := s.cache.Get(key); ok {
		return v, nil
	}
	v, err, shared := s.group.Do(key, func() (any, error) {
		val, err := s.fetch(ctx, key)
		if err != nil {
			return "", err
		}
		s.cache.Set(key, val) // (1)
		return val, nil
	})
	if shared { // (2)
		if err != nil {
			sharedErr.Add(1) // (3)
		} else {
			sharedOK.Add(1) // (4)
		}
	}
	if err != nil {
		return "", err
	}
	return v.(string), nil
}
