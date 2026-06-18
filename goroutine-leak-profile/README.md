## goroutine-leak-profile

Companion code for
[Accepted proposal: a goroutine leak profile in the Go standard library](https://rednafi.com/shards/2026/06/go-goroutine-leak-profile/).

Go 1.27 adds a `goroutineleak` profile to `runtime/pprof` (golang/go#74609). It leans on
the garbage collector to find goroutines blocked on a channel or lock that nothing
runnable can ever reach, so everything it reports is stuck for good, with no false
positives.

On the 1.26 toolchain the profile is behind an experiment, so every command below needs
the build prefix:

```sh
GOEXPERIMENT=goroutineleakprofile go run ./earlyreturn
```

It's on by default in 1.27, where you can drop the prefix. Each directory is a standalone
`main` (or test) you can run on its own. In VSCode, `.vscode/settings.json` sets the
experiment for the test runner and the integrated terminal.

### earlyreturn

A `wg.Go` worker sends its result on an unbuffered channel; the first error returns early
and the workers still queued to send block forever. Background:
[Early return and goroutine leak](https://rednafi.com/go/early-return-and-goroutine-leak/).

```sh
GOEXPERIMENT=goroutineleakprofile go run ./earlyreturn
```

### replicate

The same leak as `earlyreturn` in another guise: fan the same request out to several
replicas, keep the first answer, and the slower replicas block forever on their sends.

```sh
GOEXPERIMENT=goroutineleakprofile go run ./replicate
```

### stream

A consumer ranges over a channel the producer forgets to `close`, so it blocks on receive
forever after the last value.

```sh
GOEXPERIMENT=goroutineleakprofile go run ./stream
```

### formats

The `debug` levels (`0` gzipped protobuf, `1` text, `2` full dump) and the gotcha that
`Count()` reads `0` until a `WriteTo` runs the detecting GC cycle.

```sh
GOEXPERIMENT=goroutineleakprofile go run ./formats
```

### signaldump

Dump leaks from a running process on demand with a `SIGUSR1` handler.

```sh
GOEXPERIMENT=goroutineleakprofile go run ./signaldump
# in another shell, using the pid printed on start:
kill -USR1 <pid>
```

### server

Importing `net/http/pprof` registers `/debug/pprof/goroutineleak` on a live server. Read
it with curl or `go tool pprof`.

```sh
GOEXPERIMENT=goroutineleakprofile go run ./server
# in another shell:
curl 'localhost:6060/debug/pprof/goroutineleak?debug=1'
go tool pprof -top 'http://localhost:6060/debug/pprof/goroutineleak'
```

### goleak

A `verifyNone(t)` helper you `defer` like `goleak.VerifyNone(t)`. `TestRun` leaks on
purpose, so with the experiment on it fails and prints the stranded goroutine. Without the
experiment the profile isn't registered, so the test skips instead of failing.

```sh
GOEXPERIMENT=goroutineleakprofile go test ./goleak
```

### testmain

The same idea moved into `TestMain`, like `goleak.VerifyTestMain`. `TestRun` passes on its
own, but the run fails because `TestMain` checks for leaks after the suite.

```sh
GOEXPERIMENT=goroutineleakprofile go test ./testmain
```
