package avro

import (
	"context"
	"reflect"
)

type EncodingRegistry interface {
	// AppendSchemaID appends the given schema ID header to buf
	// and returns the resulting slice.
	AppendSchemaID(buf []byte, id int64) []byte

	// IDForSchema returns an ID for the given schema.
	// If the schema wasn't found, it returns ErrSchemaNotFound.
	IDForSchema(ctx context.Context, schema string) (int64, error)
}

type SingleEncoder struct {
	registry EncodingRegistry
}

// NewSingleEncoder returns a SingleEncoder instance that encodes single
// messages along with their schema identifier.
func NewSingleEncoder(r EncodingRegistry) *SingleEncoder {
	return &SingleEncoder{
		registry: r,
	}
}

// Marshal returns x marshaled as using the Avro binary encoding,
// along with an identifier that records the type that it was encoded
// with.
func (enc *SingleEncoder) Marshal(ctx context.Context, x interface{}) ([]byte, error) {
	xv := reflect.ValueOf(x)
	avroType, err := avroTypeOf(xv.Type())
	if err != nil {
		return nil, err
	}
	id, err := enc.registry.IDForSchema(ctx, avroType.String())
	if err != nil {
		return nil, err
	}
	buf := make([]byte, 0, 100)
	buf = enc.registry.AppendSchemaID(buf, id)
	return marshalAppend(buf, xv, avroType)
}
