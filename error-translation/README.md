## error-translation

Companion code for
[Error translation between layers in Go](https://rednafi.com/go/error-translation/).

### http

The HTTP example is self-contained with an in-memory SQLite database.

```sh
cd http
CGO_ENABLED=1 go run .
```

```sh
curl -X POST localhost:8080/users -d '{"name":"alice","email":"alice@example.com"}'
curl localhost:8080/users/1
curl localhost:8080/users/999
curl -X POST localhost:8080/users -d '{"name":"bob","email":"alice@example.com"}'
```

### grpc

The gRPC example requires generated protobuf code. Install `protoc`,
`protoc-gen-go`, and `protoc-gen-go-grpc`, then generate:

```sh
cd grpc
make generate
CGO_ENABLED=1 go run .
```
