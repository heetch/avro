# Avro - Go-idiomatic encoding and decoding of Avro data

This package provides both a [code generator](https://pkg.go.dev/github.com/heetch/avro/cmd/avro-generate-go) that generates Go data
structures from [Avro schemas](https://avro.apache.org/docs/1.9.1/spec.html) and a mapping between native
Go data types and Avro schemas.

The API is modelled after that of Go's standard library [encoding/json
package](https://golang.org/pkg/encoding/json).

The documentation can be found [here](https://pkg.go.dev/github.com/heetch/avro).

It also provides support for encoding and decoding messages
using an [Avro schema registry](https://docs.confluent.io/current/schema-registry/index.html) - see
[github.com/heetch/avro/avroregistry](https://pkg.go.dev/github.com/heetch/avro/avroregistry).

## How are Avro schemas represented as Go datatypes?

When the `avro-generate-go` command generates Go datatypes from Avro schemas, it uses the following rules:

- `"int"` is represented as `int`
- `"long"` is represented as `int64`
- `"float"` is represented as `float32`
- `"double"` is represented as `float64`
- `"string"` is represented as `string`
- `"boolean"` is represented as `bool`
- `"bytes"` is represented as `[]byte`
- `"null"` is represented as the Go value `nil`
- `{"type": "array", "items": T}` is represented as `[]T`
- `{"type": "map", "values": T}` is represented as `map[string]T`
- `{"type": "enum", "name": "E", "symbols": ["red", "green", "blue"]}` is represented a Go int type with `String`, `MarshalText` and `UnmarshalText` methods so it will encode as a string when used in JSON.
- `{"type": "fixed", "size": 123, "name": "F"}` will encode as a Go `[123]byte`  type named `F`
- `["null", T]` encodes as `*T`
- `[T, "null"]` encodes as `*T`
- `[T₁, T₂, ...]` (a union) encodes as `interface{}` that should hold only the types for `T₁`, `T₂`, etc.
- `{"type": "record", "name": "R", "fields": [....]}` encodes as a Go struct type named `R` with corresponding fields.

## Comparison with other Go Avro packages

[github.com/linkedin/goavro/v2](https://pkg.go.dev/github.com/linkedin/goavro/v2),
is oriented towards dynamic processing of Avro data. It does not provide an idiomatic way to marshal/unmarshal
Avro data into Go struct values. It does, however, provide good support for encoding and decoding with the
standard [Avro JSON format](https://avro.apache.org/docs/1.9.1/spec.html#json_encoding), which this
package does not.

[github.com/actgardner/gogen-avro](https://github.com/actgardner/gogen-avro) was the original
inspiration for this package. It generates Go code for Avro schemas. It uses a neat VM-based schema
for encoding and decoding (and is also used by this package under the hood), but the generated Go
data structures are awkward to use and don't reflect the data structures that people would idiomatically
define in Go.

For example,  in `gogen-avro` the Avro type `["null", "int"]` (either `null` or an integer) is represented as a
struct containing three members, and an associated enum type:

```go
type UnionNullIntTypeEnum int

const (
	UnionNullIntTypeEnumNull UnionNullIntTypeEnum = 0
	UnionNullIntTypeEnumInt UnionNullIntTypeEnum = 1
)

type UnionNullInt struct {
	Null *types.NullVal
	Int int32
	UnionType UnionNullIntTypeEnum
}
```

With `heetch/avro`, the above type is simply represented as a `*int`, a representation
likely to be familiar to most Go users.
