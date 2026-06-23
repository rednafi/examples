## channel-iteration-leak

Companion code for
[Goroutine leak through channel iteration misuse](https://rednafi.com/go/channel-iteration-goroutine-leak/).

A `for range` over a channel that nobody closes leaks the receiver: the range only ends on a
`close`, so after the last value the goroutine blocks forever. Draining the same channel with
a fixed number of receives is fine.

The leaks are caught with the `goroutineleak` profile that Go 1.27 adds to `runtime/pprof`
(golang/go#74609). On the 1.26 toolchain it sits behind an experiment, so every command below
needs the build prefix:

```sh
GOEXPERIMENT=goroutineleakprofile go test ./...
```

It's on by default in 1.27, where you can drop the prefix. See the sibling
[goroutine-leak-profile](../goroutine-leak-profile) directory for the profile's other output
formats and entry points.

### scheduler

A tiny cron-style scheduler. On a tick it dispatches the due jobs, each reports its outcome
on a channel, and one collector ranges over that channel. `tick` forgets to close it and
leaks the collector; `tickClosed` is the same code with the missing `close`. `tickFixed` is
`tick` written correctly: each job marks itself done with `wg.Go`, a separate goroutine waits
and closes `results`, and the drain runs in `tickFixed` itself. Drop the close there and the
caller deadlocks on the first run instead of leaking a background collector.

```sh
GOEXPERIMENT=goroutineleakprofile go run ./scheduler
GOEXPERIMENT=goroutineleakprofile go test ./scheduler
```

### receive-vs-range

The same channel, two ways. `explicitReceive` reads three values with three receives and the
goroutine returns. `rangeNoClose` reads the same three by ranging, then blocks for a fourth
that never comes. Only the second leaks.

```sh
GOEXPERIMENT=goroutineleakprofile go run ./receive-vs-range
GOEXPERIMENT=goroutineleakprofile go test ./receive-vs-range
```
