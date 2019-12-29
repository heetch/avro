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
	pkgFlag  = flag.String("p", "wiretypes", "package name")
	testFlag = flag.Bool("t", false, "generated files will have _test.go suffix")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "avrogen [flags] schema-file...\n")
		flag.PrintDefaults()
		os.Exit(2)
	}
	flag.Parse()
	files := flag.Args()
	if len(files) == 0 {
		flag.Usage()
	}
	if err := generateFiles(files); err != nil {
		fmt.Fprintf(os.Stderr, "avrogen: %v\n", err)
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
		fmt.Printf("-------\n%s\n", buf.Bytes())
		return fmt.Errorf("cannot format source: %v", err)
	}

	outFile := filepath.Base(f)
	outFile = strings.TrimSuffix(f, filepath.Ext(f)) + "_gen"
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
