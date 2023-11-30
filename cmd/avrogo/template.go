package main

import (
	"go/token"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/actgardner/gogen-avro/v10/parser"
	"github.com/actgardner/gogen-avro/v10/schema"
)

func newTemplate(s string) *template.Template {
	return template.Must(template.New("").
		Funcs(templateFuncs).
		Delims("«", "»").
		Parse(s),
	)
}

var templateFuncs = template.FuncMap{
	"typeof":                 typeof,
	"isExportedGoIdentifier": isExportedGoIdentifier,
	"defName":                defName,
	"symbolName":             symbolName,
	"goName": func(gc *generateContext, name string) (string, error) {
		return gc.goName(name)
	},
	"indent": indent,
	"doc":    doc,
	"import": func(gc *generateContext, pkg string) string {
		gc.addImport(pkg)
		return ""
	},
}

type headerTemplateParams struct {
	Pkg       string
	Imports   []string
	ImportIds map[string]string
}

// TODO avoid explicit package identifiers
var headerTemplate = newTemplate(`
// Code generated by avrogen. DO NOT EDIT.

package «.Pkg»

import (
«range $imp := .Imports»«printf "%s %q" (index $.ImportIds $imp) $imp»
«end»)
`)

type bodyTemplateParams struct {
	Definitions []schema.QualifiedName
	NS          *parser.Namespace
	Ctx         *generateContext
}

var bodyTemplate = newTemplate(`
«range $defName := .Definitions»
	«$def := index $.NS.Definitions $defName»
	«with $def»
	«- if eq (typeof .) "RecordDefinition"»
		«- doc "// " .»
		type «defName .» struct {
		«- range $i, $_ := .Fields»
			«- doc "\t// " .»
			«- $type := $.Ctx.GoTypeOf .Type»
			«- doc "\t// " $type»
			«- if isExportedGoIdentifier .Name»
				«- .Name» «$type.GoType»
			«- else»
				«- goName $.Ctx .Name» «$type.GoType» ` + "`" + `json:«printf "%q" .Name»` + "`" + `
			«- end»
		«end»
		}

		// AvroRecord implements the avro.AvroRecord interface.
		func («defName .») AvroRecord() avrotypegen.RecordInfo {
			return «$.Ctx.RecordInfoLiteral .»
		}
	«else if eq (typeof .) "EnumDefinition"»
		«- import $.Ctx "strconv"»
		«- import $.Ctx "fmt"»
		«- doc "// " . -»
		type «defName .» int
		const (
		«- range $i, $sym := .Symbols»
		«symbolName $.Ctx $def $sym»«if eq $i 0» «defName $def» = iota«end»
		«- end»
		)

		var _«defName .»_strings = []string{
		«range $i, $sym := .Symbols»
		«- printf "%q" $sym»,
		«end»}

		// String returns the textual representation of «defName .».
		func (e «defName .») String() string {
			if e < 0 || int(e) >= len(_«defName .»_strings) {
				return "«defName .»(" + strconv.FormatInt(int64(e), 10) + ")"
			}
			return _«defName .»_strings[e]
		}

		// MarshalText implements encoding.TextMarshaler
		// by returning the textual representation of «defName .».
		func (e «defName .») MarshalText() ([]byte, error) {
			if e < 0 || int(e) >= len(_«defName .»_strings) {
				return nil, fmt.Errorf("«defName .» value %d is out of bounds", e)
			}
			return []byte(_«defName .»_strings[e]), nil
		}

		// UnmarshalText implements encoding.TextUnmarshaler
		// by expecting the textual representation of «.Name».
		func (e *«defName .») UnmarshalText(data []byte) error {
			// Note for future: this could be more efficient.
			for i, s := range _«defName .»_strings {
				if string(data) == s {
					*e = «defName .»(i)
					return nil
				}
			}
			return fmt.Errorf("unknown value %q for «defName .»", data)
		}
	«else if eq (typeof .) "FixedDefinition"»
		«- doc "// " . -»
		type «defName .» [«.SizeBytes»]byte
	«else»
		// unknown definition type «printf "%T; name %q" . (typeof .)» .
	«end»
	«end»
«end»
`[1:])

func defName(def schema.Definition) string {
	return goTypeForDefinition(def).Name
}

func symbolName(gc *generateContext, e *schema.EnumDefinition, symbol string) string {
	return defName(e) + gc.caser(symbol)
}

func quote(s string) string {
	if !strings.Contains(s, "`") {
		return "`" + s + "`"
	}
	return strconv.Quote(s)
}

type documented interface {
	Doc() string
}

func doc(indentStr string, d interface{}) string {
	if d, ok := d.(documented); ok && d.Doc() != "" {
		return "\n" + indent(trimAVDLDoc(d.Doc()), indentStr) + "\n"
	}
	return ""
}

// trimAVDLDoc removes indentation from a doc string
// that's been generated from an IDL file that was in
// the form:
//
//	/**
//	 * comment etc
//	 */
//
// The idl2schemata tool just includes the entire body
// of the comment verbatim, including the leading * and
// indentation characters. We don't want those
// in the Go comments.
func trimAVDLDoc(s0 string) string {
	s := strings.TrimPrefix(s0, "*")
	if len(s) == len(s0) {
		return s
	}
	s = avdlDocIndentPattern.ReplaceAllString(s, "\n")
	return s
}

var avdlDocIndentPattern = regexp.MustCompile(`\n\s*\* ?`)

func indent(s, with string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	return with + strings.Replace(s, "\n", "\n"+with, -1)
}

func isExportedGoIdentifier(s string) bool {
	return token.IsExported(s) && token.IsIdentifier(s)
}

func typeof(x interface{}) string {
	if x == nil {
		return "nil"
	}
	t := reflect.TypeOf(x)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if name := t.Name(); name != "" {
		return name
	}
	return ""
}
