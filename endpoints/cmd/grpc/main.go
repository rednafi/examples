// Run with `go run ./cmd/grpc`. Serves the greet.Service over gRPC on :9090.
package main

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/rednafi/examples/endpoints/greet"
	egrpc "github.com/rednafi/examples/endpoints/grpc"
)

func main() {
	svc := greet.NewService()
	srv := grpc.NewServer()
	egrpc.Register(srv, svc)

	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("grpc listening on :9090")
	log.Fatal(srv.Serve(lis))
}
