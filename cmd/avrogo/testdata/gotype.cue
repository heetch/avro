package roundtrip

tests: goTypeMultipleFields: {
	inSchema: {
		name: "R"
		type: "record"
		fields: [{
			name: "A"
			type: "int"
		}, {
			name: "B"
			type: "int"
		}]
	}
	goType: "R"
	goTypeBody: """
		struct {
			B int
			A int
		}
	"""
	inData: {
		A: 1
		B: 2
	}
	outData: inData
}

tests: goTypePointer: {
	inSchema: {
		name: "R"
		type: "record"
		fields: [{
			name: "A"
			type: ["null", "long"]
		}]
	}
	goType: "R"
	goTypeBody: """
		struct {
			A *int
		}
	"""
}

tests: goTypePointer: subtests: non_null: {
	inData: A: long: 99
	outData: inData
}

tests: goTypePointer: subtests: null: {
	inData: A: null
	outData: inData
}

tests: goTypeSlice: {
	inSchema: {
		name: "R"
		type: "record"
		fields: [{
			name: "A"
			type: {
				type:  "array"
				items: "string"
			}
		}]
	}
	goType: "R"
	goTypeBody: """
		struct {
			A []string
		}
	"""
}

tests: goTypeSlice: subtests: non_empty: {
	inData: A: ["a", "b", "cd"]
	outData: inData
}

tests: goTypeSlice: subtests: empty: {
	inData: A: []
	outData: inData
}

tests: goTypeMap: {
	inSchema: {
		name: "R"
		type: "record"
		fields: [{
			name: "A"
			type: {
				type:   "map"
				values: "string"
			}
		}]
	}
	goType: "R"
	goTypeBody: """
		struct {
			A map[string]string
		}
	"""
}

tests: goTypeMap: subtests: non_empty: {
	inData: A: {
		a: "b"
		c: "d"
	}
	outData: inData
}

tests: goTypeMap: subtests: empty: {
	inData: A: {}
	outData: inData
}

tests: goTypeFixed: {
	inSchema: {
		name: "R"
		type: "record"
		fields: [{
			name: "A"
			type: {
				type: "fixed"
				size: 3
				name: "go.Fixed3"
			}
		}]
	}
	goType: "R"
	goTypeBody: """
		struct {
			A [3]byte
		}
	"""
	inData: A: "abc"
	outData: inData
}

tests: goTypeStruct: {
	inSchema: {
		type: "record"
		name: "R"
		fields: [{
			name: "S1"
			type: {
				type: "record"
				name: "T"
				fields: [{
					name: "A"
					type: "int"
				}, {
					name: "B"
					type: "string"
				}]
			}
		}, {
			name: "S2"
			type: "T"
		}]
	}
	goType: "R"
	goTypeBody: """
		struct {
			S1 T
			S2 T
		}
		type T struct {
			A int
			B string
		}
	"""
	inData: {
		S1: {
			A: 12345
			B: "hello"
		}
		S2: {
			A: 999
			B: "b"
		}
	}
	outData: inData
}

tests: goTypeFieldsOmitted: {
	inSchema: {
		name: "R"
		type: "record"
		fields: []
	}
	goType: "R"
	goTypeBody: """
		struct {
			A int
			B string
			C [3]byte
			D map[string]string
			E []string
			F T
		}
		type T struct {
			A int
			B string
		}
	"""
	inData: {}
	outData: {
		A: 0
		B: ""
		C: "\u0000\u0000\u0000"
		D: {}
		E: []
		F: {
			A: 0
			B: ""
		}
	}
}

tests: goTypeProtobufRecord: {
	otherTests: """
	package goTypeProtobufRecord

	import "github.com/heetch/avro/internal/testtypes"

	type R = testtypes.MessageB
	"""
	goType: "R"
	inSchema: {
		name: "MessageB"
		type: "record"
		fields: []
	}
	inData: {}
	outData: {
		arble:    null
		selected: false
	}
}

tests: goTypeExternal: {
	inSchema: {
		name: "R"
		type: "record"
		fields: [{
			name: "F"
			type: {
				type:         "record"
				name:         "com.heetch.Message"
				"go.package": "github.com/heetch/avro/internal/testtypes"
				fields: [{
					name: "Metadata"
					type: {
						type: "record"
						name: "Metadata"
						fields: [{
							name: "CloudEvent"
							type: {
								type: "record"
								name: "CloudEvent"
								fields: [{
									name: "id"
									type: "string"
								}, {
									name: "source"
									type: "string"
								}, {
									name: "specversion"
									type: "string"
								}, {
									name: "time"
									type: {
										type:        "long"
										logicalType: "timestamp-micros"
									}
								}]
							}
						}]
					}
				}]
			}
		}, {
			name: "G"
			type: "com.heetch.CloudEvent"
		}, {
			name: "H"
			type: "string"
		}]
	}
	outSchema: inSchema
	inData: {
		F: Metadata: CloudEvent: {
			id:          "xid"
			source:      "xsource"
			specversion: "xspecversion"
			time:        1580486871000000
		}
		G: {
			id:          "yd"
			source:      "ysource"
			specversion: "yspecversion"
			time:        1580495933000000
		}
		H: "xh"
	}
	outData: inData
	otherTests: """
	package goTypeExternal
	import (
		"reflect"
		"testing"

		qt "github.com/frankban/quicktest"

		"github.com/heetch/avro/internal/testtypes"
	)

	var (
		messageType    = reflect.TypeOf(testtypes.Message{})
		cloudEventType = reflect.TypeOf(testtypes.CloudEvent{})
	)

	func TestCorrectTypes(t *testing.T) {
		c := qt.New(t)
		var r R
		c.Assert(reflect.TypeOf(r.F), qt.Equals, messageType)
		c.Assert(reflect.TypeOf(r.G), qt.Equals, cloudEventType)
	}
	"""
}

tests: goTypeCustomName: {
	otherTests: """
	package goTypeCustomName

	// Check that the types exist and look correct
	var (
		_ customName
		_ customEnum = customEnumA
		_ = customFixed{}[1]
	)

	"""
	goType: "customName"
	inSchema: {
		name:      "M"
		"go.name": "customName"
		type:      "record"
		fields: [{
			name: "E"
			type: {
				type:      "enum"
				"go.name": "customEnum"
				name:      "e"
				symbols: ["a", "b"]
			}
		}, {
			name: "F"
			type: {
				type:      "fixed"
				"go.name": "customFixed"
				name:      "f"
				size:      2
			}
		}]
	}
	outSchema: inSchema
	inData: {
		E: "b"
		F: "xz"
	}
	outData: inData
}
