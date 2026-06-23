// fixed-tick is the one-line fix for the scheduler: close results after every
// job has reported, so the collector's range can finish.
//
// Run: GOEXPERIMENT=goroutineleakprofile go run ./fixed-tick
package main

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
)

// Job is a unit of scheduled work.
type Job struct {
	Name string
	Run  func() error
}

// outcome is what a job reports once it has run.
type outcome struct {
	job string
	err error
}

func tick(due []Job) []outcome {
	results := make(chan outcome)

	var wg sync.WaitGroup
	for _, j := range due {
		wg.Add(1)
		go func() {
			results <- outcome{job: j.Name, err: j.Run()}
		}()
	}

	var log []outcome
	go func() {
		for r := range results {
			log = append(log, r)
			wg.Done()
		}
	}()

	wg.Wait()
	close(results) // ends the range, the collector returns
	return log
}

func main() {
	due := []Job{
		{Name: "backup", Run: func() error { return nil }},
		{Name: "rotate-logs", Run: func() error { return nil }},
		{Name: "send-digest", Run: func() error { return nil }},
	}

	log := tick(due)
	runtime.Gosched() // let the collector observe the close and return

	p := pprof.Lookup("goroutineleak")
	if p == nil {
		fmt.Fprintln(os.Stderr, "no goroutineleak profile: rerun with GOEXPERIMENT=goroutineleakprofile")
		os.Exit(1)
	}
	p.WriteTo(io.Discard, 1)
	fmt.Printf("recorded %d outcomes\n", len(log))
	fmt.Printf("leaked goroutines: %d\n", p.Count())
}
