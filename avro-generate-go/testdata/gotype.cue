package roundtrip

tests: simpleGoType: {
	outSchema: null
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
