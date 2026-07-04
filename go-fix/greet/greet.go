// Package greet renders greetings. It stands in for any library whose API you
// want to evolve without breaking the callers who depend on it.
package greet

import "fmt"

// Greet returns a greeting for name. This is the API we want callers to use.
func Greet(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}

// Hello returns a greeting for name.
//
// Deprecated: use [Greet] instead.
//
//go:fix inline
func Hello(name string) string {
	return Greet(name)
}
