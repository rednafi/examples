// Why three explicit receives are safe but ranging over the same channel leaks.
// Run: GOEXPERIMENT=goroutineleakprofile go run ./receive-vs-range
package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
)

// explicitReceive reads three values with three receives, then the goroutine
// returns. Nothing is left blocked.
func explicitReceive() {
	ch := make(chan int)
	go func() {
		<-ch
		<-ch
		<-ch
	}()
	ch <- 1
	ch <- 2
	ch <- 3
}

// rangeNoClose reads the same three values by ranging over the channel. range
// only stops on a close, so after the third value the goroutine blocks for a
// fourth that never comes. It leaks.
func rangeNoClose() {
	ch := make(chan int)
	go func() {
		for range ch { // (1) blocks after the last value, waiting for a close
		}
	}()
	ch <- 1
	ch <- 2
	ch <- 3
	// (2) no close(ch)
}

// rangeBuffered shows a buffer doesn't fix it. The channel holds every value, so
// no send blocks, but range still waits for a close. The goroutine leaks just
// like rangeNoClose.
func rangeBuffered() {
	ch := make(chan int, 3) // room for every value
	go func() {
		for range ch {
		}
	}()
	ch <- 1
	ch <- 2
	ch <- 3
	// no close(ch)
}

func main() {
	explicitReceive()
	rangeNoClose()
	runtime.Gosched() // let the ranging goroutine reach its blocked receive

	p := pprof.Lookup("goroutineleak")
	if p == nil {
		fmt.Fprintln(os.Stderr, "no goroutineleak profile: rerun with GOEXPERIMENT=goroutineleakprofile")
		os.Exit(1)
	}
	p.WriteTo(os.Stdout, 1)
}
