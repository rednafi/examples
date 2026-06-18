// Pull leaks from a running process on demand with a signal handler.
// Run:  GOEXPERIMENT=goroutineleakprofile go run ./signaldump
// Then: kill -USR1 <pid>   (the pid is printed on start)
package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
)

func main() {
	ch := make(chan int)
	go func() { ch <- 1 }() // a leak to find

	if pprof.Lookup("goroutineleak") == nil {
		fmt.Fprintln(os.Stderr, "no goroutineleak profile: rerun with GOEXPERIMENT=goroutineleakprofile")
		os.Exit(1)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGUSR1)
	fmt.Printf("pid %d: run `kill -USR1 %d` to dump leaks, Ctrl-C to quit\n", os.Getpid(), os.Getpid())
	for range sig {
		pprof.Lookup("goroutineleak").WriteTo(os.Stdout, 1)
	}
}
