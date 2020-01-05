package roundtrip

emptyRecord :: {
	type: "record"
	name: string
	fields: [{
		name:    "_"
		type:    "int"
		default: 0
	}]
}

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
	inData: {}
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
	inData: {}
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
	inData: {}
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
			}
		}]
	}
	inData: {}
	outData: recordField: {
		F1: 44
		F2: "whee"
		F3: "hello"
	}
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
	inData: {}
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
	inData: {}
	outData: fixedField: "hello"
}
