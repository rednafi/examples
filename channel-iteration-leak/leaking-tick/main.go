// leaking-tick is the buggy scheduler. On a tick it dispatches the due jobs and
// a background collector ranges over the results channel. Nothing closes the
// channel, so after the last outcome the collector blocks forever and leaks.
//
// Run: GOEXPERIMENT=goroutineleakprofile go run ./leaking-tick
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

// tick dispatches every due job and collects their outcomes by ranging over the
// results channel in a background goroutine. Nothing closes results, so that
// goroutine leaks.
func tick(due []Job) []outcome {
	results := make(chan outcome)

	var wg sync.WaitGroup
	for _, j := range due {
		wg.Add(1)
		go func() {
			results <- outcome{job: j.Name, err: j.Run()} // producer
		}()
	}

	var log []outcome
	go func() {
		for r := range results { // collector: blocks once the jobs stop sending
			log = append(log, r)
			wg.Done()
		}
	}()

	wg.Wait()
	// no close(results): the range above never ends, so the collector leaks
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
