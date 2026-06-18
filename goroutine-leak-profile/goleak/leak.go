// Package goleak fails a test on a leaked goroutine using the stdlib
// goroutineleak profile, the way uber-go/goleak does it.
package goleak

func leak() {
	ch := make(chan int)
	go func() { ch <- 1 }() // nobody receives, the goroutine leaks
}
