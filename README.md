# Avro - Go-idiomatic encoding and decoding of Avro data

This package provides both a code generator that generates Go data
structures from [Avro schemas](https://avro.apache.org/docs/1.9.1/spec.html) and a mapping between native
Go data types and Avro schemas.

The API is modelled after that of Go's standard library [encoding/json
package](https://golang.org/pkg/encoding/json).

The documentation can be found [here](https://pkg.go.dev/github.com/heetch/avro).

It also provides support for encoding and decoding messages
using an [Avro schema registry](https://docs.confluent.io/current/schema-registry/index.html) - see
[github.com/heetch/avro/avroregistry](https://pkg.go.dev/github.com/heetch/avro/avroregistry).

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
