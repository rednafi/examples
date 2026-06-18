// Forgotten close: a consumer ranges over a channel the producer never closes,
// so it blocks on receive forever after the last value.
//
// Run: GOEXPERIMENT=goroutineleakprofile go run ./stream
package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
)

func stream(work []int) {
	out := make(chan int)
	go func() {
		for v := range out { // (1) keeps pulling until out is closed
			handle(v)
		}
	}()
	for _, v := range work {
		out <- v
	}
	// (2) no close(out): the range never ends, so the goroutine leaks
}

// handle stands in for processing each value.
func handle(v int) {}

func main() {
	stream([]int{1, 2, 3})
	runtime.Gosched() // let the consumer reach its blocked receive

	p := pprof.Lookup("goroutineleak")
	if p == nil {
		fmt.Fprintln(os.Stderr, "no goroutineleak profile: rerun with GOEXPERIMENT=goroutineleakprofile")
		os.Exit(1)
	}
	p.WriteTo(os.Stdout, 1)
}
