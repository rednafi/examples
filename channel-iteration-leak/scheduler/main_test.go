// Run: GOEXPERIMENT=goroutineleakprofile go test ./scheduler
package main

import (
	"bytes"
	"io"
	"runtime"
	"runtime/pprof"
	"strings"
	"testing"
)

// leaked runs the leak-detecting GC cycle and returns the report plus whether any
// goroutine is provably stuck. With the experiment off it reports nothing.
func leaked() (string, bool) {
	p := pprof.Lookup("goroutineleak")
	if p == nil {
		return "", false
	}
	var b bytes.Buffer
	p.WriteTo(&b, 1)
	return b.String(), p.Count() > 0
}

// leakCount runs the leak-detecting GC cycle and returns how many goroutines are
// provably stuck right now. The profile is process-global, so other tests' leaks
// linger; compare a before/after delta rather than an absolute count.
func leakCount() int {
	p := pprof.Lookup("goroutineleak")
	if p == nil {
		return 0
	}
	p.WriteTo(io.Discard, 1)
	return p.Count()
}

func TestTickStream(t *testing.T) {
	if pprof.Lookup("goroutineleak") == nil {
		t.Skip("rerun with GOEXPERIMENT=goroutineleakprofile")
	}

	due := []Job{
		{Name: "backup", Run: func() error { return nil }},
		{Name: "rotate-logs", Run: func() error { return nil }},
		{Name: "send-digest", Run: func() error { return nil }},
	}

	runtime.Gosched()
	before := leakCount()

	var streamed []string
	log := tickStream(due, func(o outcome) { streamed = append(streamed, o.job) })
	runtime.Gosched()

	if after := leakCount(); after != before {
		report, _ := leaked()
		t.Fatalf("tickStream leaked %d goroutine(s):\n%s", after-before, report)
	}
	if len(log) != len(due) {
		t.Fatalf("got %d outcomes, want %d", len(log), len(due))
	}
	if len(streamed) != len(due) {
		t.Fatalf("process called %d times, want %d", len(streamed), len(due))
	}
}

func TestTick(t *testing.T) {
	if pprof.Lookup("goroutineleak") == nil {
		t.Skip("rerun with GOEXPERIMENT=goroutineleakprofile")
	}

	due := []Job{
		{Name: "backup", Run: func() error { return nil }},
		{Name: "rotate-logs", Run: func() error { return nil }},
	}

	// The closed version is clean. Check it first, before any leak exists.
	tickClosed(due)
	runtime.Gosched()
	if report, ok := leaked(); ok {
		t.Fatalf("tickClosed should not leak:\n%s", report)
	}

	// tick forgets the close, so the collector ranges forever.
	tick(due)
	runtime.Gosched()
	report, ok := leaked()
	if !ok {
		t.Fatal("tick should leak the collector")
	}
	if !strings.Contains(report, "tick.func") {
		t.Errorf("expected the collector stack in the report:\n%s", report)
	}
}
