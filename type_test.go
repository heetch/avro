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
	}, {
		"name": "b",
		"type": {
			"type": "long",
			"logicalType": "timestamp-millis"
		}
	}]}`,
	out: `{"name":"R","type":"record","fields":[{"name":"a","type":"string"},{"name":"b","type":"long"}]}`,
}, {
	testName: "spec-STRIP-retain-defaults",
	opts:     avro.RetainDefaults,
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
	}, {
		"name": "b",
		"type": {
			"type": "long",
			"logicalType": "timestamp-micros"
		}
	}]}`,
	out: `{"name":"R","type":"record","fields":[{"name":"a","type":"string","default":"hello"},{"name":"b","type":"long"}]}`,
}, {
	testName: "spec-STRIP-retain-logical-types",
	opts:     avro.RetainLogicalTypes,
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
	}, {
		"name": "b",
		"type": {
			"type": "long",
			"logicalType": "timestamp-micros"
		}
	}, {
		"name": "c",
		"type": {
			"type": "bytes",
			"logicalType": "decimal",
			"scale": 3,
			"precision": 6
		}
	}, {
		"name": "d",
		"type": {
			"type": "bytes",
			"logicalType": "decimal",
			"scale": 0,
			"precision": 6
		}
	}]}`,
	out: `{"name":"R","type":"record","fields":[{"name":"a","type":"string"},{"name":"b","type":{"type":"long","logicalType":"timestamp-micros"}},{"name":"c","type":{"type":"bytes","logicalType":"decimal","precision":6,"scale":3}},{"name":"d","type":{"type":"bytes","logicalType":"decimal","precision":6}}]}`,
}, {
	testName: "spec-STRIP-retain-all",
	opts:     avro.RetainAll,
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
	}, {
		"name": "b",
		"type": {
			"type": "long",
			"logicalType": "timestamp-micros"
		}
	}]}`,
	out: `{"name":"R","type":"record","fields":[{"name":"a","type":"string","default":"hello"},{"name":"b","type":{"type":"long","logicalType":"timestamp-micros"}}]}`,
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
	out: `{"name":"R","type":"record","fields":[{"name":"a","type":"string"},{"name":"b","type":{"name":"E","type":"enum","symbols":["x","y"]}},{"name":"c","type":{"type":"array","items":"int"}},{"name":"d","type":{"type":"map","values":"int"}},{"name":"e","type":{"name":"F","type":"fixed","size":20}}]}`,
}, {
	testName: "spec-STRINGS",
	opts:     avro.RetainDefaults,
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
	testName: "primitive-types",
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
}, {
	testName: "union-with-default",
	opts:     avro.RetainDefaults,
	in: `{"type": "record",
              "name": "R",
              "fields":[
              {"name": "U",
               "type": ["int","string"],
               "default": 3
              }]
        }`,
	out: `{"name":"R","type":"record","fields":[{"name":"U","type":["int","string"],"default":3}]}`,
}, {
	testName: "union-with-default-null",
	opts:     avro.RetainDefaults,
	in: `{"type": "record",
              "name": "R",
              "fields":[
              {"name": "U",
               "type": ["null","string"],
               "default": null
              }]
        }`,
	out: `{"name":"R","type":"record","fields":[{"name":"U","type":["null","string"],"default":null}]}`,
}, {
	testName: "empty-record",
	in: `{
	"name": "R",
	"type": "record",
	"fields": []
}`,
	out: `{"name":"R","type":"record","fields":[]}`,
}, {
	testName: "out-of-bounds-opts",
	in:       `"string"`,
	out:      `"string"`,
	opts:     15,
}}

func TestCanonicalString(t *testing.T) {
	c := qt.New(t)
	for _, test := range canonicalStringTests {
		c.Run(test.testName, func(c *qt.C) {
			t, err := avro.ParseType(test.in)
			c.Assert(err, qt.Equals, nil)
			canon := t.CanonicalString(test.opts)
			c.Assert(canon, qt.Equals, test.out)
			// Make sure that the sync.Once machinery is working OK.
			c.Assert(t.CanonicalString(test.opts), qt.Equals, test.out)
			// The canonical type of a canonical type should be the same.
			t1, err := avro.ParseType(canon)
			c.Assert(err, qt.Equals, nil)
			c.Assert(t1.String(), qt.Equals, canon)
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
