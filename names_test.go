package avro_test

import (
	"encoding/json"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/heetch/avro"
	"github.com/heetch/avro/internal/testtypes"
)

func TestNamesRenameType(t *testing.T) {
	type S struct {
		A int
	}
	type T struct {
		S1 S
		S2 S
	}
	type W struct {
		E testtypes.Enum
		F [2]byte
		M map[string]*S
		R []TestRecord
		I int
	}
	tests := []struct {
		testName     string
		oldNames     []string
		newNames     []string
		newAliases   [][]string
		val          interface{}
		expectSchema string
	}{{
		testName: "go-type",
		oldNames: []string{"S"},
		newNames: []string{"foo"},
		val:      S{},
		expectSchema: `{
			"type": "record",
			"name": "foo",
			"fields": [{
				"name": "A",
				"type": "long",
				"default": 0
			}]
		}`,
	}, {
		testName: "generated-type",
		oldNames: []string{"TestRecord"},
		newNames: []string{"foo"},
		val:      TestRecord{},
		expectSchema: `{
			"name": "foo",
			"type": "record",
			"fields": [{
				"default": 42,
				"name": "A",
				"type": {
					"type": "int"
				}
			}, {
				"name": "B",
				"type": {
					"type": "int"
				}
			}]
		}`,
	}, {
		testName: "multiple-occurrences-of-renamed-type",
		oldNames: []string{"S"},
		newNames: []string{"bar"},
		val:      T{},
		expectSchema: `{
			"name": "T",
			"type": "record",
			"fields": [{
				"name": "S1",
				"default": {"A": 0},
				"type": {
					"type": "record",
					"name": "bar",
					"fields": [{
						"name": "A",
						"default": 0,
						"type": "long"
					}]
				}
			}, {
				"name": "S2",
				"default": {"A": 0},
				"type": "bar"
			}]
		}`,
	}, {
		testName: "multiple-renames",
		oldNames: []string{"S", "T"},
		newNames: []string{"newS", "newT"},
		val:      T{},
		expectSchema: `{
			"name": "newT",
			"type": "record",
			"fields": [{
				"name": "S1",
				"default": {"A": 0},
				"type": {
					"type": "record",
					"name": "newS",
					"fields": [{
						"name": "A",
						"default": 0,
						"type": "long"
					}]
				}
			}, {
				"name": "S2",
				"default": {"A": 0},
				"type": "newS"
			}]
		}`,
	}, {
		testName: "swap-names",
		oldNames: []string{"S", "T"},
		newNames: []string{"T", "S"},
		val:      T{},
		expectSchema: `{
			"name": "S",
			"type": "record",
			"fields": [{
				"name": "S1",
				"default": {"A": 0},
				"type": {
					"type": "record",
					"name": "T",
					"fields": [{
						"name": "A",
						"default": 0,
						"type": "long"
					}]
				}
			}, {
				"name": "S2",
				"default": {"A": 0},
				"type": "T"
			}]
		}`,
	}, {
		testName: "different-types",
		oldNames: []string{"W", "Enum", "go.Fixed2", "S", "TestRecord", "T", "S"},
		newNames: []string{"newW", "newEnum", "myFixed", "newS", "newTestRecord", "newT", "newS"},
		val:      W{},
		expectSchema: `{
			"name": "newW",
			"type": "record",
			"fields": [{
				"name": "E",
				"default": "One",
				"type": {
					"type": "enum",
					"name": "newEnum",
					"symbols": ["One", "Two", "Three"]
				}
			}, {
				"name": "F",
				"default": "\u0000\u0000",
				"type": {
					"type": "fixed",
					"name": "myFixed",
					"size": 2
				}
			}, {
				"name": "M",
				"default": {},
				"type": {
					"type": "map",
					"values": [
						"null",
						{
							"type": "record",
							"name": "newS",
							"fields": [{
								"name": "A",
								"type": "long",
								"default": 0
							}]
						}
					]
				}
			}, {
				"name": "R",
				"default": [],
				"type": {
					"type": "array",
					"items": {
						"name": "newTestRecord",
						"type": "record",
						"fields": [{
							"default": 42,
							"name": "A",
							"type": {
								"type": "int"
							}
						}, {
							"name": "B",
							"type": {
								"type": "int"
							}
						}]
					}
				}
			}, {
				"name": "I",
				"default": 0,
				"type": "long"
			}]
		}`,
	}, {
		testName: "relative-namespace",
		oldNames: []string{"S", "T"},
		newNames: []string{"a.newS", "a.newT"},
		val:      T{},
		expectSchema: `{
			"name": "a.newT",
			"type": "record",
			"fields": [{
				"name": "S1",
				"default": {"A": 0},
				"type": {
					"type": "record",
					"name": "newS",
					"fields": [{
						"name": "A",
						"default": 0,
						"type": "long"
					}]
				}
			}, {
				"name": "S2",
				"default": {"A": 0},
				"type": "newS"
			}]
		}`,
	}, {
		testName: "non-relative-namespace",
		oldNames: []string{"S", "T"},
		newNames: []string{"a.newS", "b.newT"},
		val:      T{},
		expectSchema: `{
			"name": "b.newT",
			"type": "record",
			"fields": [{
				"name": "S1",
				"default": {"A": 0},
				"type": {
					"type": "record",
					"name": "a.newS",
					"fields": [{
						"name": "A",
						"default": 0,
						"type": "long"
					}]
				}
			}, {
				"name": "S2",
				"default": {"A": 0},
				"type": "a.newS"
			}]
		}`,
	}, {
		testName:   "relative-alias-namespace",
		oldNames:   []string{"S"},
		newNames:   []string{"a.newS"},
		newAliases: [][]string{{"a.alias1", "b.alias2", "c.d.alias3"}},
		val:        S{},
		expectSchema: `{
			"type": "record",
			"name": "a.newS",
			"aliases": ["alias1", "b.alias2", "c.d.alias3"],
			"fields": [{
				"name": "A",
				"type": "long",
				"default": 0
			}]
		}`,
	}}

	c := qt.New(t)
	for _, test := range tests {
		c.Run(test.testName, func(c *qt.C) {
			names := new(avro.Names)
			for i, oldName := range test.oldNames {
				var aliases []string
				if i < len(test.newAliases) {
					aliases = test.newAliases[i]
				}
				names = names.Rename(oldName, test.newNames[i], aliases...)
			}
			at, err := names.TypeOf(test.val)
			c.Assert(err, qt.Equals, nil)
			c.Assert(at.String(), qt.JSONEquals, json.RawMessage(test.expectSchema))
		})
	}
}

func TestRenameBuiltinAvroTypePanics(t *testing.T) {
	c := qt.New(t)
	c.Assert(func() {
		new(avro.Names).Rename("int", "evil")
	}, qt.PanicMatches, `rename of built-in type "int" to "evil"`)
}

func TestRenameTypeSuccess(t *testing.T) {
	type S struct {
		A int
	}
	c := qt.New(t)
	names := new(avro.Names).RenameType(S{}, "newS")
	at, err := names.TypeOf(S{})
	c.Assert(err, qt.Equals, nil)
	c.Assert(at.String(), qt.JSONEquals, json.RawMessage(`{
		"type": "record",
		"name": "newS",
		"fields": [{
			"name": "A",
			"type": "long",
			"default": 0
		}]
	}`))
}

func TestRenameTypeBadType(t *testing.T) {
	c := qt.New(t)
	c.Assert(func() {
		new(avro.Names).RenameType(struct{}{}, "empty")
	}, qt.PanicMatches, `cannot rename struct {} to "empty": cannot get Avro type: cannot use unnamed type struct {} as Avro type`)
}

func TestRenameTypeNonDefinition(t *testing.T) {
	c := qt.New(t)
	c.Assert(func() {
		new(avro.Names).RenameType("", "myString")
	}, qt.PanicMatches, `cannot rename string to "myString": it does not represent an Avro definition`)
}
