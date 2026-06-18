// Run: GOEXPERIMENT=goroutineleakprofile go test ./testmain
package testmain

import (
	"bytes"
	"fmt"
	"os"
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

// verifyTestMain mirrors goleak.VerifyTestMain.
func verifyTestMain(m *testing.M) {
	code := m.Run()
	if code == 0 {
		if report, ok := leaked(); ok {
			fmt.Fprintf(os.Stderr, "leaked goroutines:\n%s", report)
			code = 1
		}
	}
	os.Exit(code)
}

func TestMain(m *testing.M) {
	verifyTestMain(m)
}

func TestRun(t *testing.T) {
	leak()
	runtime.Gosched()
}
