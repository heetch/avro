package avro

import (
	"fmt"
	"log"
	"reflect"
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
	// Fields holds information on each field in the record.
	// If it's shorter than the actual number of fields in the
	// associated struct type, additional zero-valued entries
	// are inferred as needed.
	Fields []FieldInfo
}

type FieldInfo struct {
	// Default returns the default value for the
	// type if any. If there is no default value
	// or the default value is the zero value, Default
	// will be nil.
	Default func() interface{}

	// Info holds information about the field type.
	// This is only set when the field is a union
	// so the information can't be inferred from the
	// field type itself.
	Info *TypeInfo
}

type TypeInfo struct {
	// Type holds a value of type *T where T is
	// the type described by the TypeInfo,
	// except when the TypeInfo represents the null
	// type, in which case Type will be nil.
	Type interface{}

	// When the TypeInfo describes a union,
	// Union holds an entry for each member
	// of the union.
	Union []TypeInfo
}

// azTypeInfo is the representation of the above types used
// by the analyzer. It can represent a record, a field or a type
// inside one of those.
type azTypeInfo struct {
	// ftype holds the Go type used for this Avro type.
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
			var fieldInfo FieldInfo
			if i < len(r.Fields) {
				fieldInfo = r.Fields[i]
			}
			entry, err := newAzTypeInfoFromField(t, t.Field(i).Type, fieldInfo)
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

func newAzTypeInfoFromField(refType, t reflect.Type, f FieldInfo) (azTypeInfo, error) {
	var info azTypeInfo
	if f.Info != nil {
		info1, err := newAzTypeInfoFromType(refType, *f.Info)
		if err != nil {
			return azTypeInfo{}, err
		}
		info = info1
	}
	info.ftype = t
	info.makeDefault = f.Default
	info.referenceType = refType
	return info, nil
}

func newAzTypeInfoFromType(refType reflect.Type, t TypeInfo) (azTypeInfo, error) {
	info := azTypeInfo{
		entries:       make([]azTypeInfo, len(t.Union)),
		referenceType: refType,
	}
	if t.Type != nil {
		info.ftype = reflect.TypeOf(t.Type).Elem()
	}
	if len(t.Union) == 0 {
		return info, nil
	}
	info.entries = make([]azTypeInfo, len(t.Union))
	for i, u := range t.Union {
		entry, err := newAzTypeInfoFromType(refType, u)
		if err != nil {
			return azTypeInfo{}, err
		}
		info.entries[i] = entry
	}
	return info, nil
}

const debugging = false

func debugf(f string, a ...interface{}) {
	if debugging {
		log.Printf(f, a...)
	}
}
