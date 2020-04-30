package typeinfo

import "github.com/rogpeppe/gogen-avro/v7/schema"

// IsEmptyRecord reports whether def represents the "empty"
// record type that we use to represent records with no fields
// in Avro. It also returns true for genuinely empty record types,
// on the basis that these should be technically compatible
// with the placeholder record type.
func IsEmptyRecord(def *schema.RecordDefinition) bool {
	fields := def.Fields()
	if len(fields) != 1 {
		return len(fields) == 0
	}
	f := fields[0]
	if _, ok := f.Type().(*schema.NullField); !ok {
		return false
	}
	return f.Name() == "_" && f.HasDefault() && f.Default() == nil
}
