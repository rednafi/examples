// Early-return goroutine leak: an unbuffered result channel strands the workers
// still queued to send once the first error returns early.
//
// Companion to https://rednafi.com/go/early-return-and-goroutine-leak/.
// Run: GOEXPERIMENT=goroutineleakprofile go run ./earlyreturn
package main

import (
	"errors"
	"fmt"
	"os"
	"runtime/pprof"
	"sync"
	"time"
)

func run(tasks []func() error) error {
	errs := make(chan error) // unbuffered
	var wg sync.WaitGroup
	for _, task := range tasks {
		wg.Go(func() { errs <- task() }) // (1) each task sends on the unbuffered channel
	}
	for range tasks {
		if err := <-errs; err != nil {
			return err // (2) first error returns early; the queued senders block forever
		}
	}
	wg.Wait()
	return nil
}

func main() {
	tasks := []func() error{
		func() error { return errors.New("boom") },                     // fails immediately
		func() error { time.Sleep(20 * time.Millisecond); return nil }, // slower work
		func() error { time.Sleep(20 * time.Millisecond); return nil },
	}
	fmt.Println("run returned:", run(tasks))

	// The two slow tasks wake up and block on their send. Give them a beat,
	// then dump whatever leaked.
	time.Sleep(100 * time.Millisecond)

	p := pprof.Lookup("goroutineleak")
	if p == nil {
		fmt.Fprintln(os.Stderr, "no goroutineleak profile: rerun with GOEXPERIMENT=goroutineleakprofile")
		os.Exit(1)
	}
	p.WriteTo(os.Stdout, 1)
}
