package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"gopkg.in/errgo.v2/fmt/errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"

	"github.com/actgardner/gogen-avro/v10/parser"
	"github.com/actgardner/gogen-avro/v10/schema"

	"github.com/heetch/avro/cmd/avrogo/avrotypemap"
	"github.com/heetch/avro/internal/typeinfo"
)

type goType = avrotypemap.GoType

// typeMap returns a map from definitions
// in ns to the external Go types used.
func externalTypeMap(ns *parser.Namespace) (map[schema.QualifiedName]goType, error) {
	extGoTypes := make(map[goType]bool)
	for _, def := range ns.Definitions {
		if gt := goTypeForDefinition(def); gt.PkgPath != "" {
			extGoTypes[gt] = true
		}
	}
	if len(extGoTypes) == 0 {
		// No external types found.
		return nil, nil
	}
	extTypeInfo, err := externalTypeInfoForGoTypes(extGoTypes)
	if err != nil {
		return nil, err
	}

	// For each definition, we want to know the external go type to use (if any).
	// When we find an external go type, go through all the avro names in it,
	// adding them to extGoTypes. If we find an existing name that has
	// a different go type, it's an error.

	avroToGo := make(map[schema.QualifiedName]goType)
	for name, def := range ns.Definitions {
		gt := goTypeForDefinition(def)
		if gt.PkgPath == "" {
			continue
		}
		// The definition refers to a type in an external package.
		// That means that:
		// a) the type must be compatible with the schema.
		// Specifically, it must have all the fields the enum symbols
		// that the schema does.
		// b) any types within it that refer to definitions must also
		// be replaced.
		info, ok := extTypeInfo[gt]
		if !ok {
			// We should have acquired the necessary info
			// with externalTypeInfoForGoTypes above.
			panic(errors.Newf("type info for %s not found", gt))
		}
		extType, err := typeinfo.ParseSchema(info.Schema, nil)
		if err != nil {
			return nil, errors.Newf("cannot parse schema from external type: %v", err)
		}
		if err := checkGoCompatible(name.String(), &schema.Reference{
			TypeName: name,
			Def:      def,
		}, extType, make(map[schema.AvroType]bool)); err != nil {
			return nil, errors.Newf("external type is not compatible; external schema %s; name %s: %v", info.Schema, name, err)
		}
		for qname, extType := range info.Map {
			name := parser.ParseAvroName("", qname)
			if extType0, ok := avroToGo[name]; !ok {
				avroToGo[name] = extType
			} else if extType != extType0 {
				// We've found another mention of the same Avro type that's
				// associated with a different Go type.
				return nil, errors.Newf("different external names for Avro name %s (%s vs %s)", name, extType, extType0)
			}
		}
	}
	return avroToGo, nil
}

// externalTypeInfoForGoTypes find information on all the given
// Go types by creating and running a temporary Go program
// that obtains the Avro type information by calling avro.TypeOf.
//
// We generate the code in the output directory so that it
// will use the same module used by that and have access
// to exactly the same set of types that other Go files
// in that directory will do (for example internal types).
//
// TODO This isn't nice, but it's not clear how we can avoid it because
// the enum logic relies on calling the String method, which
// we can't do unless we actually run it.
func externalTypeInfoForGoTypes(gts map[goType]bool) (map[goType]avrotypemap.ExternalTypeResult, error) {

	pkgs := make(map[string]int)
	var pkgPaths []string
	addPkg := func(pkgPath string) {
		if _, ok := pkgs[pkgPath]; !ok {
			pkgs[pkgPath] = len(pkgPaths)
			pkgPaths = append(pkgPaths, pkgPath)
		}
	}
	addPkg("github.com/heetch/avro")
	addPkg("github.com/heetch/avro/cmd/avrogo/avrotypemap")
	for gt := range gts {
		addPkg(gt.PkgPath)
	}
	mp := typeInfoMainParams{
		ImportIDs: make(map[string]string),
	}
	for _, p := range pkgPaths {
		mp.Imports = append(mp.Imports, p)
		// TODO make sure there's no duplicate name.
		mp.ImportIDs[p] = importPathToName(p)
	}
	for gt := range gts {
		mp.Types = append(mp.Types, gt)
	}
	var buf bytes.Buffer
	if err := typeInfoMainTemplate.Execute(&buf, mp); err != nil {
		return nil, errors.Newf("cannot execute type info main template: %v", err)
	}
	resultData, err := format.Source(buf.Bytes())
	if err != nil {
		fmt.Printf("%s\n", buf.Bytes())
		return nil, errors.Newf("cannot format typeinfo source: %v", err)
	}
	f, err := ioutil.TempFile(*dirFlag, "avro-introspect*.go")
	if err != nil {
		return nil, err
	}
	prog := f.Name()
	defer os.Remove(prog)
	defer f.Close()
	if _, err := f.Write(resultData); err != nil {
		return nil, errors.Newf("cannot write %q: %v", f.Name(), err)
	}
	f.Close()
	var runStdout bytes.Buffer
	cmd := exec.Command("go", "run", filepath.Base(prog))
	cmd.Dir = *dirFlag
	cmd.Stdout = &runStdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, errors.Newf("failed to run introspect program: %v", err)
	}
	var resultSlice []avrotypemap.ExternalTypeResult
	if err := json.Unmarshal(runStdout.Bytes(), &resultSlice); err != nil {
		return nil, errors.Newf("cannot unmarshal introspect output %q: %v", runStdout.Bytes(), err)
	}
	if len(resultSlice) != len(gts) {
		return nil, errors.Newf("unexpected result count, got %d want %d", len(resultSlice), len(gts))
	}
	results := make(map[goType]avrotypemap.ExternalTypeResult)
	for i, result := range resultSlice {
		results[mp.Types[i]] = result
	}
	return results, nil
}

type typeInfoMainParams struct {
	Imports   []string
	ImportIDs map[string]string
	Types     []goType
}

var typeInfoMainTemplate = newTemplate(`
//+build ignore

// Code generated by avrogo. DO NOT EDIT.

// This program prints information about all
// types in the types slice in JSON format.
//
// REMOVE ME:
// This program is a temporary artifact of avrogo and
// if you're seeing it, something has gone wrong.
package main

import (
	"encoding/json"
	"fmt"
	"reflect"

«range $imp := .Imports»«printf "%s %q" (index $.ImportIDs $imp) $imp»
«end»)

var types = []interface{}{«range .Types»
	new(«index $.ImportIDs .PkgPath».«.Name»),«end»
}

func main() {
	var results []avrotypemap.ExternalTypeResult
	for _, t := range types {
		avroType, typeMap, err := infoForType(reflect.TypeOf(t).Elem())
		if err != nil {
			results = append(results, avrotypemap.ExternalTypeResult{
				Error: err.Error(),
			})
		} else {
			results = append(results, avrotypemap.ExternalTypeResult{
				Schema: avroType.String(),
				Map: typeMap,
			})
		}
	}
	data, err := json.Marshal(results)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", data)
}

func infoForType(t reflect.Type) (*avro.Type, map[string]avrotypemap.GoType, error) {
	at, err := avro.TypeOf(reflect.Zero(t).Interface())
	if err != nil {
		return nil, nil, errors.Newf("cannot get Avro type for %s: %v", t, err)
	}
	m, err := avrotypemap.AvroTypeMap(t)
	if err != nil {
		return nil, nil, errors.Newf("cannot get type map for %s: %v", t, err)
	}
	return at, m, nil
}
`)

type compatibilityError struct {
	path string
	msg  string
}

func (e *compatibilityError) Error() string {
	return fmt.Sprintf("incompatible at %s: %v", e.path, e.msg)
}

func compatErrorf(path string, f string, args ...interface{}) error {
	return &compatibilityError{
		path: path,
		msg:  fmt.Sprintf(f, args...),
	}
}

// checkGoCompatible checks that the external schema type t2
// is compatible with t1.
func checkGoCompatible(path string, t1, t2 schema.AvroType, checked map[schema.AvroType]bool) error {
	// Protect against infinite recursion.
	if checked[t1] {
		return nil
	}
	checked[t1] = true
	if reflect.TypeOf(t1) != reflect.TypeOf(t2) {
		return compatErrorf(path, "incompatible types")
	}
	switch t1 := t1.(type) {
	case *schema.Reference:
		t2 := t2.(*schema.Reference)
		if reflect.TypeOf(t1.Def) != reflect.TypeOf(t2.Def) {
			return compatErrorf(path, "incompatible types")
		}
		if t1.TypeName != t2.TypeName {
			return compatErrorf(path, "incompatible type name %v vs %v", t1.TypeName, t2.TypeName)
		}
		switch t1def := t1.Def.(type) {
		case *schema.RecordDefinition:
			// All fields in t1 must appear in t2.
			t2def := t2.Def.(*schema.RecordDefinition)
			for _, f1 := range t1def.Fields() {
				path := path + "." + f1.Name()
				f2 := t2def.FieldByName(f1.Name())
				if f2 == nil {
					return compatErrorf(path, "field not found in external definition")
				}
				if err := checkGoCompatible(path, f1.Type(), f2.Type(), checked); err != nil {
					return err
				}
			}
			return nil
		case *schema.EnumDefinition:
			// All symbols in t1 must be in t2.
			t2def := t2.Def.(*schema.EnumDefinition)
			syms1, syms2 := t1def.Symbols(), t2def.Symbols()
			if len(syms2) < len(syms1) {
				return compatErrorf(path, "external enum does not contain all symbols")
			}
			for i, sym1 := range syms1 {
				if sym1 != syms2[i] {
					return compatErrorf(path, "enum symbol changed value")
				}
			}
			return nil
		case *schema.FixedDefinition:
			// The size must be the same.
			t2def := t2.Def.(*schema.FixedDefinition)
			if t1def.SizeBytes() != t2def.SizeBytes() {
				return compatErrorf(path, "fixed size changed value")
			}
		default:
			panic(errors.Newf("unknown definition type %T", t1def))
		}
	case *schema.UnionField:
		// All members of the union must be compatible.
		t2 := t2.(*schema.UnionField)
		itemTypes1, itemTypes2 := t1.ItemTypes(), t2.ItemTypes()

		if len(itemTypes1) != len(itemTypes2) {
			// TODO we could potentially let the external type add new
			// union members.
			return errors.Newf("union type mismatch")
		}
		for i := range itemTypes1 {
			if err := checkGoCompatible(path+fmt.Sprintf("[u%d]", i), itemTypes1[i], itemTypes2[i], checked); err != nil {
				return err
			}
		}
	case *schema.MapField:
		return checkGoCompatible(path+".{*}", t1.ItemType(), t2.(*schema.MapField).ItemType(), checked)
	case *schema.ArrayField:
		return checkGoCompatible(path+".[*]", t1.ItemType(), t2.(*schema.ArrayField).ItemType(), checked)
	case *schema.LongField:
		if logicalType(t1) != logicalType(t2) {
			// TODO check for timestamp-micros only?
			return errors.Newf("logical type mismatch")
		}
	}
	// We already know the type matches and there's nothing
	// else to check.
	return nil
}
