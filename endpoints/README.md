## endpoints

Companion code for [Generalizing RPC wire plumbing with "endpoints"](https://rednafi.com/go/endpoint-pattern/).

The same `greet.Service` is served over HTTP (`:8080`) and gRPC (`:9090`).
Each transport package contains its own generic `Wrap` adapter, per-endpoint
decoders and encoders, and a small error mapper. The domain package never
imports either transport.

### Layout

```
endpoints/
├── greet/                # domain: types, validation, service, error vocabulary
├── http/                 # HTTP wiring (package http, alias on import)
├── grpc/                 # gRPC wiring (package grpc, alias on import)
│   └── api/              # generated protobuf
└── cmd/
    ├── http/             # HTTP server binary
    └── grpc/             # gRPC server binary
```

### Run tests

```sh
go test ./...
```

### Run servers

```sh
go run ./cmd/http   # POST http://localhost:8080/greet
go run ./cmd/grpc   # greeterpb.Greeter on :9090
```
