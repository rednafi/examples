## go-fix

Companion code for [Modernizers & go fix](https://rednafi.com/go/go-fix/).

Go 1.26 rebuilt `go fix` on the analysis framework. A library author can ship an API
migration that downstream users apply by running plain `go fix`: keep the old function as a
forwarding wrapper, deprecate it, and mark it with a `//go:fix inline` directive.

### greet + demo

`greet.Hello` is deprecated in favor of `greet.Greet`. It forwards to `Greet` and carries a
`//go:fix inline` directive. `demo` is a caller still on the old API:

```sh
go fix -diff ./demo
go fix ./demo
```

`-diff` prints the patch and changes nothing, exiting non-zero when there's something to
apply, which also makes it a CI check.

### client + clientdemo

The directive handles more than renames. `client.Fetch` hardcodes a 30-second timeout that
`client.FetchTimeout` makes explicit, and the deprecated `client.Config` alias points at
`client.Options`. One run migrates a caller off both, spelling out the hidden default and
adding the `time` import:

```sh
go fix -diff ./clientdemo
go fix ./clientdemo
```
