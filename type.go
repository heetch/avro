package avro

import (
	"fmt"

	"github.com/actgardner/gogen-avro/parser"
	"github.com/actgardner/gogen-avro/resolver"
	"github.com/actgardner/gogen-avro/schema"
)

// Type represents an Avro schema type.
type Type struct {
	schema   string
	avroType schema.AvroType
}

// ParseType parses an Avro schema in the format defined by the Avro
// specification at https://avro.apache.org/docs/current/spec.html.
func ParseType(s string) (*Type, error) {
	ns := parser.NewNamespace(false)
	avroType, err := ns.TypeForSchema([]byte(s))
	if err != nil {
		return nil, err
	}
	for _, def := range ns.Roots {
		if err := resolver.ResolveDefinition(def, ns.Definitions); err != nil {
			return nil, fmt.Errorf("cannot resolve references in schema: %v", err)
		}
	}
	return &Type{
		schema:   s,
		avroType: avroType,
	}, nil
}

func (t *Type) String() string {
	return t.schema
}
