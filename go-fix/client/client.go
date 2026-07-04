package client

import "time"

// FetchTimeout fetches url with an explicit timeout.
func FetchTimeout(url string, timeout time.Duration) ([]byte, error) {
	_ = timeout
	return []byte(url), nil
}

// Fetch fetches url with a 30-second timeout.
//
// Deprecated: use [FetchTimeout] to pick the timeout.
//
//go:fix inline
func Fetch(url string) ([]byte, error) {
	return FetchTimeout(url, 30*time.Second)
}

// Options configures the client.
type Options struct{ Retries int }

// Deprecated: use [Options].
//
//go:fix inline
type Config = Options
