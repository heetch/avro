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

tests: linkedListThenSomethingElse: {
	inSchema: {
		type: "record"
		name: "R"
		fields: [{
			name: "L"
			type: {
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
		}, {
			name: "M"
			type: "int"
		}]
	}
	outSchema: inSchema
	inData: {
		L: {
			Item: 1234
			Next: List: {
				Item: 9999
				Next: null
			}
		}
		M: 1234
	}
	outData: inData
}
