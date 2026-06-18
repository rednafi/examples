// Run: GOEXPERIMENT=goroutineleakprofile go test ./goleak
package goleak

import (
	"bytes"
	"runtime"
	"runtime/pprof"
	"testing"
)

func leaked() (string, bool) {
	p := pprof.Lookup("goroutineleak")
	if p == nil {
		return "", false // experiment off, nothing to detect
	}
	var b bytes.Buffer
	p.WriteTo(&b, 1)
	return b.String(), p.Count() > 0
}

// verifyNone mirrors goleak.VerifyNone.
func verifyNone(t *testing.T) {
	t.Helper()
	if report, ok := leaked(); ok {
		t.Fatalf("leaked goroutines:\n%s", report)
	}
}

func TestRun(t *testing.T) {
	defer verifyNone(t)
	leak() // the code under test leaks a goroutine
	runtime.Gosched()
}
