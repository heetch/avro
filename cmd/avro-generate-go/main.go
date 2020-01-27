// The avro-generate-go command generates Go types for the Avro schemas specified on the
// command line. Each schema file results in a Go file with the same basename but with a ".go" suffix.
//
// Usage:
//
//	usage: avrogen [flags] schema-file...
//	  -d string
//	    	directory to write Go files to (default ".")
//	  -p string
//	    	package name
//	  -t	generated files will have _test.go suffix
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var (
	dirFlag  = flag.String("d", ".", "directory to write Go files to")
	pkgFlag  = flag.String("p", "", "package name")
	testFlag = flag.Bool("t", false, "generated files will have _test.go suffix")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: avrogen [flags] schema-file...\n")
		flag.PrintDefaults()
		os.Exit(2)
	}
	flag.Parse()
	files := flag.Args()
	if len(files) == 0 {
		flag.Usage()
	}
	if *pkgFlag == "" {
		fmt.Fprintf(os.Stderr, "avrogen: -p flag must specify a package name\n")
		os.Exit(1)
	}
	if err := generateFiles(files); err != nil {
		fmt.Fprintf(os.Stderr, "avrogen: %v\n", err)
		os.Exit(1)
	}
}

func generateFiles(files []string) error {
	for _, f := range files {
		if err := generateFile(f); err != nil {
			return err
		}
	}
	return nil
}

func generateFile(f string) error {
	data, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := generate(&buf, data, *pkgFlag); err != nil {
		return err
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
