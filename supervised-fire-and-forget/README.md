# supervised-fire-and-forget

Companion code for
[Supervised fire-and-forget in Go](https://rednafi.com/go/supervised-fire-and-forget/).

```sh
go test -race ./...
go run ./cmd
curl -i -X POST -H 'X-Request-ID: req-123' 'http://127.0.0.1:8080/orders?user=alice'
```
