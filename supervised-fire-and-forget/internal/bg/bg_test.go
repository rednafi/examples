package bg

import (
	"bytes"
	"runtime"
	"runtime/pprof"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestBackgroundDrainsInOrder(t *testing.T) {
	background := New(3, 1, nil)

	var got []int
	for value := range 3 {
		mustSubmit(t, background, func() {
			got = append(got, value)
		})
	}
	background.Stop()

	want := []int{0, 1, 2}
	if !slices.Equal(got, want) {
		t.Errorf("task order = %v, want %v", got, want)
	}
}

func TestBackgroundRunsWorkerPool(t *testing.T) {
	const workers = 3
	background := New(workers+1, workers, nil)

	started := make(chan struct{}, workers+1)
	release := make(chan struct{})
	for range workers + 1 {
		mustSubmit(t, background, func() {
			started <- struct{}{}
			<-release
		})
	}

	for range workers {
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("worker did not start")
		}
	}
	select {
	case <-started:
		t.Fatal("New() started more workers than requested")
	case <-time.After(20 * time.Millisecond):
	}

	close(release)
	background.Stop()
}

func TestBackgroundAcceptsConcurrentSubmissions(t *testing.T) {
	const submissions = 100
	background := New(4, 4, nil)

	var submitters sync.WaitGroup
	var accepted, ran atomic.Int64
	for range submissions {
		submitters.Go(func() {
			if background.Submit(func() {
				ran.Add(1)
			}) {
				accepted.Add(1)
			}
		})
	}
	submitters.Wait()
	background.Stop()

	if got, want := accepted.Load(), int64(submissions); got != want {
		t.Errorf("tasks accepted = %d, want %d", got, want)
	}
	if got, want := ran.Load(), int64(submissions); got != want {
		t.Errorf("tasks run = %d, want %d", got, want)
	}
}

func TestBackgroundSubmitStopRace(t *testing.T) {
	const (
		iterations = 50
		submitters = 32
		stoppers   = 4
	)

	for iteration := range iterations {
		background := New(4, 4, nil)
		start := make(chan struct{})

		var calls sync.WaitGroup
		var accepted, ran atomic.Int64
		for range submitters {
			calls.Go(func() {
				<-start
				if background.Submit(func() {
					ran.Add(1)
				}) {
					accepted.Add(1)
				}
			})
		}

		var stops sync.WaitGroup
		for range stoppers {
			stops.Go(func() {
				<-start
				background.Stop()
			})
		}

		close(start)
		calls.Wait()
		stops.Wait()

		if background.Submit(func() { ran.Add(1_000) }) {
			t.Fatalf("iteration %d: Submit() after Stop() = true, want false", iteration)
		}
		background.Stop()

		if got, want := ran.Load(), accepted.Load(); got != want {
			t.Fatalf(
				"iteration %d: tasks run = %d, want accepted count %d",
				iteration,
				got,
				want,
			)
		}
	}
}

func TestNewRejectsInvalidConfiguration(t *testing.T) {
	tests := []struct {
		name     string
		capacity int
		workers  int
	}{
		{name: "negative_capacity", capacity: -1, workers: 1},
		{name: "zero_workers", capacity: 1, workers: 0},
		{name: "negative_workers", capacity: 1, workers: -1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("New() did not panic")
				}
			}()
			New(test.capacity, test.workers, nil)
		})
	}
}

func TestBackgroundBoundsPendingWork(t *testing.T) {
	background := New(1, 1, nil)

	var ran atomic.Int64
	started := make(chan struct{})
	release := make(chan struct{})
	mustSubmit(t, background, func() {
		close(started)
		<-release
		ran.Add(1)
	})
	<-started

	mustSubmit(t, background, func() { ran.Add(1) })

	submitted := make(chan bool, 1)
	go func() {
		submitted <- background.Submit(func() { ran.Add(1) })
	}()
	waitForSubmit(t, background)

	stopped := make(chan struct{})
	stopStarted := make(chan struct{})
	go func() {
		close(stopStarted)
		background.Stop()
		close(stopped)
	}()
	<-stopStarted
	select {
	case <-stopped:
		t.Fatal("Stop() returned while Submit() was in progress")
	case <-time.After(20 * time.Millisecond):
	}

	close(release)
	select {
	case accepted := <-submitted:
		if !accepted {
			t.Fatal("Submit() in progress when Stop() started returned false")
		}
	case <-time.After(time.Second):
		t.Fatal("Submit() remained blocked after a worker received a task")
	}
	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("Stop() did not drain the accepted tasks")
	}

	if got, want := ran.Load(), int64(3); got != want {
		t.Errorf("tasks run = %d, want %d", got, want)
	}
}

func TestBackgroundStopIsConcurrentAndIdempotent(t *testing.T) {
	const callers = 8
	background := New(1, 1, nil)

	started := make(chan struct{})
	release := make(chan struct{})
	var releaseOnce sync.Once
	defer releaseOnce.Do(func() { close(release) })

	mustSubmit(t, background, func() {
		close(started)
		<-release
	})
	<-started

	stopped := make(chan struct{}, callers)
	for range callers {
		go func() {
			background.Stop()
			stopped <- struct{}{}
		}()
	}

	select {
	case <-stopped:
		t.Fatal("Stop() returned before the running task")
	case <-time.After(20 * time.Millisecond):
	}

	releaseOnce.Do(func() { close(release) })
	for range callers {
		select {
		case <-stopped:
		case <-time.After(time.Second):
			t.Fatal("Stop() did not return after the task finished")
		}
	}

	background.Stop()
	if background.Submit(func() {}) {
		t.Fatal("Submit() after Stop() = true, want false")
	}
}

func TestBackgroundRecoversPanicAndContinues(t *testing.T) {
	var panicValue any
	background := New(2, 1, func(value any) {
		panicValue = value
	})

	var ran atomic.Bool
	mustSubmit(t, background, func() { panic("boom") })
	mustSubmit(t, background, func() { ran.Store(true) })
	background.Stop()

	if got, want := panicValue, "boom"; got != want {
		t.Errorf("panic value = %v, want %v", got, want)
	}
	if !ran.Load() {
		t.Error("task after panic did not run")
	}
}

func TestBackgroundRecoversConcurrentPanics(t *testing.T) {
	const panics = 32
	var recovered atomic.Int64
	background := New(panics, 4, func(any) {
		recovered.Add(1)
	})

	for range panics {
		mustSubmit(t, background, func() { panic("boom") })
	}
	var ran atomic.Bool
	mustSubmit(t, background, func() { ran.Store(true) })
	background.Stop()

	if got, want := recovered.Load(), int64(panics); got != want {
		t.Errorf("panics recovered = %d, want %d", got, want)
	}
	if !ran.Load() {
		t.Error("task after panics did not run")
	}
}

func TestBackgroundRecoversPanicWithoutHandler(t *testing.T) {
	background := New(2, 1, nil)

	var ran atomic.Bool
	mustSubmit(t, background, func() { panic("boom") })
	mustSubmit(t, background, func() { ran.Store(true) })
	background.Stop()

	if !ran.Load() {
		t.Error("task after panic did not run")
	}
}

func TestBackgroundRecoversPanicFromHandler(t *testing.T) {
	background := New(2, 1, func(any) {
		panic("panic handler failed")
	})

	var ran atomic.Bool
	mustSubmit(t, background, func() { panic("task failed") })
	mustSubmit(t, background, func() { ran.Store(true) })
	background.Stop()

	if !ran.Load() {
		t.Error("task after panic handler failure did not run")
	}
}

func TestBackgroundDoesNotLeak(t *testing.T) {
	profile := pprof.Lookup("goroutineleak")
	if profile == nil {
		t.Skip("goroutine leak profile requires Go 1.27 or GOEXPERIMENT=goroutineleakprofile")
	}

	background := New(1, 1, nil)
	mustSubmit(t, background, func() {})
	background.Stop()
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

func mustSubmit(t *testing.T, background *Background, task func()) {
	t.Helper()
	if !background.Submit(task) {
		t.Fatal("Submit() = false, want true")
	}
}

func waitForSubmit(t *testing.T, background *Background) {
	t.Helper()

	deadline := time.Now().Add(time.Second)
	for {
		if !background.mu.TryLock() {
			return
		}
		background.mu.Unlock()

		if time.Now().After(deadline) {
			t.Fatal("Submit() did not acquire the read lock")
		}
		runtime.Gosched()
	}
}
