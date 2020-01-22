package roundtrip

tests: simpleEnum: {
	inSchema: {
		type: "record"
		name: "R"
		fields: [{
			name: "E"
			type: {
				type: "enum"
				name: "MyEnum"
				symbols: ["a", "b", "c"]
			}
		}]
	}
	outSchema: inSchema
	inData: E: "b"
	outData: inData
}
