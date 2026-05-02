package grpc_test

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/rednafi/examples/wire-plumb/greet"
	egrpc "github.com/rednafi/examples/wire-plumb/grpc"
	pb "github.com/rednafi/examples/wire-plumb/grpc/api"
)

func newServer() *egrpc.Server {
	users := greet.NewMemoryStore(
		greet.User{ID: 1, Name: "red"},
	)
	svc := greet.NewService(users, slog.New(slog.NewTextHandler(io.Discard, nil)))
	return egrpc.NewServer(svc)
}

func TestGreet(t *testing.T) {
	srv := newServer()
	ctx := context.Background()

	tests := []struct {
		name        string
		req         *pb.GreetRequest
		wantCode    codes.Code
		wantMessage string
	}{
		{"casual", &pb.GreetRequest{UserId: 1, Formality: 0}, codes.OK, "hey red!"},
		{"formal", &pb.GreetRequest{UserId: 1, Formality: 1}, codes.OK, "Good day, red."},
		{"missing user_id", &pb.GreetRequest{}, codes.InvalidArgument, "user_id is required"},
		{"unknown user", &pb.GreetRequest{UserId: 99}, codes.NotFound, "user 99"},
		{"bad formality", &pb.GreetRequest{UserId: 1, Formality: 7}, codes.InvalidArgument, "formality must be 0 or 1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := srv.Greet(ctx, tt.req)
			if tt.wantCode == codes.OK {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if out.Message != tt.wantMessage {
					t.Fatalf("message = %q, want %q", out.Message, tt.wantMessage)
				}
				return
			}
			st, ok := status.FromError(err)
			if !ok {
				t.Fatalf("err = %v, want grpc status", err)
			}
			if st.Code() != tt.wantCode {
				t.Fatalf("code = %v, want %v", st.Code(), tt.wantCode)
			}
			if !strings.Contains(st.Message(), tt.wantMessage) {
				t.Fatalf("message = %q, want substring %q", st.Message(), tt.wantMessage)
			}
		})
	}
}

func TestFarewell(t *testing.T) {
	srv := newServer()
	ctx := context.Background()

	tests := []struct {
		name        string
		req         *pb.FarewellRequest
		wantCode    codes.Code
		wantMessage string
	}{
		{"known user", &pb.FarewellRequest{UserId: 1}, codes.OK, "bye red!"},
		{"missing user_id", &pb.FarewellRequest{}, codes.InvalidArgument, "user_id is required"},
		{"unknown user", &pb.FarewellRequest{UserId: 99}, codes.NotFound, "user 99"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := srv.Farewell(ctx, tt.req)
			if tt.wantCode == codes.OK {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if out.Message != tt.wantMessage {
					t.Fatalf("message = %q, want %q", out.Message, tt.wantMessage)
				}
				return
			}
			st, ok := status.FromError(err)
			if !ok {
				t.Fatalf("err = %v, want grpc status", err)
			}
			if st.Code() != tt.wantCode {
				t.Fatalf("code = %v, want %v", st.Code(), tt.wantCode)
			}
			if !strings.Contains(st.Message(), tt.wantMessage) {
				t.Fatalf("message = %q, want substring %q", st.Message(), tt.wantMessage)
			}
		})
	}
}
