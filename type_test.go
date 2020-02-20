package avro_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/heetch/avro"
)

// NOTE: the INTEGERS part of the canonicalizing specification is redundant
// because JSON doesn't allow leading zeros anyway.

var canonicalStringTests = []struct {
	testName string
	opts     avro.CanonicalOpts
	in       string
	out      string
}{{
	testName: "spec-WHITESPACE",
	in:       "    \"string\"   \n ",
	out:      `"string"`,
}, {
	testName: "spec-PRIMITIVES",
	in:       `{"type": "int"}`,
	out:      `"int"`,
}, {
	testName: "spec-STRIP",
	opts:     avro.LeaveDefaults,
	in: `{
	"type": "record",
	"name":"R",
	"doc": "documentation",
	"extra-meta":"hello",
	"aliases": ["a", "b"],
	"fields": [{
		"name": "a",
		"type": "string",
		"default": "hello"
	}]}`,
	out: `{"name":"R","type":"record","fields":[{"name":"a","type":"string","default":"hello"}]}`,
}, {
	testName: "spec-STRIP-include-defaults",
	in: `{
	"type": "record",
	"name":"R",
	"doc": "documentation",
	"extra-meta":"hello",
	"aliases": ["a", "b"],
	"fields": [{
		"name": "a",
		"type": "string",
		"default": "hello"
	}]}`,
	out: `{"name":"R","type":"record","fields":[{"name":"a","type":"string"}]}`,
}, {
	testName: "spec-ORDER",
	in: `{
		"fields":[{
			"name":"a",
			"type":"string"
		}, {
			"name": "b",
			"type": {
				"symbols": ["x", "y"],
				"type": "enum",
				"name": "E"
			}
		}, {
			"name": "c",
			"type": {
				"items": "int",
				"type": "array"
			}
		}, {
			"name": "d",
			"type": {
				"values": "int",
				"type": "map"
			}
		}, {
			"name": "e",
			"type": {
				"size": 20,
				"type": "fixed",
				"name": "F"
			}
		}],
		"type":"record",
		"name":"R"
	}`,
	out: `{"name":"R","type":"record","fields":[{"name":"a","type":"string"},{"name":"b","type":{"name":"E","type":"enum","symbols":["x","y"]}},{"name":"c","type":{"type":"array","items":"int"}},{"name":"d","type":{"type":"map","items":"int"}},{"name":"e","type":{"name":"F","type":"fixed","size":20}}]}`,
}, {
	testName: "spec-STRINGS",
	opts:     avro.LeaveDefaults,
	in: `{
		"name":"\u0052",
		"type":"record",
		"fields":[{
			"name":"a",
			"type":"string",
			"default":"hello<>&\u00e9\u003e"
		}]
	}`,
	out: `{"name":"R","type":"record","fields":[{"name":"a","type":"string","default":"hello<>&é>"}]}`,
}, {
	in: `{
	"name": "R",
	"namespace": "com.example",
	"type": "record",
	"fields": [{
		"name": "a",
		"type": {
			"type": "enum",
			"name": "E",
			"symbols": ["a", "b"]
		}
	}, {
		"name": "b",
		"type": {
			"type": "enum",
			"name": "foo.F",
			"symbols": ["a", "b"]
		}
	}, {
		"name": "c",
		"type": "E"
	}]
}`,
	out: `{"name":"com.example.R","type":"record","fields":[{"name":"a","type":{"name":"com.example.E","type":"enum","symbols":["a","b"]}},{"name":"b","type":{"name":"foo.F","type":"enum","symbols":["a","b"]}},{"name":"c","type":"com.example.E"}]}`,
}, {
	testName: "primitive_types",
	in: `{
	"name": "R",
	"type": "record",
	"fields": [{
		"name": "a",
		"type": "int"
	}, {
		"name": "b",
		"type": "long"
	}, {
		"name": "c",
		"type": "float"
	}, {
		"name": "d",
		"type": "double"
	}, {
		"name": "e",
		"type": "null"
	}, {
		"name": "f",
		"type": "long"
	}, {
		"name": "g",
		"type": "boolean"
	}, {
		"name": "h",
		"type": "bytes"
	}]
}`,
	out: `{"name":"R","type":"record","fields":[{"name":"a","type":"int"},{"name":"b","type":"long"},{"name":"c","type":"float"},{"name":"d","type":"double"},{"name":"e","type":"null"},{"name":"f","type":"long"},{"name":"g","type":"boolean"},{"name":"h","type":"bytes"}]}`,
}, {
	testName: "union",
	in: `["int","string", {
		"type": "enum",
		"name": "E",
		"symbols": ["a", "b"]
	}]`,
	out: `["int","string",{"name":"E","type":"enum","symbols":["a","b"]}]`,
}}

func TestCanonicalString(t *testing.T) {
	c := qt.New(t)
	for _, test := range canonicalStringTests {
		c.Run(test.testName, func(c *qt.C) {
			t, err := avro.ParseType(test.in)
			c.Assert(err, qt.Equals, nil)
			c.Assert(t.CanonicalString(test.opts), qt.Equals, test.out)
		})
	}
}

func mustParseType(s string) *avro.Type {
	t, err := avro.ParseType(s)
	if err != nil {
		panic(err)
	}
	return t
}
