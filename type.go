package avro

import (
	"github.com/rogpeppe/gogen-avro/v7/schema"

	"github.com/heetch/avro/internal/typeinfo"
)

// Type represents an Avro schema type.
type Type struct {
	schema   string
	avroType schema.AvroType
}

// ParseType parses an Avro schema in the format defined by the Avro
// specification at https://avro.apache.org/docs/current/spec.html.
func ParseType(s string) (*Type, error) {
	avroType, err := typeinfo.ParseSchema(s, nil)
	if err != nil {
		return nil, err
	}
	return &Type{
		schema:   s,
		avroType: avroType,
	}, nil
}

func (t *Type) String() string {
	return t.schema
}

// name returns the fully qualified Avro name for the type,
// or the empty string if it's not a definition.
func (t *Type) name() string {
	ref, ok := t.avroType.(*schema.Reference)
	if !ok {
		return ""
	}
	return ref.TypeName.String()
}
