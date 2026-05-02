// Run with `go run ./cmd/http`. Serves greet.Service over HTTP on :8080.
package main

import (
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/rednafi/examples/wire-plumb/greet"
	ehttp "github.com/rednafi/examples/wire-plumb/http"
)

func main() {
	users := greet.NewMemoryStore(
		greet.User{ID: 1, Name: "red"},
		greet.User{ID: 2, Name: "blue"},
	)
	svc := greet.NewService(users, slog.Default())
	mux := http.NewServeMux()
	ehttp.Register(mux, svc)
	log.Fatal(http.ListenAndServe(":8080", RequestLogger(mux)))
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s took=%s", r.Method, r.URL.Path, time.Since(start))
	})
}
