package runtime

import (
	"strings"
	"testing"
)

type User struct {
	Name  string `check:"required,min=2"`
	Email string `check:"required,email"`
}

func TestNaive_OK(t *testing.T) {
	if err := ValidateNaive(&User{Name: "ada", Email: "a@b.com"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNaive_Min(t *testing.T) {
	err := ValidateNaive(&User{Name: "a", Email: "a@b.com"})
	if err == nil || !strings.Contains(err.Error(), "min") {
		t.Fatalf("want min error, got %v", err)
	}
}

func TestNaive_Required(t *testing.T) {
	err := ValidateNaive(&User{Email: "a@b.com"})
	if err == nil || !strings.Contains(err.Error(), "required") {
		t.Fatalf("want required error, got %v", err)
	}
}

func TestNaive_BadEmail(t *testing.T) {
	err := ValidateNaive(&User{Name: "ada", Email: "bad"})
	if err == nil {
		t.Fatal("want email parse error")
	}
}

func TestRegistry_OK(t *testing.T) {
	if err := Validate(&User{Name: "ada", Email: "a@b.com"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegistry_Min(t *testing.T) {
	err := Validate(&User{Name: "a", Email: "a@b.com"})
	if err == nil || !strings.Contains(err.Error(), "min") {
		t.Fatalf("want min error, got %v", err)
	}
}

func TestRegistry_UnknownRuleIgnored(t *testing.T) {
	type X struct {
		Name string `check:"required,nosuchrule,min=2"`
	}
	if err := Validate(&X{Name: "ada"}); err != nil {
		t.Fatalf("unknown rule should be skipped, got %v", err)
	}
}
