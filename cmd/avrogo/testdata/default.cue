package roundtrip

// Some implementations don't like records with no fields in,
// so we'll use a record with one arbitrary field instead.
emptyRecord :: {
	type: "record"
	name: string
	fields: [{
		name:    "_"
		type:    "int"
		default: 0
	}]
}

emptyRecordData :: "_": 0

tests: primitiveDefaults: {
	inSchema: emptyRecord
	inSchema: name: "R"
	outSchema: {
		type: "record"
		name: "R"
		fields: [{
			name:    "int"
			type:    "int"
			default: 1111
		}, {
			name:    "long"
			type:    "long"
			default: 2222
		}, {
			name:    "string"
			type:    "string"
			default: "hello"
		}, {
			name:    "float"
			type:    "float"
			default: 1.5
		}, {
			name:    "double"
			type:    "double"
			default: 2.75
		}, {
			name:    "boolean"
			type:    "boolean"
			default: true
		}]
	}
	inData: emptyRecordData
	outData: {
		int:     1111
		long:    2222
		string:  "hello"
		float:   1.5
		double:  2.75
		boolean: true
	}
}

tests: arrayDefault: {
	inSchema: emptyRecord
	inSchema: name: "R"
	outSchema: {
		type: "record"
		name: "R"
		fields: [{
			name: "arrayOfInt"
			type: {
				type:  "array"
				items: "int"
			}
			default: [2, 3, 4]
		}]
	}
	inData: emptyRecordData
	outData: arrayOfInt: [2, 3, 4]
}

tests: mapDefault: {
	inSchema: emptyRecord
	inSchema: name: "R"
	outSchema: {
		type: "record"
		name: "R"
		fields: [{
			name: "mapOfInt"
			type: {
				type:   "map"
				values: "int"
			}
			default: {
				a: 2
				b: 5
				c: 99
			}
		}]
	}
	inData: emptyRecordData
	outData: mapOfInt: {
		a: 2
		b: 5
		c: 99
	}
}

tests: recordDefault: {
	inSchema: emptyRecord
	inSchema: name: "R"
	outSchema: {
		type: "record"
		name: "R"
		fields: [{
			name: "recordField"
			type: {
				type: "record"
				name: "Foo"
				fields: [{
					name: "F1"
					type: "int"
				}, {
					name: "F2"
					type: "string"
				}, {
					name:    "F3"
					type:    "string"
					default: "hello"
				}]
			}
			default: {
				F1: 44
				F2: "whee"
				F3: "ok"
			}
		}]
	}
	inData: emptyRecordData
	outData: recordField: {
		F1: 44
		F2: "whee"
		F3: "ok"
	}
}

tests: recordDefaultFieldNotProvided: {
	inSchema: emptyRecord
	inSchema: name: "R"
	outSchema: {
		type: "record"
		name: "R"
		fields: [{
			name: "recordField"
			type: {
				type: "record"
				name: "Foo"
				fields: [{
					name:    "F1"
					type:    "string"
				}, {
					name:    "F2"
					type:    "string"
					default: "hello"
				}]
			}
			default: {
				F1: ""
			}
		}]
	}
	generateError: #"avrogo: cannot generate code for schema.avsc: template: .*: executing "" at <\$.Ctx.RecordInfoLiteral>: error calling RecordInfoLiteral: cannot generate code for field recordField of record R: field "F2" of record Foo must be present in default value but is missing"#
}

tests: enumDefault: {
	inSchema: emptyRecord
	inSchema: name: "R"
	outSchema: {
		type: "record"
		name: "R"
		fields: [{
			name: "enumField"
			type: {
				type: "enum"
				name: "Foo"
				symbols: ["a", "b", "c"]
			}
			default: "b"
		}]
	}
	inData: emptyRecordData
	outData: enumField: "b"
}

tests: fixedDefault: {
	inSchema: emptyRecord
	inSchema: name: "R"
	outSchema: {
		type: "record"
		name: "R"
		fields: [{
			name: "fixedField"
			type: {
				type: "fixed"
				size: 5
				name: "five"
			}
			default: "hello"
		}]
	}
	inData: emptyRecordData
	outData: fixedField: "hello"
}
