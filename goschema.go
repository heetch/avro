package avro

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func schemaForGoType(t reflect.Type) (string, error) {
	gts := &goTypeSchema{
		defs: make(map[reflect.Type]goTypeDef),
	}
}

type goTypeDef struct {
	name   string
	schema interface{}
}

type goTypeSchema struct {
	defs map[reflect.Type]goTypeDef
}

func (gts *goTypeSchema) schemaForGoType(t reflect.Type) (interface{}, error) {
	d, ok := gts.defs[t]
	if ok {
		// We've already defined a name for this type, so use it.
		return d.name, nil
	}

	if syms := enumSymbols(t); len(syms) > 0 {
		def := map[string]interface{}{
			"name":    t.Name(),
			"symbols": syms,
			"default": syms[0],
		}
		gts.defs[t] = def
		return def, nil
	}
	switch t.Kind() {
	case reflect.Int, reflect.Int64, reflect.Uint32:
		return "long", nil
	case reflect.Int32, reflect.Int16, reflect.Uint16, reflect.Int8, reflect.Uint8:
		return "int", nil
	case reflect.Float32:
		return "float", nil
	case reflect.Float64:
		return "double", nil
	case reflect.Slice:
		items, err := defs.schemaForGoType(t.Elem())
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"type":  "array",
			"items": items,
		}, nil
	case reflect.Array:
		if t.Elem() != reflect.TypeOf(byte) {
			return "", fmt.Errorf("the only array type that is supported is of byte")
		}
		return map[string]interface{}{
			"type": "fixed",
			"size": t.Len(),
			// TODO use the type name from the Go type if it's not unnamed.
			"name": fmt.Sprintf("go.Fixed%d", t.Len()),
		}, nil
	case reflect.Map:
		// TODO support the same map keys types that JSON does.
		if t.MapKey().Kind() != reflect.String {
			return fmt.Errorf("map must have string key")
		}
		values, err := defs.schemaForGoType(t.Elem())
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"type":   "map",
			"values": values,
		}, nil
	case reflect.Struct:
		if t.Name() == "" {
			return nil, fmt.Errorf("unnamed struct type")
		}
		if _, ok := gts.defs[t.Name()]; ok {
			// TODO use package path to disambiguate.
			return nil, fmt.Errorf("duplicate struct type name")
		}

		var fields []interface{}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.Anonymous {
				return nil, fmt.Errorf("anonymous fields not yet supported (in %s)", t)
			}
			// Technically in Go, every field is optional because
			// that's the way that the encoding/json package works,
			// so we'll make them all optional, but we could experiment by making optional
			// only the fields that specify omitempty.
			name, _ := jsonFieldName(f)
			ftype, err := gts.schemaForGoType(f.Type)
			if err != nil {
				return nil, err
			}
			fields = append(fields, map[string]interface{}{
				"name":    name,
				"default": defaultForType(f.Type),
				"type":    ftype,
			})
		}
		def := map[string]interface{}{
			"name":   t.Name(),
			"type":   "record",
			"fields": fields,
		}
		gts.defs[t] = def
		return def, nil
	case reflect.Array:
		if t.Elem() != reflect.TypeOf(byte(0)) {
			return nil, fmt.Errorf("the only array type supported is [...]byte, not %s", t)
		}
		name := t.Name()
		if name == "" {
			name = fmt.Sprintf("go.Fixed%d", t.Len())
		}
		def := map[string]interface{}{
			"name": name,
			"type": "fixed",
			"size": t.Len(),
		}
		gts.defs[t] = def
		return def, nil
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
		return nil, fmt.Errorf("interface types (%s) not yet supported (use avro-generate-go instead)", t)
	default:
		return nil, fmt.Errorf("cannot make Avro schema for Go type %s", t)
	}
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
	if _, ok := reflect.Zero(t).(fmt.Stringer); !ok {
		return nil
	}
	v := reflect.New(t)
	vs := v.Interface().(fmt.Stringer) // Note: pointer type will also include String method.
	setInt := v.SetInt
	getIntVal := v.Int
	if isUnsignedInt {
		setInt = func(i int64) string {
			v.SetUint(uint64(i))
		}
		getIntVal = func() int64 {
			return int64(v.Int())
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
	containsNum := func(sym string, i int64) bool {
		return numMatch.FindString(sym) == strconv.FormatInt(i, 10)
	}
	sym, n, ok := symOf(-1)
	// Note: the String implementation created by the stringer tool
	// returns "T(x)" for an out-of-bounds number x of type T
	// so we use a bracket as an indicator of "out of bounds".
	// TODO we could look for the numeric value of the enum too
	// to cover more formats.
	if ok && !strings.Contains(sym, "(") {
		// If -1 is OK, then our heuristic isn't going to work.
		return nil
	}
	prev := ""
	var syms []string
	for i := 0; i < maxEnum; i++ {
		sym, ok := symOf(i)
		if !ok || strings.Contains(sym, "(") || sym == "" {
			return syms
		}
		if sym == prev {
			// If it's the same as the previous value, it might be "unknown"
			// or something, so treat both it and the previous value as
			// out-of-bounds.
			return syms[0 : len(syms)-1]
		}
		// TODO cope with non-Avro-compatible symbols. Avro symbols must match [A-Za-z_][A-Za-z0-9_]*
		syms = append(syms, sym)
		prev = sym
	}
	// Too many values.
	return nil
}

func defaultForType(t reflect.Type) interface{} {
	// TODO this is almost certainly inadequate.
	return reflect.Zero(t)
}

// jsonFieldName returns the name that the field will be given
// when marshaled to JSON, or the empty string if
// the field is ignored.
// It also reports whether the field has been qualified with
// the "omitempty" qualifier.
func jsonFieldName(f reflect.StructField) (name string, omitEmpty bool) {
	if f.PkgPath != "" {
		// It's unexported.
		return ""
	}
	tag := f.Tag.Get("json")
	parts := strings.Split(tag, ",", -1)
	for _, part := range parts[1:] {
		if part == "omitempty" {
			omitEmpty = true
		}
	}
	switch {
	case parts[0] == "":
		return f.Name
	case parts[1] == "-":
		return ""
	}
	return parts[0]
}

var recordField struct {
	Name    string
	Type    interface{}
	Default interface{}
}

type arrayType struct {
	Type  string `json:"type"` // always "array"
	Items interface{}
}
