package roundtrip

tests: simpleArray: {
	inSchema: {
		type: "record"
		name: "R"
		fields: [{
			name: "A"
			type: {
				type:  "array"
				items: "int"
			}
		}]
	}
	outSchema: inSchema
	inData: A: [23, 57, 444]
	outData: inData
}
