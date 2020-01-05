package avro

import (
	"fmt"
	"log"
	"reflect"

	"github.com/rogpeppe/avro/avrotypegen"
)

// azTypeInfo is the representation of the above types used
// by the analyzer. It can represent a record, a field or a type
// inside one of those.
type azTypeInfo struct {
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

func newAzTypeInfo(t reflect.Type) (azTypeInfo, error) {
	debugf("azTypeInfo(%v)", t)
	switch t.Kind() {
	case reflect.Struct:
		info := azTypeInfo{
			ftype:   t,
			entries: make([]azTypeInfo, t.NumField()),
		}
		// Note that RecordInfo is defined in such a way that
		// the zero value gives useful defaults for a normal Go
		// value that doesn't return the any RecordInfo - all
		// fields will default to their zero value and the only
		// unions will be pointer types.
		// We don't need to diagnose all bad Go types here - they'll
		// be caught earlier - when we try to determine the Avro schema
		// from the Go type.
		var r avrotypegen.RecordInfo
		if v, ok := reflect.Zero(t).Interface().(avrotypegen.AvroRecord); ok {
			r = v.AvroRecord()
		}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.Anonymous {
				// TODO consider struct embedding.
				return azTypeInfo{}, fmt.Errorf("anonymous fields not supported")
			}
			var required bool
			var makeDefault func() interface{}
			var unionVals []interface{}
			if i < len(r.Required) {
				required = r.Required[i]
			}
			if i < len(r.Defaults) {
				makeDefault = r.Defaults[i]
			}
			if i < len(r.Unions) {
				unionVals = r.Unions[i]
			}
			entry, err := newAzTypeInfoFromField(t, f.Type, required, makeDefault, unionVals)
			if err != nil {
				return azTypeInfo{}, err
			}
			info.entries[i] = entry
		}
		debugf("-> record, %d entries", len(info.entries))
		return info, nil
	default:
		// TODO check for top-level union types too.
		debugf("-> unknown")
		return azTypeInfo{
			ftype: t,
		}, nil
	}
}

func newAzTypeInfoFromField(refType, t reflect.Type, required bool, makeDefault func() interface{}, unionVals []interface{}) (azTypeInfo, error) {
	if t.Kind() == reflect.Ptr && len(unionVals) == 0 {
		// It's a pointer but there's no explicit union entry, which means that
		// the union defaults to ["null", type]
		unionVals = []interface{}{
			nil,
			reflect.New(t.Elem()).Interface(),
		}
	}
	// Make an appropriate makeDefault function, even when one isn't explicitly specified.
	switch {
	case required:
		// Keep to the letter of the contract.
		makeDefault = nil
	case makeDefault == nil && len(unionVals) > 0:
		var v interface{}
		if unionVals[0] != nil {
			v = reflect.Zero(reflect.TypeOf(unionVals[0]).Elem()).Interface()
		}
		makeDefault = func() interface{} {
			return v
		}
	case makeDefault == nil:
		v := reflect.Zero(t).Interface()
		makeDefault = func() interface{} {
			return v
		}
	}
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

const debugging = false

func debugf(f string, a ...interface{}) {
	if debugging {
		log.Printf(f, a...)
	}
}
