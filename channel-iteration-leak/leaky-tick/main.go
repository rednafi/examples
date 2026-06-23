// leaky-tick is the buggy scheduler from the post. On a tick it dispatches the
// due jobs, a background collector ranges over the results channel, and nothing
// closes the channel. After the last result, the collector waits forever.
//
// Run: GOEXPERIMENT=goroutineleakprofile go run ./leaky-tick
package main

import (
	"fmt"
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

// tick starts every due job and collects results in a background goroutine.
// Nothing closes results, so the collector leaks after all jobs have sent.
func tick(due []Job) []outcome {
	results := make(chan outcome)

	var wg sync.WaitGroup
	for _, j := range due {
		wg.Add(1)
		go func() {
			results <- outcome{job: j.Name, err: j.Run()} // (1)
		}()
	}

	var log []outcome
	go func() {
		for r := range results { // (2)
			log = append(log, r)
			wg.Done()
		}
	}()

	wg.Wait()
	// (3) no close(results)
	return log
}

func main() {
	due := []Job{
		{Name: "backup", Run: func() error { return nil }},
		{Name: "rotate-logs", Run: func() error { return nil }},
		{Name: "send-digest", Run: func() error { return nil }},
	}

	tick(due)
	runtime.Gosched() // let the collector reach its blocked receive

	p := pprof.Lookup("goroutineleak")
	if p == nil {
		fmt.Fprintln(os.Stderr, "no goroutineleak profile: rerun with GOEXPERIMENT=goroutineleakprofile")
		os.Exit(1)
	}
	fmt.Println("--- debug=1: the leaking line ---")
	p.WriteTo(os.Stdout, 1)
	fmt.Println("--- debug=2: full dump, leaked goroutine tagged (leaked) ---")
	p.WriteTo(os.Stdout, 2)
}
