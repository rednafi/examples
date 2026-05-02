.PHONY: generate test run-http run-grpc

generate:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		grpc/api/greeter.proto

test:
	go test ./...

run-http:
	go run ./cmd/http

run-grpc:
	go run ./cmd/grpc
