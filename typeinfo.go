package avro

import (
	"fmt"
	"log"
	"reflect"

	"github.com/actgardner/gogen-avro/schema"
)

type AvroRecord interface {
	AvroRecord() RecordInfo
}

type AvroEnum interface {
	NumMembers() int
	String() string
}

type RecordInfo struct {
	// Schema holds the Avro schema of the record.
	Schema string

	// Defaults holds default values for the fields.
	// Each item corresponds to the field at that index and returns
	// a newly created default value for the field.
	// Missing or nil entries are assumed to have no default.
	// TODO assuming a missing entry implies a zero default.
	Defaults []func() interface{}

	// Unions holds entries for union fields.
	// Each item corresponds to the field at that index
	// and holds slice with one value for each member
	// of the union, of type *T, where T is the type used
	// for that member of the union.
	Unions [][]interface{}
}
XXXXX

// azTypeInfo is the representation of the above types used
// by the analyzer. It can represent a record, a field or a type
// inside one of those.
type azTypeInfo struct {
	// avroType holds the Avro schema for the type (with
	// references resolved).
	avroSchema schema.AvroType

	// ftype holds the Go type used for this Avro type (or nil for null).
	ftype reflect.Type

	// makeDefault is a function that returns the default
	// value for a field, or nil if there is no default value.
	makeDefault func() interface{}

	// Entries holds the possible types that can
	// be descended in from this type.
	// For structs (records) this holds an entry
	// for each field; for union types, this holds an
	// entry for each possible type in the union.
	entries []azTypeInfo

	// referenceType holds the type of the closest
	// ancestor struct type containing this type.
	referenceType reflect.Type
}

type goTypeChecker struct {
	defs map[reflect.Type]goTypeDef
}

func newAzTypeInfo(t reflect.Type) (azTypeInfo, error) {
	debugf("azTypeInfo(%v)", t)
	switch v := reflect.Zero(t).Interface().(type) {
	case AvroRecord:
		info := azTypeInfo{
			ftype:   t,
			entries: make([]azTypeInfo, t.NumField()),
		}
		r := v.AvroRecord()
		// TODO consider struct embedding.
		for i := 0; i < t.NumField(); i++ {
			var makeDefault func() interface{}
			var unionVals []interface{}
			if i < len(r.Defaults) {
				makeDefault = r.Defaults[i]
			}
			if i < len(r.Unions) {
				unionVals = r.Unions[i]
			}
			entry, err := newAzTypeInfoFromField(t, t.Field(i).Type, makeDefault, unionVals)
			if err != nil {
				return azTypeInfo{}, err
			}
			info.entries[i] = entry
		}
		debugf("-> record, %d entries", len(info.entries))
		return info, nil
	case AvroEnum:
		return azTypeInfo{}, fmt.Errorf("enum not implemented yet")
	default:
		debugf("-> unknown")
		return azTypeInfo{
			ftype: t,
		}, nil
	}
}

func newAzTypeInfoFromField(refType, t reflect.Type, makeDefault func() interface{}, unionVals []interface{}) (azTypeInfo, error) {
	info := azTypeInfo{
		ftype:         t,
		makeDefault:   makeDefault,
		referenceType: refType,
	}
	if len(unionVals) == 0 {
		return info, nil
	}
	info.entries = make([]azTypeInfo, len(unionVals))
	for i, v := range unionVals {
		var ut reflect.Type
		if v != nil {
			ut = reflect.TypeOf(v).Elem()
		}
		info.entries[i] = azTypeInfo{
			ftype:         ut,
			referenceType: refType,
		}
	}
	return info, nil
}

const debugging = true

func debugf(f string, a ...interface{}) {
	if debugging {
		log.Printf(f, a...)
	}
}
