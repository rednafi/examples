// Run with `go run ./cmd/http`. Serves the greet.Service over HTTP on :8080.
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/rednafi/examples/endpoints/greet"
	ehttp "github.com/rednafi/examples/endpoints/http"
)

func main() {
	svc := greet.NewService()
	mux := http.NewServeMux()
	ehttp.Register(mux, svc)

	fmt.Println("http listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
