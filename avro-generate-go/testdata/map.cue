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
}

tests: simpleMap: subtests: non_empty: {
	inData: M: {
		a: 32
		b: 54
	}
	outData: inData
}

tests: simpleMap: subtests: empty: {
	inData: M: {}
	outData: inData
}
