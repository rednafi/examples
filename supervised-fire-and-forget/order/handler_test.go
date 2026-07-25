package order

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rednafi/examples/supervised-fire-and-forget/internal/bg"
)

type inheritedKey struct{}

type taskFuncs struct {
	notify     func(context.Context, string) error
	diagnostic func(context.Context, string) error
}

func (f taskFuncs) SendNotification(ctx context.Context, user string) error {
	return f.notify(ctx, user)
}

func (f taskFuncs) WriteDiagnosticLog(ctx context.Context, user string) error {
	return f.diagnostic(ctx, user)
}

type taskObservation struct {
	ctx       context.Context
	err       error
	inherited string
	requestID string
	user      string
	deadline  time.Time
}

func observeTask(ctx context.Context, user string) taskObservation {
	inherited, _ := ctx.Value(inheritedKey{}).(string)
	deadline, _ := ctx.Deadline()
	return taskObservation{
		ctx:       ctx,
		err:       ctx.Err(),
		inherited: inherited,
		requestID: requestIDFromContext(ctx),
		user:      user,
		deadline:  deadline,
	}
}

func TestHandlerDetachesRequestCancellation(t *testing.T) {
	background := bg.New(2, 1, nil)
	workerStarted := make(chan struct{})
	releaseWorker := make(chan struct{})
	released := false
	defer func() {
		if !released {
			close(releaseWorker)
		}
		background.Stop()
	}()
	mustSubmit(t, background, func() {
		close(workerStarted)
		<-releaseWorker
	})
	<-workerStarted

	var notification, diagnostic taskObservation
	tasks := taskFuncs{
		notify: func(ctx context.Context, user string) error {
			notification = observeTask(ctx, user)
			return nil
		},
		diagnostic: func(ctx context.Context, user string) error {
			diagnostic = observeTask(ctx, user)
			return nil
		},
	}
	h, err := NewHandler(
		background,
		tasks,
		slog.New(slog.DiscardHandler),
		time.Minute,
	)
	if err != nil {
		t.Fatalf("NewHandler() returned unexpected error: %v", err)
	}

	parent := context.WithValue(t.Context(), inheritedKey{}, "trace-data")
	parent = context.WithValue(parent, requestIDKey{}, "req-123")
	parent, cancel := context.WithCancel(parent)
	req := httptest.NewRequestWithContext(
		parent,
		http.MethodPost,
		"/orders?user=alice",
		nil,
	)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)
	cancel()

	if got, want := rec.Code, http.StatusAccepted; got != want {
		t.Fatalf("POST /orders status = %d, want %d", got, want)
	}

	started := time.Now()
	close(releaseWorker)
	released = true
	background.Stop()

	assertTaskContext(t, notification, started, "", "")
	assertTaskContext(t, diagnostic, started, "trace-data", "req-123")
}

func assertTaskContext(
	t *testing.T,
	got taskObservation,
	started time.Time,
	wantInherited string,
	wantRequestID string,
) {
	t.Helper()

	if got.err != nil {
		t.Errorf("task context error while running = %v, want nil", got.err)
	}
	if got, want := got.inherited, wantInherited; got != want {
		t.Errorf("inherited context value = %q, want %q", got, want)
	}
	if got, want := got.requestID, wantRequestID; got != want {
		t.Errorf("request ID = %q, want %q", got, want)
	}
	if got, want := got.user, "alice"; got != want {
		t.Errorf("task user = %q, want %q", got, want)
	}
	if got.deadline.Before(started.Add(59*time.Second)) ||
		got.deadline.After(started.Add(61*time.Second)) {
		t.Errorf(
			"task deadline = %s, want about one minute after %s",
			got.deadline,
			started,
		)
	}
	if got, want := got.ctx.Err(), context.Canceled; !errors.Is(got, want) {
		t.Errorf("task context error after return = %v, want %v", got, want)
	}
}

func TestHandlerAppliesTaskTimeout(t *testing.T) {
	background := bg.New(2, 1, nil)

	var notificationErr, diagnosticErr error
	tasks := taskFuncs{
		notify: func(ctx context.Context, _ string) error {
			<-ctx.Done()
			notificationErr = ctx.Err()
			return notificationErr
		},
		diagnostic: func(ctx context.Context, _ string) error {
			<-ctx.Done()
			diagnosticErr = ctx.Err()
			return diagnosticErr
		},
	}
	h, err := NewHandler(
		background,
		tasks,
		slog.New(slog.DiscardHandler),
		time.Millisecond,
	)
	if err != nil {
		t.Fatalf("NewHandler() returned unexpected error: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/orders?user=alice", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	background.Stop()

	for name, got := range map[string]error{
		"notification":   notificationErr,
		"diagnostic_log": diagnosticErr,
	} {
		t.Run(name, func(t *testing.T) {
			if want := context.DeadlineExceeded; !errors.Is(got, want) {
				t.Errorf("task error = %v, want %v", got, want)
			}
		})
	}
}

func TestHandlerStartsTimeoutAtExecution(t *testing.T) {
	background := bg.New(2, 1, nil)
	workerStarted := make(chan struct{})
	releaseWorker := make(chan struct{})
	released := false
	defer func() {
		if !released {
			close(releaseWorker)
		}
		background.Stop()
	}()
	mustSubmit(t, background, func() {
		close(workerStarted)
		<-releaseWorker
	})
	<-workerStarted

	var notificationErr, diagnosticErr error
	tasks := taskFuncs{
		notify: func(ctx context.Context, _ string) error {
			notificationErr = ctx.Err()
			return notificationErr
		},
		diagnostic: func(ctx context.Context, _ string) error {
			diagnosticErr = ctx.Err()
			return diagnosticErr
		},
	}
	h, err := NewHandler(
		background,
		tasks,
		slog.New(slog.DiscardHandler),
		100*time.Millisecond,
	)
	if err != nil {
		t.Fatalf("NewHandler() returned unexpected error: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/orders?user=alice", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	// Hold both tasks in the queue for three times the task timeout. If the
	// timeout clock started at enqueue, their contexts are expired by the time
	// the worker runs them.
	time.Sleep(300 * time.Millisecond)
	close(releaseWorker)
	released = true
	background.Stop()

	for name, got := range map[string]error{
		"notification":   notificationErr,
		"diagnostic_log": diagnosticErr,
	} {
		t.Run(name, func(t *testing.T) {
			if got != nil {
				t.Errorf("task context error at execution = %v, want nil", got)
			}
		})
	}
}

func TestNewHandlerRunsOperations(t *testing.T) {
	background := bg.New(2, 1, nil)
	var logs bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, nil))
	h, err := NewHandler(
		background,
		Operations{},
		logger,
		time.Second,
	)
	if err != nil {
		t.Fatalf("NewHandler() returned unexpected error: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/orders?user=alice", nil)
	req.Header.Set("X-Request-ID", "req-default-actions")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got, want := rec.Code, http.StatusAccepted; got != want {
		t.Fatalf("POST /orders status = %d, want %d", got, want)
	}
	background.Stop()

	output := logs.String()
	for _, message := range []string{
		"notification sent",
		"diagnostic log written",
		"user=alice",
		"request_id=req-default-actions",
	} {
		if !bytes.Contains([]byte(output), []byte(message)) {
			t.Errorf("handler log does not contain %q:\n%s", message, output)
		}
	}
}

func TestHandlerRejectsInvalidRequest(t *testing.T) {
	background := bg.New(1, 1, nil)

	var ran atomic.Int64
	tasks := taskFuncs{
		notify: func(context.Context, string) error {
			ran.Add(1)
			return nil
		},
		diagnostic: func(context.Context, string) error {
			ran.Add(1)
			return nil
		},
	}
	h, err := NewHandler(
		background,
		tasks,
		slog.New(slog.DiscardHandler),
		time.Second,
	)
	if err != nil {
		t.Fatalf("NewHandler() returned unexpected error: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/orders", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got, want := rec.Code, http.StatusBadRequest; got != want {
		t.Errorf("POST /orders status = %d, want %d", got, want)
	}
	background.Stop()
	if got := ran.Load(); got != 0 {
		t.Errorf("tasks run for an invalid request = %d, want 0", got)
	}
}

func TestWait(t *testing.T) {
	t.Run("delay_elapsed", func(t *testing.T) {
		if err := wait(t.Context(), 0); err != nil {
			t.Errorf("wait(ctx, 0) = %v, want nil", err)
		}
	})

	t.Run("context_canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		if got, want := wait(ctx, time.Hour), context.Canceled; !errors.Is(got, want) {
			t.Errorf("wait(canceledCtx, 1h) = %v, want %v", got, want)
		}
	})
}

func TestOperationsHonorCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	tasks := Operations{}

	for name, action := range map[string]func(context.Context, string) error{
		"notification":   tasks.SendNotification,
		"diagnostic_log": tasks.WriteDiagnosticLog,
	} {
		t.Run(name, func(t *testing.T) {
			if got, want := action(ctx, "alice"), context.Canceled; !errors.Is(got, want) {
				t.Errorf("action(canceledCtx, alice) = %v, want %v", got, want)
			}
		})
	}
}

func TestNewHandlerRejectsInvalidConfiguration(t *testing.T) {
	background := bg.New(1, 1, nil)
	defer background.Stop()
	logger := slog.New(slog.DiscardHandler)
	tasks := taskFuncs{
		notify:     func(context.Context, string) error { return nil },
		diagnostic: func(context.Context, string) error { return nil },
	}

	tests := []struct {
		name       string
		background *bg.Background
		tasks      Tasks
		logger     *slog.Logger
		timeout    time.Duration
	}{
		{
			name:    "nil_background",
			tasks:   tasks,
			logger:  logger,
			timeout: time.Second,
		},
		{
			name:       "nil_tasks",
			background: background,
			logger:     logger,
			timeout:    time.Second,
		},
		{
			name:       "nil_logger",
			background: background,
			tasks:      tasks,
			timeout:    time.Second,
		},
		{
			name:       "zero_timeout",
			background: background,
			tasks:      tasks,
			logger:     logger,
		},
		{
			name:       "negative_timeout",
			background: background,
			tasks:      tasks,
			logger:     logger,
			timeout:    -time.Second,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewHandler(
				test.background,
				test.tasks,
				test.logger,
				test.timeout,
			)
			if err == nil {
				t.Error("NewHandler() error = nil, want non-nil")
			}
		})
	}
}

func mustSubmit(t *testing.T, background *bg.Background, task func()) {
	t.Helper()
	if !background.Submit(task) {
		t.Fatal("Submit() = false, want true")
	}
}
