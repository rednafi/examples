## endpoints

Companion code for [Generalizing RPC wire plumbing with "endpoints"](https://rednafi.com/go/endpoint-pattern/).

The same `greet.Service` is served over HTTP (`:8080`) and gRPC (`:9090`) by a
single generic adapter per transport. Adding an endpoint means writing one
input struct, one output struct, one business function, and a small decode and
encode pair per transport — the wire plumbing is written once.

### Layout

```
endpoints/
├── api/               # protobuf-generated types
├── errs/              # transport-agnostic error codes
├── greet/             # domain: In, Out, Validate, Service.Greet, Service.CreateUser
├── transport/         # WrapHTTP, WrapGRPC, error mappers
├── httpapi/           # HTTP decoders/encoders + RegisterRoutes
├── grpcapi/           # gRPC decoders/encoders + Server
└── main.go            # `go run . http` or `go run . grpc`
```

### Run tests

```sh
go test ./...
```

### Run servers

```sh
go run . http  # POST http://localhost:8080/greet
go run . grpc  # endpointspb.Endpoints on :9090
```
