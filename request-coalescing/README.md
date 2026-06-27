# request-coalescing

Request coalescing with [`golang.org/x/sync/singleflight`][sf]: many concurrent callers
asking for the same key make the work happen once. Companion code for the post
[Request coalescing with Go singleflight](https://rednafi.com/go/request-coalescing/).

- `store.go` is the cache-aside path: serve from a local cache, and on a miss fetch the
  value once (coalescing concurrent misses) before caching it. The fetch stands in for an
  HTTP/gRPC call or a database query.
- `coalesce.go` is the bare wrapper. `Load` uses `Do`; `LoadChan` uses `DoChan` so each
  caller keeps its own cancellation while the shared fetch gets its own timeout.
- `LoadMaxWait` shows `Forget`: a caller can stop waiting and make the next caller start a
  fresh fetch without canceling the one already in flight.
- The tests prove a hundred concurrent loads trigger exactly one fetch, a warm cache serves
  the rest, a caller can cancel without killing shared work, shared fetches are bounded, and
  shared-result metrics count recipients.

```
go test -race -v ./...
```

[sf]: https://pkg.go.dev/golang.org/x/sync/singleflight
