package roundtrip

tests: goTypeSimple: {
	inSchema: {
		"name": "TestRecord",
		"type": "record",
		"fields": [{
			"name": "A",
			"type": "int"
		}, {
			"name": "B",
			"type": "int"
		}]
	}
	goType: "TestRecord"
	goTypeBody: """
		struct {
			B int
			A int
		}
	"""
	inData: {
		A: 1
		B: 2
	}
	outData: inData
}

tests: goTypePointer: {
	inSchema: {
		"name": "TestRecord",
		"type": "record",
		"fields": [{
			"name": "A",
			"type": ["null", "long"]
		}]
	}
	goType: "TestRecord"
	goTypeBody: """
		struct {
			A *int
		}
	"""
}

tests: goTypePointer: subtests: non_null: {
	inData: {
		A: long: 99
	}
	outData: inData
}

tests: goTypePointer: subtests: "null": {
	inData: {
		A: null
	}
	outData: inData
}
