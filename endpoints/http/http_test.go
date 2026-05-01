package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rednafi/examples/endpoints/greet"
	ehttp "github.com/rednafi/examples/endpoints/http"
)

func newMux() *http.ServeMux {
	mux := http.NewServeMux()
	ehttp.Register(mux, greet.NewService())
	return mux
}

func TestGreet(t *testing.T) {
	mux := newMux()

	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantBody   string
	}{
		{"casual", `{"name":"red","formality":0}`, 200, "hey red!"},
		{"formal", `{"name":"red","formality":1}`, 200, "Good day, red."},
		{"empty name", `{"name":""}`, 400, "name is required"},
		{"bad formality", `{"name":"red","formality":7}`, 400, "formality must be 0 or 1"},
		{"malformed", `not json`, 400, "malformed json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/greet", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body %q)", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if !strings.Contains(rec.Body.String(), tt.wantBody) {
				t.Fatalf("body = %q, want substring %q", rec.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestSubscribe(t *testing.T) {
	mux := newMux()

	first := httptest.NewRequest("POST", "/subscribe", strings.NewReader(`{"email":"a@b.com","formality":0}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, first)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var got struct{ ID string }
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.ID == "" {
		t.Fatal("missing id")
	}

	dup := httptest.NewRequest("POST", "/subscribe", strings.NewReader(`{"email":"a@b.com","formality":0}`))
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, dup)
	if rec.Code != http.StatusConflict {
		t.Fatalf("dup status = %d", rec.Code)
	}

	bad := httptest.NewRequest("POST", "/subscribe", strings.NewReader(`{"email":"no-at","formality":0}`))
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, bad)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("invalid status = %d", rec.Code)
	}
}
