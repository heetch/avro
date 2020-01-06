package roundtrip

tests: unionInOut: {
	inSchema: {
		type: "record"
		name: "PrimitiveUnionTestRecord"
		fields: [{
			name: "UnionField"
			type: ["int", "long", "float", "double", "string", "boolean", "null"]
			default: 1234
		}]
	}
	outSchema: inSchema
	inData: UnionField: int: 999
	outData: inData
}

tests: unionIntVsLong: {
	inSchema: {
		type: "record"
		name: "R"
		fields: [{
			name: "F"
			type: ["int", "string"]
			default: 1234
		}]
	}
	outSchema: {
		type: "record"
		name: "R"
		fields: [{
			name: "F"
			type: ["long", "int", "string"]
			default: 1234
		}]
	}
	inData: F: int: 999
	outData: F: long: 999
}

tests: arrayOfUnion: {
	inSchema: {
		type: "record"
		name: "R"
		fields: [{
			name: "F"
			type: {
				type: "array"
				items: [
					"int",
					"string",
				]
			}
		}]
	}
	outSchema: inSchema
	inData: F: [{int: 1}, {string: "hello"}]
	outData: inData
}

tests: unionToScalar: {
	inSchema: tests.unionInOut.inSchema
	outSchema: {
		type: "record"
		name: "PrimitiveUnionTestRecord"
		fields: [{
			name:    "UnionField"
			type:    "int"
			default: 1234
		}]
	}
	inData: UnionField: int: 999
	outData: UnionField: 999
}

tests: unionNullString: {
	inSchema: {
		type: "record"
		name: "R"
		fields: [{
			name: "OptionalString"
			type: ["null", "string"]
		}]
	}
	outSchema: inSchema
}

tests: unionNullString: subtests: withNull: {
	inData: OptionalString: null
	outData: inData
}

tests: unionNullString: subtests: withString: {
	inData: OptionalString: string: "hello"
	outData: inData
}

tests: unionNullStringReverse: {
	inSchema: {
		type: "record"
		name: "R"
		fields: [{
			name: "OptionalString"
			type: ["string", "null"]
		}]
	}
	outSchema: inSchema
}

tests: unionNullStringReverse: subtests: withNull: {
	inData: OptionalString: null
	outData: inData
}

tests: unionNullStringReverse: subtests: withString: {
	inData: OptionalString: string: "hello"
	outData: inData
}

tests: sharedUnion: {
	inSchema: {
		type: "record"
		name: "R"
		fields: [{
			name: "A"
			type: ["int", "string", "float"]
		}, {
			name: "B"
			type: ["int", "string", "float"]
		}]
	}
	outSchema: inSchema
	inData: {
		A: int:    244
		B: string: "hello"
	}
	outData: inData
}
