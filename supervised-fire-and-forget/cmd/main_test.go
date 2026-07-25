package main

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"runtime"
	"runtime/pprof"
	"strings"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	testCtx, cancelTest := context.WithTimeout(t.Context(), 3*time.Second)
	defer cancelTest()
	serverCtx, stopServer := context.WithCancel(testCtx)

	var logs bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, nil))
	ready := make(chan string, 1)
	errs := make(chan error, 1)
	go func() {
		errs <- run(serverCtx, runConfig{
			address: "127.0.0.1:0",
			logger:  logger,
			ready:   ready,
		})
	}()

	var address string
	select {
	case address = <-ready:
	case <-testCtx.Done():
		t.Fatalf("run() did not become ready: %v", testCtx.Err())
	}

	req, err := http.NewRequestWithContext(
		testCtx,
		http.MethodPost,
		"http://"+address+"/orders?user=alice",
		nil,
	)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("X-Request-ID", "req-main-test")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /orders: %v", err)
	}
	_, readErr := io.Copy(io.Discard, resp.Body)
	closeErr := resp.Body.Close()
	if readErr != nil {
		t.Errorf("read response body: %v", readErr)
	}
	if closeErr != nil {
		t.Errorf("close response body: %v", closeErr)
	}
	if got, want := resp.StatusCode, http.StatusAccepted; got != want {
		t.Errorf("POST /orders status = %d, want %d", got, want)
	}

	stopServer()
	select {
	case err := <-errs:
		if err != nil {
			t.Fatalf("run() returned unexpected error: %v", err)
		}
	case <-testCtx.Done():
		t.Fatalf("run() did not stop: %v", testCtx.Err())
	}

	output := logs.String()
	for _, message := range []string{
		"server listening",
		"notification sent",
		"diagnostic log written",
		"request_id=req-main-test",
	} {
		if !strings.Contains(output, message) {
			t.Errorf("run() log does not contain %q:\n%s", message, output)
		}
	}
	assertNoGoroutineLeaks(t)
}

func TestRunRejectsInvalidConfiguration(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	tests := []struct {
		name   string
		parent context.Context
		cfg    runConfig
	}{
		{
			name: "nil_context",
			cfg:  runConfig{address: "127.0.0.1:0", logger: logger},
		},
		{
			name:   "empty_address",
			parent: t.Context(),
			cfg:    runConfig{logger: logger},
		},
		{
			name:   "nil_logger",
			parent: t.Context(),
			cfg:    runConfig{address: "127.0.0.1:0"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := run(test.parent, test.cfg); err == nil {
				t.Error("run() error = nil, want non-nil")
			}
		})
	}
}

func TestRunReportsListenError(t *testing.T) {
	err := run(t.Context(), runConfig{
		address: "127.0.0.1:-1",
		logger:  slog.New(slog.DiscardHandler),
	})
	if err == nil {
		t.Fatal("run() error = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), "listen on") {
		t.Errorf("run() error = %q, want it to contain %q", err, "listen on")
	}
}

func assertNoGoroutineLeaks(t *testing.T) {
	t.Helper()

	profile := pprof.Lookup("goroutineleak")
	if profile == nil {
		t.Log("goroutine leak profile requires Go 1.27 or GOEXPERIMENT=goroutineleakprofile")
		return
	}

	runtime.Gosched()
	// WriteTo runs the leak detection pass; Count is only meaningful after it.
	var report bytes.Buffer
	if err := profile.WriteTo(&report, 1); err != nil {
		t.Fatalf("goroutine leak profile: %v", err)
	}
	if got := profile.Count(); got != 0 {
		t.Errorf("goroutine leak profile count = %d, want 0\n%s", got, report.String())
	}
}
