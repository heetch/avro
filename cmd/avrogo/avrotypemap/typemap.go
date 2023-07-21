// Package avrotypemap is an internal implementation detail of the avrogo
// program and should not be used externally. The API has no guarantees at all.
package avrotypemap

import (
	"fmt"
	"reflect"

	"github.com/actgardner/gogen-avro/v10/schema"

	"github.com/heetch/avro"
	"github.com/heetch/avro/internal/typeinfo"
)

// ExternalTypeResult is used as a JSON-marshaled value
// in the result printed by the program generated as part
// of a avrogo run to find information about
// external Avro types.
type ExternalTypeResult struct {
	// Error is non-empty when there was a problem getting
	// info with the type.
	Error string
	// Schema holds the Avro schema for the type.
	Schema string
	// Map maps from Avro fully qualified names used
	// in the types to Go types.
	Map map[string]GoType
}

type GoType struct {
	PkgPath string
	Name    string
}

func (t GoType) String() string {
	return fmt.Errorf("%q.%s", t.PkgPath, t.Name)
}

// AvroTypeMap returns a value that maps the names used by the Avro schema
// implied by t to the Go types that are used to implement them.
func AvroTypeMap(t reflect.Type) (map[string]GoType, error) {
	avroType, err := avro.TypeOf(reflect.Zero(t).Interface())
	if err != nil {
		return nil, err
	}
	at, err := typeinfo.ParseSchema(avroType.String(), nil)
	if err != nil {
		return nil, err
	}
	w := &walker{
		walked: make(map[schema.AvroType]bool),
		types:  make(map[schema.QualifiedName]reflect.Type),
	}
	if err := w.walk(at, t, typeinfo.Info{}); err != nil {
		return nil, err
	}
	m := make(map[string]GoType)
	for qname, t := range w.types {
		m[qname.String()] = GoType{
			PkgPath: t.PkgPath(),
			Name:    t.Name(),
		}
	}
	return m, nil
}

type walker struct {
	walked map[schema.AvroType]bool
	types  map[schema.QualifiedName]reflect.Type
}

func (w *walker) walk(at schema.AvroType, t reflect.Type, info typeinfo.Info) error {
	if w.walked[at] {
		return nil
	}
	w.walked[at] = true
	switch at := at.(type) {
	case *schema.Reference:
		if oldt, ok := w.types[at.TypeName]; ok {
			// We've found this name before. Check that it
			// maps to the same Go type here.
			if oldt != t {
				return fmt.Errorf("type clash on %s: %v vs %v", at.TypeName, t, oldt)
			}
			return nil
		} else {
			w.types[at.TypeName] = t
		}
		switch def := at.Def.(type) {
		case *schema.RecordDefinition:
			if t.Kind() != reflect.Struct {
				return fmt.Errorf("expected struct")
			}
			if len(info.Entries) == 0 {
				// The type itself might contribute information.
				info1, err := typeinfo.ForType(t)
				if err != nil {
					return fmt.Errorf("cannot get info for %s: %v", info.Type, err)
				}
				info = info1
			}
			if len(info.Entries) != len(def.Fields()) {
				return fmt.Errorf("field count mismatch")
			}
			for i, f := range def.Fields() {
				ft := t.Field(info.Entries[i].FieldIndex)
				err := w.walk(f.Type(), ft.Type, info.Entries[i])
				if err != nil {
					return err
				}
			}
		}
	case *schema.ArrayField:
		if t.Kind() != reflect.Slice {
			return fmt.Errorf("expected slice, got %s", t)
		}
		return w.walk(at.ItemType(), t.Elem(), info)
	case *schema.MapField:
		if t.Kind() != reflect.Map {
			return fmt.Errorf("expected map, got %s", t)
		}
		return w.walk(at.ItemType(), t.Elem(), info)
	case *schema.UnionField:
		if len(info.Entries) != len(at.ItemTypes()) {
			return fmt.Errorf("union entry count mismatch")
		}
		for i, at := range at.ItemTypes() {
			err := w.walk(at, info.Entries[i].Type, info.Entries[i])
			if err != nil {
				return err
			}
		}
	}
	return nil
}
