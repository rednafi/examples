## wire-plumb

Companion code for [Generalizing wire plumbing in Go unary RPCs](https://rednafi.com/go/wire-plumb/).

The same `greet.Service` is served over HTTP (`:8080`) and gRPC (`:9090`). Each
transport package contains its own generic `Wrap` adapter, per-endpoint decode
and encode functions, and a small error mapper. The domain package never imports
either transport.

### Layout

```
wire-plumb/
├── greet/                # domain: types, validation, service, error vocabulary
│   ├── service.go        # Service, Greet, Farewell, In/Out types
│   ├── store.go          # User, UserStore interface, in-memory store
│   └── errors.go         # Code enum, *greet.Error, error constructors
├── http/                 # HTTP wiring (alias as ehttp on import)
│   ├── http.go           # Wrap, decode/encode, Register, error map
│   └── http_test.go
├── grpc/                 # gRPC wiring (alias as egrpc on import)
│   ├── grpc.go           # Wrap, decode/encode, Server, Register, error map
│   ├── grpc_test.go
│   └── api/              # generated protobuf
└── cmd/
    ├── http/             # HTTP server binary, request logger
    └── grpc/             # gRPC server binary, unary interceptor, reflection
```

### Sample data

Both server binaries seed two users:

| ID  | Name |
| --- | ---- |
| 1   | red  |
| 2   | blue |

### Run tests

```sh
make test     # or: go test ./...
```

### Run the HTTP server

```sh
make run-http # or: go run ./cmd/http
```

Talk to it with `curl`:

```sh
curl -s localhost:8080/greet    -d '{"user_id":1,"formality":0}'
# {"message":"hey red!"}

curl -s localhost:8080/greet    -d '{"user_id":1,"formality":1}'
# {"message":"Good day, red."}

curl -s localhost:8080/farewell -d '{"user_id":2}'
# {"message":"bye blue!"}

curl -s -i localhost:8080/greet -d '{"user_id":99}'
# HTTP/1.1 404 Not Found … {"message":"user 99"}

curl -s -i localhost:8080/greet -d 'not json'
# HTTP/1.1 400 Bad Request   … {"message":"malformed json"}
```

### Run the gRPC server

```sh
make run-grpc # or: go run ./cmd/grpc
```

Reflection is on, so `grpcurl` works without the `.proto` file:

```sh
grpcurl -plaintext localhost:9090 list
# greeterpb.Greeter

grpcurl -plaintext -d '{"user_id":1,"formality":1}' \
    localhost:9090 greeterpb.Greeter/Greet
# {"message":"Good day, red."}

grpcurl -plaintext -d '{"user_id":2}' \
    localhost:9090 greeterpb.Greeter/Farewell
# {"message":"bye blue!"}

grpcurl -plaintext -d '{"user_id":99}' \
    localhost:9090 greeterpb.Greeter/Greet
# ERROR: Code: NotFound, Message: user 99
```

Install `grpcurl` from `https://github.com/fullstorydev/grpcurl` if you don't have
it (`go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest`).

### Regenerate protobuf

```sh
brew install protobuf  # or your OS package manager
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

make generate
```
