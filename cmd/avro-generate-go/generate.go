package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/rogpeppe/gogen-avro/v7/parser"
	"github.com/rogpeppe/gogen-avro/v7/resolver"
	"github.com/rogpeppe/gogen-avro/v7/schema"
)

const (
	timestampMicros = "timestamp-micros"
	timestampMillis = "timestamp-millis"
)

func generate(w io.Writer, s []byte, pkg string) error {
	ns := parser.NewNamespace(false)
	sType, err := parseSchema(ns, s)
	if err != nil {
		return err
	}
	if _, ok := sType.(*schema.Reference); !ok {
		// The schema doesn't have a top-level name.
		// TODO how should we cope with a schema that's not
		// a definition? In that case we don't have
		// a name for the type, and we may not be able to define
		// methods on it because it might be a union type which
		// is represented by an interface type in Go.
		// See https://github.com/heetch/avro/issues/13
		return fmt.Errorf("cannot generate code for a schema which hasn't got a name (%T)", sType)
	}
	imports = make(map[string]bool)
	addImport("github.com/heetch/avro/avrotypegen")
	var body bytes.Buffer
	if err := genTemplate.Execute(&body, templateParams{
		NS: ns,
	}); err != nil {
		return err
	}
	var importList []string
	for imp := range imports {
		importList = append(importList, imp)
	}
	sort.Strings(importList)
	if err := headerTemplate.Execute(w, headerTemplateParams{
		Pkg:     pkg,
		Imports: importList,
	}); err != nil {
		return fmt.Errorf("cannot execute header template: %v", err)
	}
	if _, err := w.Write(body.Bytes()); err != nil {
		return err
	}
	return nil
}

func parseSchema(ns *parser.Namespace, s []byte) (schema.AvroType, error) {
	avroType, err := ns.TypeForSchema(s)
	if err != nil {
		return nil, err
	}
	for _, def := range ns.Roots {
		if err := resolver.ResolveDefinition(def, ns.Definitions); err != nil {
			return nil, fmt.Errorf("cannot resolve references in schema: %v", err)
		}
	}
	return avroType, nil
}

type typeInfo struct {
	// GoType holds the name of the type used
	// in Go. The "null" type is represented by
	// the string "nil".
	GoType string

	// Union holds type info for all the members of a union.
	Union []typeInfo
}

func (info typeInfo) Doc() string {
	var buf strings.Builder
	writeUnionComment(&buf, info.Union, "")
	return buf.String()
}

func fprintf(w io.Writer, f string, a ...interface{}) {
	fmt.Fprintf(w, f, a...)
}

func recordInfoLiteral(t *schema.RecordDefinition) string {
	w := new(strings.Builder)
	fprintf(w, "avrotypegen.RecordInfo{\n")
	schemaStr, err := t.Schema()
	if err != nil {
		panic(err)
	}
	fprintf(w, "Schema: %s,\n", quote(schemaStr))
	doneRequired := false
	for i, f := range t.Fields() {
		if f.HasDefault() {
			continue
		}
		if !doneRequired {
			fprintf(w, "Required: []bool{\n")
			doneRequired = true
		}
		fprintf(w, "%d: %v,\n", i, true)
	}
	if doneRequired {
		fprintf(w, "},\n")
	}

	doneDefaults := false
	for i, f := range t.Fields() {
		if !f.HasDefault() || isZeroDefault(f.Default(), f.Type()) {
			continue
		}
		if !doneDefaults {
			fprintf(w, "Defaults: []func() interface{}{\n")
			doneDefaults = true
		}
		fprintf(w, "%d: ", i)
		lit, err := defaultFuncLiteral(f.Default(), f.Type())
		if err != nil {
			fprintf(w, "func() interface{} {}, // ERROR: %v\n", err)
		} else {
			fprintf(w, "func() interface{} {\nreturn %s\n},\n", lit)
		}
	}
	if doneDefaults {
		fprintf(w, "},\n")
	}

	doneUnions := false
	for i, f := range t.Fields() {
		info := goType(f.Type())
		if canOmitUnionInfo(info) {
			continue
		}
		if !doneUnions {
			fprintf(w, "Unions: []avrotypegen.UnionInfo{\n")
			doneUnions = true
		}
		fprintf(w, "%d: ", i)
		writeUnionInfo(w, info)
		fprintf(w, ",\n")
	}
	if doneUnions {
		fprintf(w, "},\n")
	}
	fprintf(w, "}")
	return w.String()
}

// canOmitUnionInfo reports whether the info for the
// given union can be omitted from the UnionInfo.
func canOmitUnionInfo(u typeInfo) bool {
	// Check that either there's no union or the union is ["null", T]
	// (the default union type for a pointer) and the Go type is also
	// a pointer, meaning the avro package can infer that it's a
	// pointer union.
	return len(u.Union) == 0 || (len(u.Union) == 2 && u.Union[0].GoType == "nil" && u.GoType[0] == '*')
}

func writeUnionInfo(w io.Writer, info typeInfo) {
	fprintf(w, "{\n")
	if info.GoType == "nil" {
		// Technically we could omit this, but it
		// looks nicer if we don't.
		fprintf(w, "Type: nil,\n")
	} else {
		fprintf(w, "Type: new(%s),\n", info.GoType)
	}
	if len(info.Union) > 0 {
		fprintf(w, "Union: []avrotypegen.UnionInfo{")
		for i, u := range info.Union {
			if i > 0 {
				fprintf(w, ", ")
			}
			writeUnionInfo(w, u)
		}
		fprintf(w, "},\n")
	}
	fprintf(w, "}")
}

// isZeroDefault reports whether x is the zero default value of type t.
func isZeroDefault(x interface{}, t schema.AvroType) bool {
	switch t := t.(type) {
	case *schema.UnionField:
		// Defaults for unions fields use the first member of the union.
		return isZeroDefault(x, t.AvroTypes()[0])
	case *schema.NullField:
		return x == nil
	case *schema.BoolField:
		return x == false
	case *schema.IntField,
		*schema.LongField,
		*schema.FloatField,
		*schema.DoubleField:
		return x == float64(0)
	case *schema.BytesField,
		*schema.StringField:
		return x == ""
	case *schema.ArrayField:
		x, ok := x.([]interface{})
		return ok && len(x) == 0
	case *schema.MapField:
		x, ok := x.(map[string]interface{})
		return ok && len(x) == 0
	case *schema.Reference:
		switch def := t.Def.(type) {
		case *schema.EnumDefinition:
			s, ok := x.(string)
			syms := def.Symbols()
			return ok && len(syms) > 0 && s == syms[0]
		case *schema.FixedDefinition:
			s, ok := x.(string)
			return ok && s == strings.Repeat(string(0), def.SizeBytes())
		case *schema.RecordDefinition:
			m, ok := x.(map[string]interface{})
			if !ok {
				return false
			}
			for _, field := range def.Fields() {
				f, ok := m[field.Name()]
				if !ok || !isZeroDefault(f, field.Type()) {
					return false
				}
			}
			return true
		}
	}
	return false
}

func jsonMarshal(x interface{}) string {
	data, err := json.Marshal(x)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func defaultFuncLiteral(v interface{}, t schema.AvroType) (string, error) {
	switch t := t.(type) {
	case *schema.UnionField:
		// Defaults for unions fields always use the first member
		// of the union.
		return defaultFuncLiteral(v, t.AvroTypes()[0])
	case *schema.NullField:
		if v != nil {
			return "", fmt.Errorf("must be null but got %s", jsonMarshal(v))
		}
		return "nil", nil
	case *schema.BoolField:
		v, ok := v.(bool)
		if !ok {
			return "", fmt.Errorf("must be boolean but got %s", jsonMarshal(v))
		}
		return fmt.Sprintf("%v", v), nil
	case *schema.IntField:
		return numberDefault(v, "int")
	case *schema.LongField:
		return numberDefault(v, "int64")
	case *schema.FloatField:
		return numberDefault(v, "float32")
	case *schema.DoubleField:
		return numberDefault(v, "float64")
	case *schema.BytesField:
		s, ok := v.(string)
		if !ok {
			return "", fmt.Errorf("must be string but got %v", jsonMarshal(v))
		}
		bytes, err := decodeBytes(s)
		if err != nil {
			return "", fmt.Errorf("cannot decode bytes literal %v: %v", jsonMarshal(v), err)
		}
		return fmt.Sprintf("[]byte(%q)", bytes), nil
	case *schema.StringField:
		s, ok := v.(string)
		if !ok {
			return "", fmt.Errorf("must be string but got %v", jsonMarshal(v))
		}
		return fmt.Sprintf("%q", s), nil
	case *schema.ArrayField:
		a, ok := v.([]interface{})
		if !ok {
			return "", fmt.Errorf("must be array but got %v", jsonMarshal(v))
		}
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "[]%s{", goType(t.ItemType()).GoType)
		for i, item := range a {
			val, err := defaultFuncLiteral(item, t.ItemType())
			if err != nil {
				return "", fmt.Errorf("at index %d: %v", i, err)
			}
			buf.WriteString(val)
			buf.WriteString(",")
		}
		buf.WriteString("}")
		return buf.String(), nil
	case *schema.MapField:
		m, ok := v.(map[string]interface{})
		if !ok {
			return "", fmt.Errorf("must be map but got %v", jsonMarshal(v))
		}
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "map[string]%s{\n", goType(t.ItemType()).GoType)
		var keys []string
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, key := range keys {
			val, err := defaultFuncLiteral(m[key], t.ItemType())
			if err != nil {
				return "", fmt.Errorf("at key %q: %v", key, err)
			}
			fmt.Fprintf(&buf, "%q: %s,\n", key, val)
		}
		buf.WriteString("}")
		return buf.String(), nil
	case *schema.Reference:
		switch def := t.Def.(type) {
		case *schema.EnumDefinition:
			s, ok := v.(string)
			if !ok {
				return "", fmt.Errorf("enum default value must be string, not %s", jsonMarshal(v))
			}
			for _, sym := range def.Symbols() {
				if sym == s {
					return def.SymbolName(s), nil
				}
			}
			return "", fmt.Errorf("unknown value %q for enum %s", s, def.Name())
		case *schema.FixedDefinition:
			s, ok := v.(string)
			if !ok {
				return "", fmt.Errorf("fixed default value must be string, not %s", jsonMarshal(v))
			}
			b, err := decodeBytes(s)
			if err != nil {
				return "", fmt.Errorf("invalid fixed default value %q: %v", b, err)
			}
			if len(b) != def.SizeBytes() {
				return "", fmt.Errorf("fixed value %s is wrong length (got %d; want %d)", jsonMarshal(v), len(b), def.SizeBytes())
			}
			var buf bytes.Buffer
			fmt.Fprintf(&buf, "%s{", def.Name())
			for _, x := range b {
				fmt.Fprintf(&buf, "%#x, ", x)
			}
			fmt.Fprintf(&buf, "}")
			return buf.String(), nil
		case *schema.RecordDefinition:
			m, ok := v.(map[string]interface{})
			if !ok {
				return "", fmt.Errorf("invalid record default value %s", jsonMarshal(v))
			}
			var buf bytes.Buffer
			fmt.Fprintf(&buf, "%s{\n", def.Name())
			for _, field := range def.Fields() {
				fieldVal, ok := m[field.Name()]
				var lit string
				if !ok {
					if !field.HasDefault() {
						return "", fmt.Errorf("field %q not present", field.Name())
					}
					fieldVal = field.Default()
				}
				lit, err := defaultFuncLiteral(fieldVal, field.Type())
				if err != nil {
					return "", fmt.Errorf("at field %s: %v", field.Name(), err)
				}
				fmt.Fprintf(&buf, "%s: %s,\n", field.GoName(), lit)
			}
			buf.WriteString("}")
			return buf.String(), nil
		default:
			return "", fmt.Errorf("unknown definition type %T", def)
		}

	default:
		return "", fmt.Errorf("literal of type %T not yet implemented", t)
	}
}

func decodeBytes(s string) ([]byte, error) {
	b := make([]byte, 0, len(s))
	for _, r := range s {
		if r > 0xff {
			return nil, fmt.Errorf("rune out of range (%d) in byte literal %q", r, s)
		}
		b = append(b, byte(r))
	}
	return b, nil
}

func numberDefault(v interface{}, goType string) (string, error) {
	switch v := v.(type) {
	case float64:
		s := fmt.Sprintf("%v", v)
		switch {
		case goType == "int" && isValidInt(s):
			return s, nil
		case goType == "float64" && !isValidInt(s):
			return s, nil
		}
		// TODO omit type conversion when it's not needed?
		return fmt.Sprintf("%s(%v)", goType, v), nil
	default:
		return "", fmt.Errorf("must be number but got %T", v)
	}
}

func isValidInt(s string) bool {
	_, err := strconv.ParseInt(s, 10, 64)
	return err == nil
}

func writeUnionComment(w io.Writer, union []typeInfo, indent string) {
	if len(union) == 0 {
		return
	}
	if len(union) == 2 && (union[0].GoType == "nil" || union[1].GoType == "nil") {
		// No need to comment a nil union.
		// TODO we may want to document whether a map or array may
		// be nil though. https://github.com/heetch/avro/issues/19
		return
	}
	printf := func(a string, f ...interface{}) {
		fmt.Fprintf(w, indent+a, f...)
	}
	printf("Allowed types for interface{} value:\n")
	for _, t := range union {
		printf("\t%s\n", t.GoType)
		writeUnionComment(w, t.Union, indent+"\t")
	}
}

func goType(t schema.AvroType) typeInfo {
	var info typeInfo
	switch t := t.(type) {
	case *schema.NullField:
		info.GoType = "nil"
	case *schema.BoolField:
		info.GoType = "bool"
	case *schema.IntField:
		// Note: Go int is at least 32 bits.
		info.GoType = "int"
	case *schema.LongField:
		// TODO support timestampMillis. https://github.com/heetch/avro/issues/3
		if logicalType(t) == timestampMicros {
			info.GoType = "time.Time"
			addImport("time")
		} else {
			info.GoType = "int64"
		}
	case *schema.FloatField:
		info.GoType = "float32"
	case *schema.DoubleField:
		info.GoType = "float64"
	case *schema.BytesField:
		info.GoType = "[]byte"
	case *schema.StringField:
		info.GoType = "string"
	case *schema.UnionField:
		types := t.AvroTypes()
		switch {
		case len(types) == 2 && isNullField(types[0]):
			// TODO if inner type is array or map, we don't need
			// the pointer - both of those types already have nil
			// values in Go.
			// https://github.com/heetch/avro/issues/19
			inner := goType(types[1])
			info.GoType = "*" + inner.GoType
			info.Union = []typeInfo{
				{
					GoType: "nil",
				},
				inner,
			}
		case len(types) == 2 && isNullField(types[1]):
			inner := goType(types[0])
			info.GoType = "*" + inner.GoType
			info.Union = []typeInfo{
				inner,
				{
					GoType: "nil",
				},
			}
		default:
			info.GoType = "interface{}"
			info.Union = make([]typeInfo, len(types))
			for i, t := range types {
				info.Union[i] = goType(t)
			}
		}
	case *schema.ArrayField:
		inner := goType(t.ItemType())
		info.GoType = "[]" + inner.GoType
		info.Union = inner.Union
	case *schema.MapField:
		inner := goType(t.ItemType())
		info.GoType = "map[string]" + inner.GoType
		info.Union = inner.Union
	case *schema.Reference:
		// TODO this is wrong! SimpleName might be unexported.
		info.GoType = t.SimpleName()
	default:
		panic(fmt.Sprintf("unknown avro type %T", t))
	}
	return info
}

func isNullField(t schema.AvroType) bool {
	_, ok := t.(*schema.NullField)
	return ok
}

func logicalType(t schema.AvroType) string {
	// Until https://github.com/actgardner/gogen-avro/issues/119
	// is fixed, we can't access metadata in general without a
	// race condition, so implement logicalType only
	// for the types that we currently care about, which
	// don't mutate themselves when Definition is called.
	switch t := t.(type) {
	case *schema.LongField, *schema.IntField:
		defn, _ := t.Definition(nil)
		defn1, _ := defn.(map[string]interface{})
		lt, _ := defn1["logicalType"].(string)
		return lt
	}
	return ""
}

var imports map[string]bool

// addImport adds a package to the required imports.
// This is seriously sleazy, but easy to do for the time being
// without refactoring the way that templates work.
// TODO avoid the global mutable variable.
func addImport(pkg string) {
	imports[pkg] = true
}
