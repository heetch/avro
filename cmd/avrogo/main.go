// The avrogo command generates Go types for the Avro schemas specified on the
// command line. Each schema file results in a Go file with the same basename but with a ".go" suffix.
//
// Type names within different schemas may refer to one another;
// for example to put a shared definition in a separate .avsc file.
//
// Usage:
//
//	usage: avrogo [flags] schema-file...
//	  -d string
//	    	directory to write Go files to (default ".")
//	  -p string
//	    	package name (defaults to $GOPACKAGE)
//	  -t	generated files will have _test.go suffix
//	  -map string
//	    	map from Avro namespace to Go package.
//
// By default, a type is generated for each Avro definition
// in the schema. Some additional metadata fields are
// recognized:
//
// - If a definition has a "go.package" metadata
// field, the type from that package will be used instead.
// - If a definition has a "go.name" metadata field,
// the associated string will be used for the Go type name.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rogpeppe/gogen-avro/v7/parser"
	"github.com/rogpeppe/gogen-avro/v7/resolver"
	"github.com/rogpeppe/gogen-avro/v7/schema"
)

// Generate the tests.

//go:generate go run ./generatetestcode.go

var (
	dirFlag  = flag.String("d", ".", "directory to write Go files to")
	pkgFlag  = flag.String("p", os.Getenv("GOPACKAGE"), "package name (defaults to $GOPACKAGE)")
	testFlag = flag.Bool("t", strings.HasSuffix(os.Getenv("GOFILE"), "_test.go"), "generated files will have _test.go suffix (defaults to true if $GOFILE is a test file)")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: avrogo [flags] schema-file...\n")
		flag.PrintDefaults()
		os.Exit(2)
	}
	flag.Parse()
	files := flag.Args()
	if len(files) == 0 {
		flag.Usage()
	}
	if *pkgFlag == "" {
		fmt.Fprintf(os.Stderr, "avrogo: -p flag must specify a package name or set $GOPACKAGE\n")
		os.Exit(1)
	}
	if err := generateFiles(files); err != nil {
		fmt.Fprintf(os.Stderr, "avrogo: %v\n", err)
		os.Exit(1)
	}
}

func generateFiles(files []string) error {
	ns, fileDefinitions, err := parseFiles(files)
	if err != nil {
		return err
	}
	for i, f := range files {
		if err := generateFile(f, ns, fileDefinitions[i]); err != nil {
			return fmt.Errorf("cannot generate code for %s: %v", f, err)
		}
	}
	return nil
}

func generateFile(f string, ns *parser.Namespace, definitions []schema.QualifiedName) error {
	var buf bytes.Buffer
	if err := generate(&buf, *pkgFlag, ns, definitions); err != nil {
		return err
	}
	if buf.Len() == 0 {
		// No code produced (probably because all the definitions in this
		// avsc file are external).
		return nil
	}
	resultData, err := format.Source(buf.Bytes())
	if err != nil {
		fmt.Printf("%s\n", buf.Bytes())
		return fmt.Errorf("cannot format source: %v", err)
	}

	outFile := filepath.Base(f)
	outFile = strings.TrimSuffix(outFile, filepath.Ext(f)) + "_gen"
	if *testFlag {
		outFile += "_test"
	}
	outFile += ".go"
	outFile = filepath.Join(*dirFlag, outFile)
	if err := ioutil.WriteFile(outFile, resultData, 0666); err != nil {
		return err
	}
	return nil
}

// parseFiles parses the Avro schemas in the given files and returns
// a namespace containing all of the definitions in all of the files
// and a slice with an element for each file holding a slice
// of all the definitions within that file.
func parseFiles(files []string) (*parser.Namespace, [][]schema.QualifiedName, error) {
	var fileDefinitions [][]schema.QualifiedName
	ns := parser.NewNamespace(false)
	for _, f := range files {
		data, err := ioutil.ReadFile(f)
		if err != nil {
			return nil, nil, err
		}
		var definitions []schema.QualifiedName
		// Make a new namespace just for this file only
		// so we can tell which names are defined in this
		// file alone.
		singleNS := parser.NewNamespace(false)
		avroType, err := singleNS.TypeForSchema(data)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid schema in %s: %v", f, err)
		}
		if _, ok := avroType.(*schema.Reference); !ok {
			// The schema doesn't have a top-level name.
			// TODO how should we cope with a schema that's not
			// a definition? In that case we don't have
			// a name for the type, and we may not be able to define
			// methods on it because it might be a union type which
			// is represented by an interface type in Go.
			// See https://github.com/heetch/avro/issues/13
			return nil, nil, fmt.Errorf("cannot generate code for schema %q which hasn't got a name (%T)", f, avroType)
		}
		for name, def := range singleNS.Definitions {
			if name != def.AvroName() {
				// It's an alias, so ignore it.
				continue
			}
			definitions = append(definitions, name)
		}
		// Sort the definitions so we get deterministic output.
		// TODO sort topologically so we get top level definitions
		// before lower level definitions.
		sort.Slice(definitions, func(i, j int) bool {
			return definitions[i].String() < definitions[j].String()
		})
		fileDefinitions = append(fileDefinitions, definitions)
		// Parse the schema again but use the global namespace
		// this time so all the schemas can share the same definitions.
		if _, err := ns.TypeForSchema(data); err != nil {
			return nil, nil, fmt.Errorf("cannot parse schema in %s: %v", f, err)
		}
	}
	// Now we've accumulated all the available types,
	// resolve the names with respect to the complete
	// namespace.
	for _, def := range ns.Roots {
		if err := resolver.ResolveDefinition(def, ns.Definitions); err != nil {
			// TODO find out which file(s) the definition came from
			// and include that file name in the error.
			return nil, nil, fmt.Errorf("cannot resolve reference %q: %v", def, err)
		}
	}
	return ns, fileDefinitions, nil
}
