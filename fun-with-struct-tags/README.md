## fun-with-struct-tags

Companion code for
[Fun with Go struct tags](https://rednafi.com/go/fun-with-struct-tags/).

Two takes on parsing a `check:"required,min=2,email"` struct tag:

- `runtime/` — reflection-based validator, in the naive switch shape and the
  registry shape the blog covers.
- `codegen/` — a `go generate` tool that walks the AST of `types.go` and emits
  a per-type `Validate` method into `zz_generated_check.go`. No reflection at
  call time.

### Run runtime tests

```sh
go test ./runtime/...
```

### Regenerate and run codegen tests

```sh
go generate ./codegen/...
go test ./codegen/...
```
