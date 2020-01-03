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
	inData: F: [1, "hello"]
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

tests: unionNullStringWithNull: {
	inSchema: {
		type: "record"
		name: "R"
		fields: [{
			name: "OptionalString"
			type: ["null", "string"]
		}]
	}
	outSchema: inSchema
	inData: OptionalString: null
	outData: inData
}

tests: unionNullStringReverseWithNull: {
	inSchema: {
		type: "record"
		name: "R"
		fields: [{
			name: "OptionalString"
			type: ["string", "null"]
		}]
	}
	outSchema: inSchema
	inData: OptionalString: null
	outData: inData
}

tests: unionNullStringWithString: {
	inSchema: {
		type: "record"
		name: "R"
		fields: [{
			name: "OptionalString"
			type: ["null", "string"]
		}]
	}
	outSchema: inSchema
	inData: OptionalString: string: "hello"
	outData: inData
}

tests: unionNullStringReverseWithString: {
	inSchema: {
		type: "record"
		name: "R"
		fields: [{
			name: "OptionalString"
			type: ["string", "null"]
		}]
	}
	outSchema: inSchema
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
