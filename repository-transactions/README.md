## repository-transactions

Companion code for
[How do you handle transactions with the repository pattern?](https://rednafi.com/shards/2026/03/transactions-with-repository-pattern/)

Also see:
[Do you need a repository layer on top of sqlc?](https://rednafi.com/shards/2026/03/repository-layer-over-sqlc/)
[Repository pattern & transactions in Go](https://rednafi.com/go/repository-pattern-and-transactions/)

### Run tests

```sh
CGO_ENABLED=1 go test ./...
```

### Run the server

```sh
CGO_ENABLED=1 go run ./cmd/
```

```sh
curl -X POST localhost:8080/books -d '{"title":"DDIA"}'
curl localhost:8080/books/1
```
