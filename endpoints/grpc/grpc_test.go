package grpc_test

import (
	"context"
	"strings"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/rednafi/examples/endpoints/greet"
	egrpc "github.com/rednafi/examples/endpoints/grpc"
	pb "github.com/rednafi/examples/endpoints/grpc/api"
)

func TestGreet(t *testing.T) {
	srv := egrpc.NewServer(greet.NewService())
	ctx := context.Background()

	out, err := srv.Greet(ctx, &pb.GreetRequest{Name: "red", Formality: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Message != "Good day, red." {
		t.Fatalf("message = %q", out.Message)
	}

	_, err = srv.Greet(ctx, &pb.GreetRequest{Name: ""})
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

func TestSubscribe(t *testing.T) {
	srv := egrpc.NewServer(greet.NewService())
	ctx := context.Background()

	out, err := srv.Subscribe(ctx, &pb.SubscribeRequest{Email: "a@b.com", Formality: 0})
	if err != nil {
		t.Fatalf("first subscribe: %v", err)
	}
	if out.Id == "" {
		t.Fatal("missing id")
	}

	_, err = srv.Subscribe(ctx, &pb.SubscribeRequest{Email: "a@b.com", Formality: 0})
	st, _ := status.FromError(err)
	if st.Code() != codes.AlreadyExists {
		t.Fatalf("dup code = %v", st.Code())
	}

	_, err = srv.Subscribe(ctx, &pb.SubscribeRequest{Email: "no-at", Formality: 0})
	st, _ = status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Fatalf("invalid code = %v", st.Code())
	}
}
