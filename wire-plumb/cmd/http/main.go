// Run with `go run ./cmd/http`. Serves greet.Service over HTTP on :8080.
package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/rednafi/examples/wire-plumb/greet"
	ehttp "github.com/rednafi/examples/wire-plumb/http"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	users := greet.NewMemoryStore(
		greet.User{ID: 1, Name: "red"},
		greet.User{ID: 2, Name: "blue"},
	)
	svc := greet.NewService(users, logger)

	mux := http.NewServeMux()
	ehttp.Register(mux, svc)

	logger.Info("http listening", "addr", ":8080")
	if err := http.ListenAndServe(":8080", RequestLogger(logger, mux)); err != nil {
		logger.Error("http server", "err", err)
		os.Exit(1)
	}
}

func RequestLogger(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logger.Info("request", "method", r.Method, "path", r.URL.Path, "took", time.Since(start))
	})
}
