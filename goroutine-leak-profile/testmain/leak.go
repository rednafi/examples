// Package testmain fails the whole test run on a leaked goroutine using a
// TestMain check, the way uber-go/goleak's VerifyTestMain does.
package testmain

func leak() {
	ch := make(chan int)
	go func() { ch <- 1 }() // nobody receives, the goroutine leaks
}
