// Package codegen holds the structs the generator reads and the
// generated Validate methods it emits.
package codegen

//go:generate go run ./gen

type User struct {
	Name  string `check:"required,min=2"`
	Email string `check:"required,email"`
}

type Signup struct {
	Username string `check:"required,min=3"`
	Email    string `check:"required,email"`
}
