// Run with `go run ./cmd/grpc`. Serves greet.Service over gRPC on :9090.
// Reflection is registered so grpcurl can list and call methods without the
// .proto file.
package main

import (
	"context"
	"log"
	"log/slog"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/rednafi/examples/wire-plumb/greet"
	egrpc "github.com/rednafi/examples/wire-plumb/grpc"
)

func main() {
	users := greet.NewMemoryStore(
		greet.User{ID: 1, Name: "red"},
		greet.User{ID: 2, Name: "blue"},
	)
	svc := greet.NewService(users, slog.Default())
	srv := grpc.NewServer(grpc.UnaryInterceptor(LoggingInterceptor))
	egrpc.Register(srv, svc)
	reflection.Register(srv)

	lis, _ := net.Listen("tcp", ":9090")
	log.Fatal(srv.Serve(lis))
}

func LoggingInterceptor(
	ctx context.Context, req any,
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
) (any, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	log.Printf("%s took=%s err=%v", info.FullMethod, time.Since(start), err)
	return resp, err
}
