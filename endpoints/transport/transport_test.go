package transport_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/rednafi/examples/endpoints/api"
	"github.com/rednafi/examples/endpoints/greet"
	"github.com/rednafi/examples/endpoints/grpcapi"
	"github.com/rednafi/examples/endpoints/httpapi"
)

// The HTTP and gRPC tests below run the same set of cases against the same
// greet.Service, just over different transports. They demonstrate that one
// business function generalizes across transports without modification.

func TestHTTP_Greet(t *testing.T) {
	mux := http.NewServeMux()
	httpapi.RegisterRoutes(mux, greet.NewService())

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

func TestHTTP_CreateUser(t *testing.T) {
	mux := http.NewServeMux()
	httpapi.RegisterRoutes(mux, greet.NewService())

	first := httptest.NewRequest("POST", "/users", strings.NewReader(`{"email":"a@b.com","age":21}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, first)
	if rec.Code != http.StatusCreated {
		t.Fatalf("first create status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var got struct{ ID string }
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.ID != "u_a@b.com" {
		t.Fatalf("id = %q", got.ID)
	}

	dup := httptest.NewRequest("POST", "/users", strings.NewReader(`{"email":"a@b.com","age":21}`))
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, dup)
	if rec.Code != http.StatusConflict {
		t.Fatalf("dup status = %d, body = %s", rec.Code, rec.Body.String())
	}

	bad := httptest.NewRequest("POST", "/users", strings.NewReader(`{"email":"no-at","age":21}`))
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, bad)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("invalid status = %d", rec.Code)
	}
}

func TestGRPC_Greet(t *testing.T) {
	srv := grpcapi.NewServer(greet.NewService())
	ctx := context.Background()

	out, err := srv.Greet(ctx, &api.GreetRequest{Name: "red", Formality: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Message != "Good day, red." {
		t.Fatalf("message = %q", out.Message)
	}

	_, err = srv.Greet(ctx, &api.GreetRequest{Name: ""})
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("err = %v, want grpc status", err)
	}
	if st.Code() != codes.InvalidArgument {
		t.Fatalf("code = %v, want InvalidArgument", st.Code())
	}
	if !strings.Contains(st.Message(), "name is required") {
		t.Fatalf("message = %q", st.Message())
	}
}

func TestGRPC_CreateUser(t *testing.T) {
	srv := grpcapi.NewServer(greet.NewService())
	ctx := context.Background()

	out, err := srv.CreateUser(ctx, &api.CreateUserRequest{Email: "a@b.com", Age: 21})
	if err != nil {
		t.Fatalf("first create: %v", err)
	}
	if out.Id != "u_a@b.com" {
		t.Fatalf("id = %q", out.Id)
	}

	_, err = srv.CreateUser(ctx, &api.CreateUserRequest{Email: "a@b.com", Age: 21})
	st, _ := status.FromError(err)
	if st.Code() != codes.AlreadyExists {
		t.Fatalf("dup code = %v", st.Code())
	}

	_, err = srv.CreateUser(ctx, &api.CreateUserRequest{Email: "no-at", Age: 21})
	st, _ = status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Fatalf("invalid code = %v", st.Code())
	}
}
