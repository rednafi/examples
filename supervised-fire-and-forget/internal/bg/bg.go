// Package bg runs best-effort background tasks on a fixed worker pool with a
// bounded queue.
package bg

import "sync"

// Background runs tasks with a fixed worker pool and a bounded queue.
type Background struct {
	tasks   chan func()
	workers sync.WaitGroup
	onPanic func(any)

	mu      sync.RWMutex
	stopped bool
}

// New starts workers and returns a Background with room for capacity queued tasks.
// Tasks start in submission order but can finish out of order when workers is greater
// than one. onPanic may be called concurrently by workers.
func New(capacity, workers int, onPanic func(any)) *Background {
	if capacity < 0 {
		panic("bg: capacity must not be negative")
	}
	if workers <= 0 {
		panic("bg: workers must be positive")
	}

	background := &Background{
		tasks:   make(chan func(), capacity),
		onPanic: onPanic,
	}
	for range workers {
		background.workers.Go(background.work)
	}
	return background
}

// Submit adds a task to the queue. It blocks while the queue is full and returns false once
// Stop has closed admission. Submit may run concurrently with other Submit and Stop calls.
// Do not call Submit from a submitted task or onPanic. A task must not call
// runtime.Goexit (t.Fatal does), or its worker exits without a replacement.
func (b *Background) Submit(task func()) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.stopped {
		return false
	}
	b.tasks <- task
	return true
}

func (b *Background) work() {
	for task := range b.tasks {
		b.run(task)
	}
}

func (b *Background) run(task func()) {
	defer func() {
		if value := recover(); value != nil {
			b.handlePanic(value)
		}
	}()
	task()
}

func (b *Background) handlePanic(value any) {
	if b.onPanic == nil {
		return
	}
	defer func() {
		_ = recover()
	}()
	b.onPanic(value)
}

// Stop closes admission, drains the queue, and waits for the workers. It is safe to call
// Stop more than once or concurrently. Do not call Stop from a submitted task or onPanic.
func (b *Background) Stop() {
	b.mu.Lock()
	if !b.stopped {
		b.stopped = true
		close(b.tasks)
	}
	b.mu.Unlock()

	b.workers.Wait()
}
