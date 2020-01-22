package avro

import (
	"context"
	"reflect"
)

// EncodingRegistry is used by SingleEncoder to find
// ids for schemas encoded in messages.
type EncodingRegistry interface {
	// AppendSchemaID appends the given schema ID header to buf
	// and returns the resulting slice.
	AppendSchemaID(buf []byte, id int64) []byte

	// IDForSchema returns an ID for the given schema.
	IDForSchema(ctx context.Context, schema string) (int64, error)
}

// SingleEncoder encodes messages in Avro binary format.
// Each message includes a header or wrapper that indicates the schema.
type SingleEncoder struct {
	registry EncodingRegistry
	names    *Names
}

// NewSingleEncoder returns a SingleEncoder instance that encodes single
// messages along with their schema identifier.
//
// Go values unmarshaled through Marshal will have their Avro schemas
// translated with the given Names instance. If names is nil, the global
// namespace will be used.
func NewSingleEncoder(r EncodingRegistry, names *Names) *SingleEncoder {
	if names == nil {
		names = globalNames
	}
	return &SingleEncoder{
		registry: r,
		names:    names,
	}
}

// Marshal returns x marshaled as using the Avro binary encoding,
// along with an identifier that records the type that it was encoded
// with.
func (enc *SingleEncoder) Marshal(ctx context.Context, x interface{}) ([]byte, error) {
	xv := reflect.ValueOf(x)
	avroType, err := avroTypeOf(enc.names, xv.Type())
	if err != nil {
		return nil, err
	}
	id, err := enc.registry.IDForSchema(ctx, avroType.String())
	if err != nil {
		return nil, err
	}
	buf := make([]byte, 0, 100)
	buf = enc.registry.AppendSchemaID(buf, id)
	data, _, err := marshalAppend(enc.names, buf, xv)
	return data, err
}
