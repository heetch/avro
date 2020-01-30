package avrotypemap_test

import (
	"reflect"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/heetch/avro/cmd/avro-generate-go/avrotypemap"
)

//go:generate avro-generate-go union.avsc

func TestTypeInfo(t *testing.T) {
	c := qt.New(t)

	type A struct {
		F int
	}

	type B struct {
		Array []A
		Map   map[string]A
	}

	type C struct {
		F B
		G A
	}

	m, err := avrotypemap.AvroTypeMap(reflect.TypeOf(C{}))
	c.Assert(err, qt.Equals, nil)
	c.Assert(m, qt.DeepEquals, map[string]avrotypemap.GoType{
		"A": {testPkg, "A"},
		"B": {testPkg, "B"},
		"C": {testPkg, "C"},
	})
}

const testPkg = "github.com/heetch/avro/cmd/avro-generate-go/avrotypemap_test"

func TestUnion(t *testing.T) {
	c := qt.New(t)
	m, err := avrotypemap.AvroTypeMap(reflect.TypeOf(U{}))
	c.Assert(err, qt.Equals, nil)
	c.Assert(m, qt.DeepEquals, map[string]avrotypemap.GoType{
		"U":   {testPkg, "U"},
		"UR1": {testPkg, "UR1"},
		"UR2": {testPkg, "UR2"},
	})
}

func TestRecursiveType(t *testing.T) {
	c := qt.New(t)
	m, err := avrotypemap.AvroTypeMap(reflect.TypeOf(Recur{}))
	c.Assert(err, qt.Equals, nil)
	c.Assert(m, qt.DeepEquals, map[string]avrotypemap.GoType{
		"Recur": {testPkg, "Recur"},
	})
}

type Recur struct {
	A int
	R *Recur
}
