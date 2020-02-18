package roundtrip

tests: timestampMicros: {
	inSchema: {
		type: "record"
		name: "R"
		fields: [{
			name: "T"
			type: {
				type:        "long"
				logicalType: "timestamp-micros"
			}
		}]
	}
	outSchema: inSchema
	inData: T: 1579176162000001
	outData: inData
}
