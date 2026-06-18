// Replica fan-out leak: the same request goes to every replica, the first
// answer wins, and the slower replicas block forever trying to send. Same root
// cause as ./earlyreturn, just framed as replicas.
//
// Run: GOEXPERIMENT=goroutineleakprofile go run ./replicate
package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
)

func replicate(replicas []func() string) string {
	results := make(chan string) // unbuffered
	for _, r := range replicas {
		go func() { results <- r() }() // (1) every replica races to send its answer
	}
	return <-results // (2) first answer wins; the slower replicas block forever
}

func main() {
	replicas := []func() string{
		func() string { return "a" },
		func() string { return "b" },
		func() string { return "c" },
	}
	fmt.Println("replicate returned:", replicate(replicas))

	runtime.Gosched() // let the slower replicas reach their blocked sends
	p := pprof.Lookup("goroutineleak")
	if p == nil {
		fmt.Fprintln(os.Stderr, "no goroutineleak profile: rerun with GOEXPERIMENT=goroutineleakprofile")
		os.Exit(1)
	}
	p.WriteTo(os.Stdout, 1)
}
