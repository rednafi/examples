// manual-channel-drain shows why fixed receives return but ranging over an
// unclosed channel does not.
//
// Run: GOEXPERIMENT=goroutineleakprofile go run ./manual-channel-drain
package main

import (
	"fmt"
	"io"
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

// rangeNoClose reads the same three values by ranging over the channel. The
// range waits for a close that never comes, so the goroutine leaks.
func rangeNoClose() {
	ch := make(chan int)
	go func() {
		for range ch {
		}
	}()
	ch <- 1
	ch <- 2
	ch <- 3
}

func main() {
	p := pprof.Lookup("goroutineleak")
	if p == nil {
		fmt.Fprintln(os.Stderr, "no goroutineleak profile: rerun with GOEXPERIMENT=goroutineleakprofile")
		os.Exit(1)
	}

	explicitReceive()
	runtime.Gosched()
	p.WriteTo(io.Discard, 1)
	fmt.Printf("after explicit receives: %d leaked goroutines\n", p.Count())

	rangeNoClose()
	runtime.Gosched()
	fmt.Println("--- range without close ---")
	p.WriteTo(os.Stdout, 1)
}
