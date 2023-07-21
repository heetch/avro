package typeinfo

import (
	"github.com/actgardner/gogen-avro/v10/parser"
	"github.com/actgardner/gogen-avro/v10/resolver"
	"github.com/actgardner/gogen-avro/v10/schema"
	"gopkg.in/errgo.v2/fmt/errors"
)

// ParseSchema parses the given Avro type and resolves
// all its references.
// If ns is non-nil, it will be used as the namespace for
// the definitions.
func ParseSchema(s string, ns *parser.Namespace) (schema.AvroType, error) {
	// TODO this function doesn't really belong in the typeinfo package
	// but it doesn't seem worth making a new package just for this.
	if ns == nil {
		ns = parser.NewNamespace(false)
	}
	avroType, err := ns.TypeForSchema([]byte(s))
	if err != nil {
		return nil, errors.Newf("invalid schema %q: %v", s, err)
	}
	for _, def := range ns.Roots {
		if err := resolver.ResolveDefinition(def, ns.Definitions); err != nil {
			return nil, errors.Newf("cannot resolve references in schema\n%s\n: %v", s, err)
		}
	}
	return avroType, nil
}
