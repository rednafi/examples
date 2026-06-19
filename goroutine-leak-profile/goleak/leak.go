// Package goleak fails a test on a leaked goroutine using a VerifyNone-shaped
// helper around the stdlib goroutineleak profile.
package goleak

func leak() {
	ch := make(chan int)
	go func() { ch <- 1 }() // nobody receives, the goroutine leaks
}
