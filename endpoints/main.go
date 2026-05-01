// Run with `go run . http` or `go run . grpc` to start either server.
// Both serve the same greet.Service.
package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"google.golang.org/grpc"

	"github.com/rednafi/examples/endpoints/api"
	"github.com/rednafi/examples/endpoints/greet"
	"github.com/rednafi/examples/endpoints/grpcapi"
	"github.com/rednafi/examples/endpoints/httpapi"
)

func main() {
	mode := "http"
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}
	svc := greet.NewService()

	switch mode {
	case "http":
		mux := http.NewServeMux()
		httpapi.RegisterRoutes(mux, svc)
		fmt.Println("http listening on :8080")
		log.Fatal(http.ListenAndServe(":8080", mux))
	case "grpc":
		srv := grpc.NewServer()
		api.RegisterEndpointsServer(srv, grpcapi.NewServer(svc))
		lis, err := net.Listen("tcp", ":9090")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("grpc listening on :9090")
		log.Fatal(srv.Serve(lis))
	default:
		log.Fatalf("unknown mode %q (want http or grpc)", mode)
	}
}
