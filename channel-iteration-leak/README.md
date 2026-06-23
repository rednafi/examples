## channel-iteration-leak

Companion code for
[Channel iteration and goroutine leak](https://rednafi.com/go/channel-iteration-goroutine-leak/).

A `for range` over a channel only ends when the channel is closed. If nothing closes it, the
receiver can sit on the next receive forever. The directories here follow the post in order.

The leak examples use the `goroutineleak` profile that Go 1.27 adds to `runtime/pprof`
(golang/go#74609). On the 1.26 toolchain it sits behind an experiment, so every command below
uses the build prefix:

```sh
GOEXPERIMENT=goroutineleakprofile go test ./...
```

It's on by default in 1.27, where you can drop the prefix.

### leaky-tick

The buggy scheduler. Each job sends one result, and a collector ranges over `results`.
`tick` waits until the collector has seen every result, then returns without closing the
channel. The collector blocks on the next receive and leaks.

```sh
GOEXPERIMENT=goroutineleakprofile go run ./leaky-tick
```

### manual-channel-drain

The small channel example from the middle of the post. Three explicit receives return cleanly.
Ranging over the same channel without closing it blocks after the third value.

```sh
GOEXPERIMENT=goroutineleakprofile go run ./manual-channel-drain
```

### fixed-tick

The one-line fix for the scheduler: close `results` after every job has reported. This keeps
the original split where the collector calls `Done`, but it lets the collector's range finish.

```sh
GOEXPERIMENT=goroutineleakprofile go run ./fixed-tick
```

### fixed-semantics-tick

The final version. The job goroutines own their `WaitGroup` completion via `wg.Go`. A closer
goroutine waits for the jobs, closes `results`, and `tick` drains the channel itself.

```sh
GOEXPERIMENT=goroutineleakprofile go run ./fixed-semantics-tick
```
