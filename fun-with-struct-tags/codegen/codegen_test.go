package codegen

import (
	"strings"
	"testing"
)

func TestUser_OK(t *testing.T) {
	if err := (&User{Name: "ada", Email: "a@b.com"}).Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUser_Required(t *testing.T) {
	err := (&User{Email: "a@b.com"}).Validate()
	if err == nil || !strings.Contains(err.Error(), "required") {
		t.Fatalf("want required error, got %v", err)
	}
}

func TestUser_Min(t *testing.T) {
	err := (&User{Name: "a", Email: "a@b.com"}).Validate()
	if err == nil || !strings.Contains(err.Error(), "min 2") {
		t.Fatalf("want min error, got %v", err)
	}
}

func TestUser_BadEmail(t *testing.T) {
	err := (&User{Name: "ada", Email: "bad"}).Validate()
	if err == nil {
		t.Fatal("want email error")
	}
}

func TestSignup_OK(t *testing.T) {
	if err := (&Signup{Username: "ada", Email: "a@b.com"}).Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSignup_Min(t *testing.T) {
	err := (&Signup{Username: "ab", Email: "a@b.com"}).Validate()
	if err == nil || !strings.Contains(err.Error(), "min 3") {
		t.Fatalf("want min error, got %v", err)
	}
}
