package main

import (
	"bytes"
	"encoding/json"
	stdflag "flag"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/heetch/avro/internal/typeinfo"

	"github.com/actgardner/gogen-avro/v10/schema"
)

var flag = stdflag.NewFlagSet("", stdflag.ContinueOnError)

var outFile = flag.String("o", "", "output filename (default stdout)")

func main() {
	os.Exit(main1())
}

// main1 is the internal version of main that returns a status
// code instead of calling os.Exit.
func main1() int {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: avsc2avdl file.avdl\n")
		flag.PrintDefaults()
	}
	if flag.Parse(os.Args[1:]) != nil {
		return 2
	}
	if flag.NArg() != 1 {
		flag.Usage()
		return 2
	}
	if err := avsc2avdl(flag.Arg(0), *outFile); err != nil {
		fmt.Fprintf(os.Stderr, "avsc2avdl: %v\n", err)
		return 1
	}
	return 0
}

// avsc2avdl converts the AVSC data in the file avscFile and writes
// it to outFile (or stdout if outFile is empty).
func avsc2avdl(avscFile, outFile string) error {
	data, err := ioutil.ReadFile(avscFile)
	if err != nil {
		return err
	}
	at, err := typeinfo.ParseSchema(string(data), nil)
	if err != nil {
		return fmt.Errorf("cannot parse schema from %q: %v", avscFile, err)
	}
	ref, ok := at.(*schema.Reference)
	if !ok {
		return fmt.Errorf("top level of schema is not a reference")
	}
	g := &generator{
		filename: outFile,
		line:     1,
		done:     make(map[schema.QualifiedName]bool),
	}
	g.addDefinition(ref.Def)
	// Use the top level definition's namespace as the default namespace for
	// all definitions.
	namespace := ref.Def.AvroName().Namespace
	g.pushNamespace(namespace)
	if namespace != "" {
		g.printf("@namespace(%q)\n", namespace)
	}
	g.printf("protocol _ {\n")
	for i := 0; ; i++ {
		def := g.removeDefinition()
		if def == nil {
			break
		}
		if i > 0 {
			g.printf("\n")
		}
		g.writeDefinition(def)
	}
	g.printf("}\n")
	if outFile == "" {
		os.Stdout.Write(g.buf.Bytes())
	} else {
		err := ioutil.WriteFile(outFile, g.buf.Bytes(), 0666)
		if err != nil {
			return fmt.Errorf("cannot create output file: %v", err)
		}
	}
	return nil
}

type generator struct {
	filename       string
	line           int
	buf            bytes.Buffer
	queue          []schema.Definition
	namespaceStack []string
	done           map[schema.QualifiedName]bool
}

func (g *generator) writeDefinition(def schema.Definition) {
	name := def.AvroName()
	writeNamespace := func() {}
	if name.Namespace != "" && name.Namespace != g.namespace() {
		writeNamespace = func() {
			g.printf("\t@namespace(%q)\n", name.Namespace)
		}
		g.pushNamespace(name.Namespace)
		defer g.popNamespace()
	}
	switch def := def.(type) {
	case *schema.RecordDefinition:
		g.writeMetadata(def, "\t")
		writeNamespace()
		g.printf("\trecord %s {\n", name.Name)
		for _, field := range def.Fields() {
			g.writeMetadata(field, "\t\t")
			g.printf("\t\t%s %s", g.typeString(field.Type()), field.Name())
			if field.HasDefault() {
				switch {
				case isEnum(field.Type()):
					g.warningf("default value (%q) for enum-valued field in %s.%s will be ignored (see https://issues.apache.org/jira/browse/AVRO-2866)",
						field.Default(),
						name,
						field.Name(),
					)
				case isRecord(field.Type()):
					g.warningf("default value (%s) for record-valued field in %s.%s will cause an exception (see https://issues.apache.org/jira/browse/AVRO-2867)",
						jsonMarshal(field.Default(), ""),
						name,
						field.Name(),
					)
				}
				g.printf(" = %s", jsonMarshal(field.Default(), "\t\t"))
			}
			g.printf(";\n")
		}
		g.printf("\t}\n")
	case *schema.EnumDefinition:
		g.writeMetadata(def, "\t")
		writeNamespace()
		g.printf("\tenum %s {\n", name.Name)
		syms := def.Symbols()
		for i, sym := range syms {
			g.printf("\t\t%s", sym)
			if i < len(syms)-1 {
				g.printf(",")
			}
			g.printf("\n")
		}
		// TODO support enum defaults (gogen-avro doesn't make it easy to get
		// them - it might not support them currently itself)
		g.printf("\t}\n")
	case *schema.FixedDefinition:
		g.writeMetadata(def, "\t")
		writeNamespace()
		g.printf("\tfixed %s(%d);\n", name.Name, def.SizeBytes())
	default:
		panic(fmt.Errorf("unknown definition type %T", def))
	}
}

func (g *generator) typeString(at schema.AvroType) string {
	// TODO support logical types that are recognised directly
	// by IDL (decimal, date, time_millis and timestamp_millis)
	switch at := at.(type) {
	case *schema.Reference:
		g.addDefinition(at.Def)
		if at.TypeName.Namespace != g.namespace() {
			return at.TypeName.String()
		}
		return at.TypeName.Name
	case *schema.NullField:
		return "null"
	case *schema.BoolField:
		return "boolean"
	case *schema.IntField:
		return "int"
	case *schema.LongField:
		return "long"
	case *schema.FloatField:
		return "float"
	case *schema.DoubleField:
		return "double"
	case *schema.BytesField:
		return "bytes"
	case *schema.StringField:
		return "string"
	case *schema.ArrayField:
		return "array<" + g.typeString(at.ItemType()) + ">"
	case *schema.MapField:
		return "map<" + g.typeString(at.ItemType()) + ">"
	case *schema.UnionField:
		types := at.ItemTypes()
		if len(types) > 2 {
			// It's a long union type; format the types on separate lines.
			s := "union {\n\t\t\t"
			for i, ut := range types {
				if i > 0 {
					// TODO it's possible to have nested union types.
					// To fix that we'll need to have indent as an argument
					// to typeString.
					s += ",\n\t\t\t"
				}
				s += g.typeString(ut)
			}
			s += "\n\t\t}"
			return s
		}
		s := "union { "
		for i, ut := range types {
			if i > 0 {
				s += ", "
			}
			s += g.typeString(ut)
		}
		s += " }"
		return s
	default:
		panic(fmt.Errorf("unknown Avro type %T", at))
	}
}

func (g *generator) writeMetadata(d interface{}, indent string) {
	m := getMetadata(d)
	if m.doc != "" {
		g.printf("%s", indent)
		if strings.Contains(m.doc, "\n") {
			// This is dead-code by now https://github.com/actgardner/gogen-avro/pull/177
			// as gogen-avro removes the whitespaces.
			// idl2schemata strips leading and trailing whitespace,
			// so put some back again if it's a multiline comment.
			g.printf("/**\n%s %s\n%s */\n", indent, m.doc, indent)
		} else {
			g.printf("/** %s */\n", m.doc)
		}
	}
	type attr struct {
		name string
		val  interface{}
	}
	attrs := make([]attr, 0, len(m.attrs))
	for name, val := range m.attrs {
		attrs = append(attrs, attr{name, val})
	}
	sort.Slice(attrs, func(i, j int) bool {
		return attrs[i].name < attrs[j].name
	})
	for _, attr := range attrs {
		g.printf("%s@%s(%s)\n", indent, attr.name, jsonMarshal(attr.val, indent))
	}
}

func (g *generator) printf(f string, a ...interface{}) {
	s := fmt.Errorf(f, a...).Error()
	g.line += strings.Count(s, "\n")
	g.buf.WriteString(s)
}

func (g *generator) warningf(f string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s: WARNING: %s\n", g.location(), fmt.Errorf(f, a...))
}

func (g *generator) location() string {
	return fmt.Errorf("%s:%d", g.filename, g.line).Error()
}

type metadata struct {
	doc   string
	attrs map[string]interface{}
}

func getMetadata(d interface{}) metadata {
	var m metadata
	switch d := d.(type) {
	case *schema.Field:
		// fields are slightly different from top level definitions,
		// but similar enough that it still seems worth sharing
		// the getMetadata functionality. The @ attributes
		// are associated with the top level type, not the field
		// itself. See also https://issues.apache.org/jira/browse/AVRO-286

		// If the field is a reference to a named type, there
		// can be no metadata.
		if _, ok := d.Type().(*schema.Reference); !ok {
			m = getMetadata(d.Type())
		}
		if m.attrs == nil {
			m.attrs = make(map[string]interface{})
		}
		m.attrs["doc"] = d.Doc()
	case interface {
		Definition(scope map[schema.QualifiedName]interface{}) (interface{}, error)
	}:
		def, _ := d.Definition(make(map[schema.QualifiedName]interface{}))
		m.attrs, _ = def.(map[string]interface{})
	default:
		panic(fmt.Errorf("invalid type %T for definitionOf", d))
	}
	m.doc, _ = m.attrs["doc"].(string)
	delete(m.attrs, "doc")
	delete(m.attrs, "type")
	// Remove attributes that we don't want to treat as annotations.
	for _, attr := range stdAttrs[reflect.TypeOf(d)] {
		delete(m.attrs, attr)
	}
	return m
}

func isEnum(at schema.AvroType) bool {
	ref, ok := at.(*schema.Reference)
	if !ok {
		return false
	}
	_, ok = ref.Def.(*schema.EnumDefinition)
	return ok
}

func isRecord(at schema.AvroType) bool {
	ref, ok := at.(*schema.Reference)
	if !ok {
		return false
	}
	_, ok = ref.Def.(*schema.RecordDefinition)
	return ok
}

// stdAttrs maps from a given schema type to the fields
// that are defined by the standard for that type that
// we not want to appear as annotations.
var stdAttrs = map[reflect.Type][]string{
	elType(new(*schema.RecordDefinition)): {"name", "namespace", "fields"},
	elType(new(*schema.FixedDefinition)):  {"name", "namespace", "size"},
	elType(new(*schema.MapField)):         {"values"},
	elType(new(*schema.EnumDefinition)):   {"name", "namespace", "symbols"},
	elType(new(*schema.ArrayField)):       {"items"},
	elType(new(*schema.Field)):            {"name", "default"},
}

// elType returns the type of *v.
func elType(v interface{}) reflect.Type {
	return reflect.TypeOf(v).Elem()
}

func (g *generator) addDefinition(def schema.Definition) {
	if g.done[def.AvroName()] {
		return
	}
	g.queue = append(g.queue, def)
	g.done[def.AvroName()] = true
}

func (g *generator) removeDefinition() schema.Definition {
	if len(g.queue) == 0 {
		return nil
	}
	def := g.queue[0]
	g.queue = g.queue[1:]
	return def
}

func (g *generator) pushNamespace(ns string) {
	g.namespaceStack = append(g.namespaceStack, ns)
}

func (g *generator) popNamespace() {
	g.namespaceStack = g.namespaceStack[0 : len(g.namespaceStack)-1]
}

func (g *generator) namespace() string {
	return g.namespaceStack[len(g.namespaceStack)-1]
}

func jsonMarshal(v interface{}, indent string) []byte {
	data, err := json.MarshalIndent(v, indent, "\t")
	if err != nil {
		panic(fmt.Errorf("cannot json marshal default value: %v", err))
	}
	return data
}
