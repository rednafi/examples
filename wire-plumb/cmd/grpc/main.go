// Run with `go run ./cmd/grpc`. Serves greet.Service over gRPC on :9090.
//
// Reflection is registered so grpcurl can list and call methods without
// shipping the .proto file.
package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/rednafi/examples/wire-plumb/greet"
	egrpc "github.com/rednafi/examples/wire-plumb/grpc"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	users := greet.NewMemoryStore(
		greet.User{ID: 1, Name: "red"},
		greet.User{ID: 2, Name: "blue"},
	)
	svc := greet.NewService(users, logger)

	srv := grpc.NewServer(grpc.UnaryInterceptor(LoggingInterceptor(logger)))
	egrpc.Register(srv, svc)
	reflection.Register(srv)

	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		logger.Error("listen", "err", err)
		os.Exit(1)
	}

	logger.Info("grpc listening", "addr", ":9090")
	if err := srv.Serve(lis); err != nil {
		logger.Error("grpc server", "err", err)
		os.Exit(1)
	}
}

func LoggingInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context, req any,
		info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
	) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		logger.Info("rpc", "method", info.FullMethod, "took", time.Since(start), "err", err)
		return resp, err
	}
}
