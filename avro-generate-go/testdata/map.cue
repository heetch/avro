package roundtrip

tests: simpleMap: {
	inSchema: {
		type: "record"
		name: "R"
		fields: [{
			name: "M"
			type: {
				type:   "map"
				values: "int"
			}
		}]
	}
	outSchema: inSchema
	inData: M: {
		a: 32
		b: 54
	}
	outData: inData
}
