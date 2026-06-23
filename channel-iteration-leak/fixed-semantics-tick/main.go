// fixed-semantics-tick keeps the collector out of the WaitGroup accounting.
// Each job marks itself done via wg.Go. A separate goroutine waits for the jobs
// and closes results, while tick drains the channel itself.
//
// Run: GOEXPERIMENT=goroutineleakprofile go run ./fixed-semantics-tick
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
		wg.Go(func() {
			results <- outcome{job: j.Name, err: j.Run()}
		})
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var log []outcome
	for r := range results {
		log = append(log, r)
	}
	return log
}

func main() {
	due := []Job{
		{Name: "backup", Run: func() error { return nil }},
		{Name: "rotate-logs", Run: func() error { return nil }},
		{Name: "send-digest", Run: func() error { return nil }},
	}

	log := tick(due)
	runtime.Gosched()

	p := pprof.Lookup("goroutineleak")
	if p == nil {
		fmt.Fprintln(os.Stderr, "no goroutineleak profile: rerun with GOEXPERIMENT=goroutineleakprofile")
		os.Exit(1)
	}
	p.WriteTo(io.Discard, 1)
	fmt.Printf("recorded %d outcomes\n", len(log))
	fmt.Printf("leaked goroutines: %d\n", p.Count())
}
