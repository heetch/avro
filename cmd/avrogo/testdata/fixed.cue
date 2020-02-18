package roundtrip

tests: simpleFixed: {
	inSchema: {
		type: "record"
		name: "R"
		fields: [{
			name: "F"
			type: {
				type: "fixed"
				size: 5
				name: "five"
			}
		}]
	}
	outSchema: inSchema
	inData: F: "abcde"
	outData: inData
}
