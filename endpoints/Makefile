.PHONY: generate run-http run-grpc test

generate:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/endpoints.proto

test:
	go test ./...

run-http:
	go run . http

run-grpc:
	go run . grpc
