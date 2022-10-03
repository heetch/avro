// The avrogo command generates Go types for the Avro schemas specified on the
// command line. Each schema file results in a Go file with the same basename but with a ".go" suffix,
// or multiple files if -split is set.
// If multiple schema files have the same basename, successively more elements of their
// full path are used (replacing path separators with "_") until they're not the same any more.
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
//	  -s string
//	    	suffix for generated files (default "_gen")
//    -tokenize
//          if true, generate one dedicated file per type found in the schema files
//
// By default, a type is generated for each Avro definition
// in the schema. Some additional metadata fields are
// recognized:
//
// See the README for a full description of how schemas
// map to generated Go types: https://github.com/heetch/avro/blob/master/README.md
package main

import (
	"bytes"
	stdflag "flag"
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/actgardner/gogen-avro/v10/parser"
	"github.com/actgardner/gogen-avro/v10/resolver"
	"github.com/actgardner/gogen-avro/v10/schema"
)

// Generate the tests.

//go:generate go run ./generatetestcode.go

var (
	dirFlag      = flag.String("d", ".", "directory to write Go files to")
	pkgFlag      = flag.String("p", os.Getenv("GOPACKAGE"), "package name (defaults to $GOPACKAGE)")
	testFlag     = flag.Bool("t", strings.HasSuffix(os.Getenv("GOFILE"), "_test.go"), "generated files will have _test.go suffix (defaults to true if $GOFILE is a test file)")
	suffixFlag   = flag.String("s", "_gen", "suffix for generated files")
	tokenizeFlag = flag.Bool("tokenize", false, "generate one dedicated file per message")
)

var flag = stdflag.NewFlagSet("", stdflag.ContinueOnError)

func main() {
	os.Exit(main1())
}

func main1() int {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: avrogo [flags] schema-file...\n")
		flag.PrintDefaults()
	}
	if flag.Parse(os.Args[1:]) != nil {
		return 2
	}
	files := flag.Args()
	if len(files) == 0 {
		flag.Usage()
		return 2
	}
	if *pkgFlag == "" {
		fmt.Fprintf(os.Stderr, "avrogo: -p flag must specify a package name or set $GOPACKAGE\n")
		return 1
	}
	if err := generateFiles(files); err != nil {
		fmt.Fprintf(os.Stderr, "avrogo: %v\n", err)
		return 1
	}
	return 0
}

func generateFiles(files []string) error {
	ns, fileDefinitions, err := parseFiles(files)
	if err != nil {
		return err
	}

	outfiles, err := outputPaths(files, *testFlag)
	if err != nil {
		return err
	}

	if *tokenizeFlag {
		for _, fileDefinition := range fileDefinitions {
			for _, qualifiedName := range fileDefinition {
				outputPath := path.Join(strings.ToLower(qualifiedName.Name) + *suffixFlag + ".go")
				singleFileList := []schema.QualifiedName{qualifiedName}

				if err := generateFile(outputPath, ns, singleFileList); err != nil {
					return fmt.Errorf("cannot generate code for %s.%s: %v", qualifiedName.Namespace, qualifiedName.Name, err)
				}
			}
		}
	} else {
		for i, f := range files {
			if err := generateFile(outfiles[f], ns, fileDefinitions[i]); err != nil {
				return fmt.Errorf("cannot generate code for %s: %v", f, err)
			}
		}
	}

	return nil
}

func outputPaths(files []string, testFile bool) (map[string]string, error) {
	fileset := make(map[string]string)
	for _, file := range files {
		fileset[file] = outputPath(file, testFile)
	}
	need := len(fileset)
	result := make(map[string]string)
	for level := 1; len(result) < need; level++ {
		found := make(map[string]int)
		for _, new := range result {
			found[new]++
		}
		allOK := true
		for old, clean := range fileset {
			b, ok := baseN(clean, level)
			allOK = allOK && ok
			found[b]++
			// Tentatively set the result. It'll be removed below if found to
			// be ambiguous.
			result[old] = b
		}
		for old, new := range result {
			if _, ok := fileset[old]; ok && found[new] > 1 {
				// Ambiguous name found in this round. Remove from the results, and we'll
				// try again next time around the loop with another level
				// of path included.
				delete(result, old)
			} else {
				// Resolved unambiguously. We don't need to consider this in
				// future rounds.
				delete(fileset, old)
			}
		}
		if !allOK && len(fileset) > 0 {
			// We've got to the end of some paths and failed to resolve all the files
			// unambigously, so avoid the potential infinite loop by returning an error.
			return nil, fmt.Errorf("could not make unambiguous output files from input files")
		}
	}
	return result, nil
}

// outputPath returns the output Go filename to
// use for the given input avsc file. It retains the directory
// information but converts to a /-separated path for
// ease of processing.
func outputPath(filename string, testFile bool) string {
	filename = filepath.Clean(filename)
	filename = filename[len(filepath.VolumeName(filename)):]
	filename = filepath.ToSlash(filename)
	filename = strings.TrimSuffix(filename, filepath.Ext(filename)) + *suffixFlag
	if testFile {
		filename += "_test"
	}
	filename += ".go"
	return filename
}

// baseN returns the last n /-separated path elements of name
// joined by underscores.
// So baseN("foo/bar/baz", 2) would return "bar_baz".
// It reports whether there were actually n path elements to take.
func baseN(name string, n int) (string, bool) {
	parts := strings.Split(name, "/")
	if parts[0] == "" {
		// This can only happen if the path is absolute.
		// Go files aren't allowed to start
		// with _ so use an arbitrary string instead.
		parts[0] = "slash"
	}
	ok := len(parts) >= n
	if ok {
		parts = parts[len(parts)-n:]
	}
	return strings.Join(parts, "_"), ok
}

func generateFile(outFile string, ns *parser.Namespace, definitions []schema.QualifiedName) error {
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
	if err := os.MkdirAll(*dirFlag, 0777); err != nil {
		return fmt.Errorf("cannot create output directory: %v", err)
	}
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
	for name, def := range ns.Roots {
		if err := resolver.ResolveDefinition(def, ns.Definitions); err != nil {
			// TODO find out which file(s) the definition came from
			// and include that file name in the error.
			return nil, nil, fmt.Errorf("cannot resolve reference %q: %v", name, err)
		}
	}
	return ns, fileDefinitions, nil
}
