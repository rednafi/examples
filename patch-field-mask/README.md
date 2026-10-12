# patch-field-mask

Companion code for
[Partial updates with field masks](https://rednafi.com/go/patch-field-mask/).

A `PATCH /users/{id}` endpoint built with the standard library. The JSON body carries
values. The `update_mask` query parameter lists the fields to change. A field missing
from the mask keeps its stored value. A field in the mask takes its value from the body,
and an omitted or `null` value clears it.

```sh
go test -race ./...
go run .
curl -X PATCH 'http://127.0.0.1:8080/users/1?update_mask=name' -d '{"name": "Bob"}'
curl -X PATCH 'http://127.0.0.1:8080/users/1?update_mask=address' -d '{"address": "New Street 2"}'
curl -X PATCH 'http://127.0.0.1:8080/users/1?update_mask=address' -d '{}'
```
