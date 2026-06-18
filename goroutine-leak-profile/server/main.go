// Pull leaks from a running service over HTTP via net/http/pprof.
// Run:   GOEXPERIMENT=goroutineleakprofile go run ./server
// Curl:  curl 'localhost:6060/debug/pprof/goroutineleak?debug=1'
// Pprof: go tool pprof -top 'http://localhost:6060/debug/pprof/goroutineleak'
package main

import (
	"net/http"
	_ "net/http/pprof" // registers /debug/pprof/goroutineleak
)

func main() {
	ch := make(chan int)
	go func() { ch <- 1 }() // leak

	http.ListenAndServe("localhost:6060", nil)
}
