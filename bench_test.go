package avro

import (
	"testing"

	"github.com/actgardner/gogen-avro/compiler"
	"github.com/actgardner/gogen-avro/schema"
	qt "github.com/frankban/quicktest"
)

func BenchmarkMakeAvroType(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ns := schema.NewNamespace(false)
		sType, err := ns.TypeForSchema([]byte(sample))
		if err != nil {
			b.Fatal(err)
		}
		err = sType.ResolveReferences(ns)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCompileSchema(b *testing.B) {
	c := qt.New(b)
	ns := schema.NewNamespace(false)
	sType, err := ns.TypeForSchema([]byte(sample))
	c.Assert(err, qt.Equals, nil)
	err = sType.ResolveReferences(ns)
	c.Assert(err, qt.Equals, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := compiler.Compile(sType, sType)
		if err != nil {
			b.Fatal(err)
		}
	}
}

const sample = `
{
                "name": "sample",
                "type": "record",
                "fields": [
                    {
                        "name": "header",
                        "type": [
                            "null",
                            {
                                "name": "Data0",
                                "type": "record",
                                "fields": [
                                    {
                                        "name": "uuid",
                                        "type": [
                                            "null",
                                            {
                                                "name": "UUID0",
                                                "type": "record",
                                                "fields": [
                                                    {
                                                        "name": "uuid",
                                                        "type": "string",
                                                        "default": ""
                                                    }
                                                ],
                                                "namespace": "headerworks.datatype",
                                                "doc": "A Universally Unique Identifier, in canonical form in lowercase. Example: de305d54-75b4-431b-adb2-eb6b9e546014"
                                            }
                                        ],
                                        "default": null,
                                        "doc": "Unique identifier for the event used for de-duplication and tracing."
                                    },
                                    {
                                        "name": "hostname",
                                        "type": [
                                            "null",
                                            "string"
                                        ],
                                        "default": null,
                                        "doc": "Fully qualified name of the host that generated the event that generated the data."
                                    },
                                    {
                                        "name": "trace",
                                        "type": [
                                            "null",
                                            {
                                                "name": "Trace0",
                                                "type": "record",
                                                "fields": [
                                                    {
                                                        "name": "traceId",
                                                        "type": [
                                                            "null",
                                                            "headerworks.datatype.UUID0"
                                                        ],
                                                        "default": null,
                                                        "doc": "Trace Identifier"
                                                    }
                                                ],
                                                "doc": "Trace0"
                                            }
                                        ],
                                        "default": null,
                                        "doc": "Trace information not redundant with this object"
                                    }
                                ],
                                "namespace": "headerworks",
                                "doc": "Common information related to the event which must be included in any clean event"
                            }
                        ],
                        "default": null,
                        "doc": "Core data information required for any event"
                    },
                    {
                        "name": "body",
                        "type": [
                            "null",
                            {
                                "name": "Data1",
                                "type": "record",
                                "fields": [
                                    {
                                        "name": "uuid",
                                        "type": [
                                            "null",
                                            {
                                                "name": "UUID1",
                                                "type": "record",
                                                "fields": [
                                                    {
                                                        "name": "uuid",
                                                        "type": "string",
                                                        "default": ""
                                                    }
                                                ],
                                                "namespace": "bodyworks.datatype",
                                                "doc": "A Universally Unique Identifier, in canonical form in lowercase. Example: de305d54-75b4-431b-adb2-eb6b9e546014"
                                            }
                                        ],
                                        "default": null,
                                        "doc": "Unique identifier for the event used for de-duplication and tracing."
                                    },
                                    {
                                        "name": "hostname",
                                        "type": [
                                            "null",
                                            "string"
                                        ],
                                        "default": null,
                                        "doc": "Fully qualified name of the host that generated the event that generated the data."
                                    },
                                    {
                                        "name": "trace",
                                        "type": [
                                            "null",
                                            {
                                                "name": "Trace1",
                                                "type": "record",
                                                "fields": [
                                                    {
                                                        "name": "traceId",
                                                        "type": [
                                                            "null",
                                                            "headerworks.datatype.UUID0"
                                                        ],
                                                        "default": null,
                                                        "doc": "Trace Identifier"
                                                    }
                                                ],
                                                "doc": "Trace1"
                                            }
                                        ],
                                        "default": null,
                                        "doc": "Trace information not redundant with this object"
                                    }
                                ],
                                "namespace": "bodyworks",
                                "doc": "Common information related to the event which must be included in any clean event"
                            }
                        ],
                        "default": null,
                        "doc": "Core data information required for any event"
                    }
                ],
                "namespace": "com.avro.test",
                "doc": "GoGen test"
            }
`
