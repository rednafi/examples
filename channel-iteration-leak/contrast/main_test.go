// Run: GOEXPERIMENT=goroutineleakprofile go test ./contrast
package main

import (
	"bytes"
	"runtime"
	"runtime/pprof"
	"strings"
	"testing"
)

func leaked() (string, bool) {
	p := pprof.Lookup("goroutineleak")
	if p == nil {
		return "", false
	}
	var b bytes.Buffer
	p.WriteTo(&b, 1)
	return b.String(), p.Count() > 0
}

func TestContrast(t *testing.T) {
	if pprof.Lookup("goroutineleak") == nil {
		t.Skip("rerun with GOEXPERIMENT=goroutineleakprofile")
	}

	// Matched sends and receives: the sender returns, nothing leaks.
	explicitReceive()
	runtime.Gosched()
	if report, ok := leaked(); ok {
		t.Fatalf("explicitReceive should not leak:\n%s", report)
	}

	// Same values, but ranging without a close strands the receiver.
	rangeNoClose()
	runtime.Gosched()
	report, ok := leaked()
	if !ok {
		t.Fatal("rangeNoClose should leak the ranging goroutine")
	}
	if !strings.Contains(report, "rangeNoClose.func") {
		t.Errorf("expected the ranging goroutine in the report:\n%s", report)
	}

	// A buffer doesn't help: range still waits for a close, so it leaks too.
	rangeBuffered()
	runtime.Gosched()
	report, ok = leaked()
	if !ok {
		t.Fatal("rangeBuffered should still leak")
	}
	if !strings.Contains(report, "rangeBuffered.func") {
		t.Errorf("expected the buffered ranging goroutine in the report:\n%s", report)
	}
}
