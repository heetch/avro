package avro

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/actgardner/gogen-avro/v10/schema"

	"github.com/heetch/avro/avrotypegen"
	"github.com/heetch/avro/internal/typeinfo"
)

const (
	timestampMicros = "timestamp-micros"
	timestampMillis = "timestamp-millis"
	uuid            = "uuid"
)

// globalNames holds the default namespace which maps all Go types
// to their Go names.
var globalNames = new(Names)

// errorSchema is a hack - it pretends to be an AvroType
// so that it can be held as a schema map value.
//
// In fact it just holds an error so that we can cache errors.
type errorSchema struct {
	schema.AvroType
	err error
}

// TypeOf returns the Avro type for the Go type of x.
//
// If the type was generated by avrogo, the returned schema
// will be the same as the schema it was generated from.
//
// Otherwise TypeOf(T) is derived according to
// the following rules:
//
//	- int, int64 and uint32 encode as "long"
//	- int32, int16, uint16, int8 and uint8 encode as "int"
//	- float32 encodes as "float"
//	- float64 encodes as "double"
//	- string encodes as "string"
//	- Null{} encodes as "null"
//	- time.Time encodes as {"type": "long", "logicalType": "timestamp-micros"}
//	- [N]byte encodes as {"type": "fixed", "name": "go.FixedN", "size": N}
//	- a named type with underlying type [N]byte encodes as [N]byte but typeName(T) for the name.
//	- []T encodes as {"type": "array", "items": TypeOf(T)}
//	- map[string]T encodes as {"type": "map", "values": TypeOf(T)}
//	- *T encodes as ["null", TypeOf(T)]
//	- a named struct type encodes as {"type": "record", "name": typeName(T), "fields": ...}
//		where the fields are encoded as described below.
//	- interface types are disallowed.
//
// Struct fields are encoded as follows:
//
//	- unexported struct fields are ignored
//	- the field name is taken from the Go field name, or from a "json" tag for the field if present.
//	- the default value for the field is the zero value for the type.
//	- anonymous struct fields are disallowed (this restriction may be lifted in the future).
func TypeOf(x interface{}) (*Type, error) {
	return globalNames.TypeOf(x)
}

func avroTypeOf(names *Names, t reflect.Type) (*Type, error) {
	rType0, ok := names.goTypeToAvroType.Load(t)
	if ok {
		rType := rType0.(*Type)
		if es, ok := rType.avroType.(errorSchema); ok {
			return nil, es.err
		}
		return rType, nil
	}
	rType, err := avroTypeOfUncached(names, t)
	if err != nil {
		names.goTypeToAvroType.LoadOrStore(t, &Type{
			avroType: errorSchema{err: err},
		})
		return nil, err
	}
	names.goTypeToAvroType.LoadOrStore(t, rType)
	return rType, nil
}

func avroTypeOfUncached(names *Names, t reflect.Type) (*Type, error) {
	gts := &goTypeSchema{
		names: names,
		defs:  make(map[reflect.Type]goTypeDef),
	}
	// TODO pass in wType so that we can determine a schema
	// even for partially specified Go types (e.g. interface{} values)
	// See https://github.com/heetch/avro/issues/34
	schemaVal, err := gts.schemaForGoType(t)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(schemaVal)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal generated schema: %v", err)
	}
	at, err := ParseType(string(data))
	if err != nil {
		return nil, err
	}
	if len(names.renames) == 0 {
		// There are no renames, so we don't need to rename and parse again.
		return at, nil
	}
	data1, err := json.Marshal(names.renameSchema(at.avroType))
	if err != nil {
		return nil, fmt.Errorf("cannot marshal generated renamed schema: %v", err)
	}
	at, err = ParseType(string(data1))
	if err != nil {
		return nil, fmt.Errorf("cannot parse generated renamed type: %v", err)
	}
	return at, nil
}

type goTypeDef struct {
	// name holds the Avro name for the Go type.
	name   string
	// schema holds the JSON-marshalable schema for the type.
	schema interface{}
}

// goTypeSchema holds execution context for the schemaForGoType
// functionality (to avoid passing both arguments everywhere).
type goTypeSchema struct {
	names *Names
	// defs maps from Go type to Avro definition for all
	// types being traversed by schemaForGoType..
	defs  map[reflect.Type]goTypeDef
}

func (gts *goTypeSchema) schemaForGoType(t reflect.Type) (interface{}, error) {
	d, ok := gts.defs[t]
	if ok {
		// We've already defined a name for this type, so use it.
		return d.name, nil
	}
	if t == nil {
		return "null", nil
	}
	if r := avroRecordOf(t); r != nil {
		// It's a generated type which comes with its own schema.
		return gts.define(t, json.RawMessage(r.AvroRecord().Schema), "")
	}
	if syms := enumSymbols(t); len(syms) > 0 {
		// It looks like an enum.
		// TODO should we include a default here?
		return gts.define(t, map[string]interface{}{
			"type":    "enum",
			"symbols": syms,
		}, "")
	}
	switch t.Kind() {
	case reflect.Bool:
		return "boolean", nil
	case reflect.String:
		return "string", nil
	case reflect.Int, reflect.Int64, reflect.Uint32:
		return "long", nil
	case reflect.Int32, reflect.Int16, reflect.Uint16, reflect.Int8, reflect.Uint8:
		return "int", nil
	case reflect.Float32:
		return "float", nil
	case reflect.Float64:
		return "double", nil
	case reflect.Slice:
		if t.Elem() == byteType {
			return "bytes", nil
		}
		items, err := gts.schemaForGoType(t.Elem())
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"type":  "array",
			"items": items,
		}, nil
	case reflect.Map:
		// TODO support the same map keys types that JSON does.
		if t.Key().Kind() != reflect.String {
			return nil, fmt.Errorf("map must have string key")
		}
		values, err := gts.schemaForGoType(t.Elem())
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"type":   "map",
			"values": values,
		}, nil
	case reflect.Struct:
		switch t {
		case timeType:
			return map[string]interface{}{
				"type":        "long",
				"logicalType": timestampMicros,
			}, nil
		case nullType:
			return "null", nil
		}
		// Define the struct type before filling in the definition
		// so that we'll find the definition if there's a recursive type.
		// The map returned by the define method holds a reference
		// to the same object held in gts.defs, so changing it
		// below will update the final definition.
		def, err := gts.define(t, map[string]interface{}{
			"type": "record",
		}, "")
		if err != nil {
			return nil, err
		}
		// Note: don't start with nil fields because gogen-avro
		// doesn't like the nil value.
		fields := []interface{}{}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.Anonymous {
				return nil, fmt.Errorf("anonymous fields not yet supported (in %s)", t)
			}
			// Technically in Go, every field is optional because
			// that's the way that the encoding/json package works,
			// so we'll make them all optional.
			// TODO  experiment by making optional only the fields that
			// specify omitempty.
			name, _ := typeinfo.JSONFieldName(f)
			if name == "" {
				continue
			}
			ftype, err := gts.schemaForGoType(f.Type)
			if err != nil {
				return nil, err
			}
			d, err := gts.defaultForType(f.Type)
			if err != nil {
				return nil, err
			}
			fields = append(fields, map[string]interface{}{
				"name":    name,
				"default": d,
				"type":    ftype,
			})
		}
		def["fields"] = fields
		return def, nil
	case reflect.Array:
		if t == uuidType {
			return map[string]interface{}{
				"type":        "string",
				"logicalType": uuid,
			}, nil
		}

		if t.Elem() != reflect.TypeOf(byte(0)) {
			return nil, fmt.Errorf("the only array type supported is [...]byte, not %s", t)
		}
		return gts.define(t, map[string]interface{}{
			"type": "fixed",
			"size": t.Len(),
		}, fmt.Sprintf("go.Fixed%d", t.Len()))
	case reflect.Ptr:
		if t.Elem().Kind() == reflect.Ptr {
			return nil, fmt.Errorf("can only cope with a single level of pointer indirection")
		}
		elem, err := gts.schemaForGoType(t.Elem())
		if err != nil {
			return nil, err
		}
		return []interface{}{
			"null",
			elem,
		}, nil
	case reflect.Interface:
		// TODO fill in from the writer schema.
		return nil, fmt.Errorf("interface types (%s) not yet supported (use avrogo instead)", t)
	default:
		return nil, fmt.Errorf("cannot make Avro schema for Go type %s", t)
	}
}

func (gts *goTypeSchema) define(t reflect.Type, def0 interface{}, defaultName string) (map[string]interface{}, error) {
	def, ok := def0.(map[string]interface{})
	if !ok {
		if err := json.Unmarshal(def0.(json.RawMessage), &def); err != nil {
			return nil, err
		}
	}
	name, _ := def["name"].(string)
	if name == "" {
		// TODO use a fully qualified name derived from the Go package path
		// as well as the type name. See https://github.com/heetch/avro/issues/35
		if name = t.Name(); name == "" {
			if name = defaultName; name == "" {
				return nil, fmt.Errorf("cannot use unnamed type %s as Avro type", t)
			}
		}
		def["name"] = name
	}
	for _, def := range gts.defs {
		if def.name == name {
			// TODO use package path to disambiguate. See https://github.com/heetch/avro/issues/35
			return nil, fmt.Errorf("duplicate struct type name %q", name)
		}
	}
	gts.defs[t] = goTypeDef{
		name:   name,
		schema: def,
	}
	return def, nil
}

const maxEnum = 250

// enumSymbols returns the enum symbols represented by the given
// type. If the type doesn't represent an enum it returns no symbols.
func enumSymbols(t reflect.Type) []string {
	k := t.Kind()
	isSignedInt := reflect.Int <= k && k <= reflect.Int64
	isUnsignedInt := reflect.Uint <= k && k <= reflect.Uint64
	if !isSignedInt && !isUnsignedInt {
		return nil
	}
	if _, ok := reflect.Zero(t).Interface().(fmt.Stringer); !ok {
		return nil
	}
	v := reflect.New(t)
	vs := v.Interface().(fmt.Stringer) // Note: pointer type will also include String method.
	v = v.Elem()
	setInt := v.SetInt
	getIntVal := v.Int
	if isUnsignedInt {
		setInt = func(i int64) {
			v.SetUint(uint64(i))
		}
		getIntVal = func() int64 {
			return int64(v.Uint())
		}
	}
	symOf := func(i int64) (sym string, actual int64, ok bool) {
		defer func() {
			// It panics when calling String, which is a decent indication
			// that it's out of bounds.
			if recover() != nil {
				ok = false
			}
		}()
		setInt(i)
		return vs.String(), getIntVal(), true
	}
	// Assume that -1 is out-of-bounds and see what
	// we get when we call String on it.
	sym, actual, ok := symOf(-1)
	const (
		oobEmpty = iota
		oobParen
		oobNumber
		oobPanic
	)
	var oobStyle int
	// Note: the String implementation created by the stringer tool
	// returns "T(x)" for an out-of-bounds number x of type T
	// so we use a bracket as an indicator of "out of bounds".
	switch {
	case !ok:
		oobStyle = oobPanic
	case sym == "":
		oobStyle = oobEmpty
	case strings.Contains(sym, "("):
		oobStyle = oobParen
	case sym == fmt.Sprint(actual):
		oobStyle = oobNumber
	default:
		// All our heuristics for detecting out-of-bounds values
		// are exhausted.
		return nil
	}
	prev := ""
	var syms []string
	for i := 0; i < maxEnum; i++ {
		sym, actual, ok := symOf(int64(i))
		if !ok || sym == "" {
			// Panic or empty value are never acceptable.
			return syms
		}
		switch oobStyle {
		case oobParen:
			if strings.Contains(sym, "(") {
				return syms
			}
		case oobNumber:
			if sym == fmt.Sprint(actual) {
				return syms
			}
		}
		if sym == prev {
			// If it's the same as the previous value, it might be "unknown"
			// or something, so treat both it and the previous value as
			// out-of-bounds.
			return syms[0 : len(syms)-1]
		}
		if !isValidEnumSymbol(sym) {
			// TODO convert to a valid symbol somehow?
			// https://github.com/heetch/avro/issues/36
			return nil
		}
		syms = append(syms, sym)
		prev = sym
	}
	// Too many values.
	return nil
}

// From https://avro.apache.org/docs/1.9.1/spec.html#Enums :
//
//	Every symbol must match the regular expression [A-Za-z_][A-Za-z0-9_]*
func isValidEnumSymbol(s string) bool {
	if s == "" || s[0] != '_' && !isAlpha(s[0]) {
		return false
	}
	for i := 1; i < len(s); i++ {
		if c := s[i]; c != '_' && !isAlpha(c) && !isDigit(c) {
			return false
		}
	}
	return true
}

func isAlpha(c byte) bool {
	return 'A' <= c && c <= 'Z' || 'a' <= c && c <= 'z'
}

func isDigit(c byte) bool {
	return '0' <= c && c <= '9'
}

func (gts *goTypeSchema) defaultForType(t reflect.Type) (interface{}, error) {
	// TODO perhaps a Go slice/map should accept a union
	// of null and array/map? See https://github.com/heetch/avro/issues/19
	switch t.Kind() {
	case reflect.Slice:
		return reflect.MakeSlice(t, 0, 0).Interface(), nil
	case reflect.Map:
		return reflect.MakeMap(t).Interface(), nil
	case reflect.Array:
		if t == uuidType {
			return "", nil
		}
		return strings.Repeat("\u0000", t.Len()), nil
	case reflect.Struct:
		switch t {
		case timeType:
			return 0, nil
		case nullType:
			return nil, nil
		}
		if avroRecordOf(t) != nil {
			// It's a generated type - producing a correctly formed default value
			// for it needs a bit more work so we punt on doing it for now.
			// TODO make default values for struct-typed fields work in all cases.
			return nil, fmt.Errorf("value fields of struct types generated by avrogo are not yet supported (type %s)", t)
		}
		fields := make(map[string]interface{})
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.Anonymous {
				return nil, fmt.Errorf("anonymous fields not yet supported (in %s)", t)
			}
			name, _ := typeinfo.JSONFieldName(f)
			if name == "" {
				continue
			}
			v, err := gts.defaultForType(f.Type)
			if err != nil {
				return nil, err
			}
			fields[name] = v
		}
		return fields, nil
	default:
		if def, ok := gts.defs[t]; ok {
			if o, ok := def.schema.(map[string]interface{}); ok && o["type"] == "enum" {
				return reflect.Zero(t).Interface().(fmt.Stringer).String(), nil
			}
		}
		return reflect.Zero(t).Interface(), nil
	}
}

func avroRecordOf(t reflect.Type) avrotypegen.AvroRecord {
	r, _ := reflect.Zero(t).Interface().(avrotypegen.AvroRecord)
	return r
}

var nullType = reflect.TypeOf(Null{})

// Null represents the Avro null type. Its only JSON representation is null.
type Null = avrotypegen.Null
