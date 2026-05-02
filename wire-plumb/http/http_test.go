package http_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rednafi/examples/wire-plumb/greet"
	ehttp "github.com/rednafi/examples/wire-plumb/http"
)

func newMux() *http.ServeMux {
	users := greet.NewMemoryStore(
		greet.User{ID: 1, Name: "red"},
	)
	svc := greet.NewService(users, slog.New(slog.NewTextHandler(io.Discard, nil)))
	mux := http.NewServeMux()
	ehttp.Register(mux, svc)
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
		{"casual", `{"user_id":1,"formality":0}`, 200, "hey red!"},
		{"formal", `{"user_id":1,"formality":1}`, 200, "Good day, red."},
		{"missing user_id", `{"formality":0}`, 400, "user_id is required"},
		{"unknown user", `{"user_id":99,"formality":0}`, 404, "user 99"},
		{"bad formality", `{"user_id":1,"formality":7}`, 400, "formality must be 0 or 1"},
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
			if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
				t.Fatalf("Content-Type = %q, want application/json", ct)
			}
		})
	}
}

func TestFarewell(t *testing.T) {
	mux := newMux()

	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantBody   string
	}{
		{"known user", `{"user_id":1}`, 200, "bye red!"},
		{"missing user_id", `{}`, 400, "user_id is required"},
		{"unknown user", `{"user_id":99}`, 404, "user 99"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/farewell", strings.NewReader(tt.body))
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
