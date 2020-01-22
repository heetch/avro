package roundtrip

tests: linkedList: {
	inSchema: {
		type: "record"
		name: "List"
		fields: [{
			name: "Item"
			type: "int"
		}, {
			name: "Next"
			type: ["null", "List"]
			default: null
		}]
	}
	outSchema: inSchema
	inData: {
		Item: 1234
		Next: List: {
			Item: 9999
			Next: null
		}
	}
	outData: inData
}

