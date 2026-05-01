// Run with `go run ./cmd/server http` or `go run ./cmd/server grpc`.
// Both modes serve the same greet.Service.
package main

import (
	"fmt"
	"log"
	"net"
	nethttp "net/http"
	"os"

	ggrpc "google.golang.org/grpc"

	"github.com/rednafi/examples/endpoints/greet"
	apphttp "github.com/rednafi/examples/endpoints/http"
	appgrpc "github.com/rednafi/examples/endpoints/grpc"
)

func main() {
	mode := "http"
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}
	svc := greet.NewService()

	switch mode {
	case "http":
		mux := nethttp.NewServeMux()
		apphttp.Register(mux, svc)
		fmt.Println("http listening on :8080")
		log.Fatal(nethttp.ListenAndServe(":8080", mux))
	case "grpc":
		srv := ggrpc.NewServer()
		appgrpc.Register(srv, svc)
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
