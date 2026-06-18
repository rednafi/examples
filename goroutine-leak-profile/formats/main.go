// All output formats of the goroutineleak profile, plus the Count() gotcha.
// Run: GOEXPERIMENT=goroutineleakprofile go run ./formats
package main

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
)

func leakSend() {
	ch := make(chan int)
	go func() { ch <- 1 }() // nobody receives, the send blocks forever
}

func leakRange() {
	ch := make(chan int)
	go func() {
		for range ch { // ch is never closed, the range never ends
		}
	}()
}

func main() {
	leakSend()
	leakRange()
	runtime.Gosched() // let both goroutines reach their blocked state

	p := pprof.Lookup("goroutineleak")
	if p == nil {
		fmt.Fprintln(os.Stderr, "no goroutineleak profile: rerun with GOEXPERIMENT=goroutineleakprofile")
		os.Exit(1)
	}

	// Gotcha: Count() is 0 until a WriteTo runs the leak-detecting GC cycle.
	fmt.Println("Count() before WriteTo:", p.Count())

	var proto bytes.Buffer
	p.WriteTo(&proto, 0) // debug=0: gzipped pprof protobuf for `go tool pprof`
	fmt.Printf("debug=0: %d gzipped bytes\n", proto.Len())
	fmt.Println("Count() after WriteTo:", p.Count())

	fmt.Println("--- debug=1: leaked goroutines as text ---")
	p.WriteTo(os.Stdout, 1)

	fmt.Println("--- debug=2: full dump, leaked ones tagged (leaked) ---")
	p.WriteTo(os.Stdout, 2)
}
