package avro

import (
	"fmt"
	"log"
	"reflect"

	"github.com/heetch/avro/avrotypegen"
)

// azTypeInfo is the representation of the above types used
// by the analyzer. It can represent a record, a field or a type
// inside one of those.
type azTypeInfo struct {
	// ftype holds the Go type used for this Avro type (or nil for null).
	ftype reflect.Type

	// fieldIndex holds the index of the field if this entry is about
	// a struct field.
	fieldIndex int

	// makeDefault is a function that returns the default
	// value for a field, or nil if there is no default value.
	makeDefault func() reflect.Value

	// isUnion holds whether this info is about a union type
	// (if not, it's about a struct).
	isUnion bool

	// Entries holds the possible types that can
	// be descended in from this type.
	// For structs (records) this holds an entry
	// for each field; for union types, this holds an
	// entry for each possible type in the union.
	entries []azTypeInfo
}

func newAzTypeInfo(t reflect.Type) (azTypeInfo, error) {
	debugf("azTypeInfo(%v)", t)
	switch t.Kind() {
	case reflect.Struct:
		info := azTypeInfo{
			ftype:   t,
			entries: make([]azTypeInfo, 0, t.NumField()),
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
			if name, _ := jsonFieldName(f); name == "" {
				continue
			}
			var required bool
			var makeDefault func() reflect.Value
			var unionInfo avrotypegen.UnionInfo
			if i < len(r.Required) {
				required = r.Required[i]
			}
			if i < len(r.Defaults) {
				if md := r.Defaults[i]; md != nil {
					makeDefault = func() reflect.Value {
						return reflect.ValueOf(md())
					}
				}
			}
			if i < len(r.Unions) {
				unionInfo = r.Unions[i]
			}
			entry := newAzTypeInfoFromField(f, required, makeDefault, unionInfo)
			info.entries = append(info.entries, entry)
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

func newAzTypeInfoFromField(f reflect.StructField, required bool, makeDefault func() reflect.Value, unionInfo avrotypegen.UnionInfo) azTypeInfo {
	t := f.Type
	if t.Kind() == reflect.Ptr && len(unionInfo.Union) == 0 {
		// It's a pointer but there's no explicit union entry, which means that
		// the union defaults to ["null", type]
		unionInfo.Union = []avrotypegen.UnionInfo{{
			Type: nil,
		}, {
			Type: reflect.New(t.Elem()).Interface(),
		}}
	}
	// Make an appropriate makeDefault function, even when one isn't explicitly specified.
	switch {
	case required:
		// Keep to the letter of the contract (makeDefault should always
		// be nil in this case anyway).
		makeDefault = nil
	case makeDefault == nil && len(unionInfo.Union) > 0:
		// It's a ["null", T] union - we can infer the default
		// value from the field type. The default value is the
		// zero value of the first member of the union.
		var v reflect.Value
		firstMemberType := unionInfo.Union[0].Type
		if firstMemberType != nil {
			v = reflect.Zero(reflect.TypeOf(firstMemberType).Elem())
		} else {
			v = reflect.Zero(t)
		}
		makeDefault = func() reflect.Value {
			return v
		}
	case makeDefault == nil:
		v := reflect.Zero(t)
		makeDefault = func() reflect.Value {
			return v
		}
	}
	info := azTypeInfo{
		ftype:       t,
		fieldIndex:  f.Index[0],
		makeDefault: makeDefault,
	}
	setUnionInfo(&info, unionInfo)
	return info
}

func setUnionInfo(info *azTypeInfo, unionInfo avrotypegen.UnionInfo) {
	if len(unionInfo.Union) == 0 {
		return
	}
	info.isUnion = true
	info.entries = make([]azTypeInfo, len(unionInfo.Union))
	for i, u := range unionInfo.Union {
		var ut reflect.Type
		if u.Type != nil {
			ut = reflect.TypeOf(u.Type).Elem()
		}
		info.entries[i] = azTypeInfo{
			ftype: ut,
		}
		setUnionInfo(&info.entries[i], u)
	}
}

const debugging = false

func debugf(f string, a ...interface{}) {
	if debugging {
		log.Printf(f, a...)
	}
}
