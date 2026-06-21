// A tiny cron-style scheduler. On each tick it dispatches the jobs that are due,
// each job reports its outcome on a channel, and one collector ranges over that
// channel to record the run. Forgetting to close the channel leaks the collector.
//
// Run: GOEXPERIMENT=goroutineleakprofile go run ./scheduler
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
// results channel. The collector goroutine leaks: nothing ever closes results.
func tick(due []Job) []outcome {
	results := make(chan outcome)

	var wg sync.WaitGroup
	for _, j := range due {
		wg.Add(1)
		go func() {
			results <- outcome{job: j.Name, err: j.Run()} // (1) producer
		}()
	}

	var log []outcome
	go func() {
		for r := range results { // (2) collector, blocks once the jobs stop sending
			log = append(log, r)
			wg.Done()
		}
	}()

	wg.Wait()
	// (3) no close(results): the range at (2) never ends, so the collector leaks
	return log
}

// tickClosed is tick with the one missing line: close the channel once every job
// has reported, which ends the range and lets the collector return.
func tickClosed(due []Job) []outcome {
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
