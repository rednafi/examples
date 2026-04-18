package ejdemo

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestClaim_NameLiteral proves that `json:"name"` becomes the literal
// string "name" in the generated output. If the generator's behavior
// changed, the Marshal output would no longer contain this substring.
func TestClaim_NameLiteral(t *testing.T) {
	b, err := json.Marshal(User{Name: "ada", Email: "a@b.com"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"name":"ada"`) {
		t.Fatalf(`want "name":"ada" in %s`, b)
	}
}

// TestClaim_OmitemptyIsZeroCheck proves that `json:"email,omitempty"` drops
// the field when the value is the zero value (empty string for a string
// field). With omitempty the encoder emits an `if in.Email != ""` branch,
// so Email is absent from the output when empty.
func TestClaim_OmitemptyIsZeroCheck(t *testing.T) {
	b, err := json.Marshal(User{Name: "ada"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "email") {
		t.Fatalf("email should be absent when empty, got %s", b)
	}

	b, err = json.Marshal(User{Name: "ada", Email: "a@b.com"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"email":"a@b.com"`) {
		t.Fatalf(`want "email":"a@b.com" in %s`, b)
	}
}

// TestClaim_DashSkipsField proves that `json:"-"` causes Admin to never
// appear in Marshal output regardless of value, and that Unmarshal
// ignores "admin" keys in the input.
func TestClaim_DashSkipsField(t *testing.T) {
	b, err := json.Marshal(User{Name: "ada", Admin: true})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(strings.ToLower(string(b)), "admin") {
		t.Fatalf("admin must never appear in output, got %s", b)
	}

	var u User
	if err := json.Unmarshal([]byte(`{"name":"ada","admin":true}`), &u); err != nil {
		t.Fatal(err)
	}
	if u.Admin {
		t.Fatalf("Admin should stay false when decoding, got %+v", u)
	}
}

// TestClaim_RoundTrip sanity-checks that Marshal/Unmarshal are wired up
// correctly through the generated methods.
func TestClaim_RoundTrip(t *testing.T) {
	in := User{Name: "ada", Email: "a@b.com", Admin: true}
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}

	var out User
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	if out.Name != in.Name || out.Email != in.Email {
		t.Fatalf("round trip lost data: in=%+v out=%+v", in, out)
	}
	if out.Admin {
		t.Fatalf("Admin must not round-trip (json:\"-\"), got %+v", out)
	}
}
