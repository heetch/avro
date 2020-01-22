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
}

tests: simpleArray: subtests: non_empty: {
	inData: A: [23, 57, 444]
	outData: inData
}

tests: simpleArray: subtests: empty: {
	inData: A: []
	outData: inData
}
