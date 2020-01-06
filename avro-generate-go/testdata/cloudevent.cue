package roundtrip

tests: cloudEvent: {
	inSchema: {
		type:      "record"
		name:      "SomeEvent"
		namespace: "foo"
		fields: [{
			name: "Metadata"
			type: {
				type:      "record"
				name:      "Metadata"
				namespace: "avro.apache.org"
				fields: [{
					name: "id"
					type: "string"
				}, {
					name: "source"
					type: "string"
				}, {
					name: "time"
					type: "long"
					// logicalType: "timestamp-micros"
				}]
			}
		}, {
			name: "other"
			type: "string"
		}]
	}
	outSchema: {
		name:      "SomeEvent"
		namespace: "bar"
		type: "record"
		fields: [{
			name: "Metadata"
			type: {
				type:      "record"
				name:      "Metadata"
				namespace: "avro.apache.org"
				fields: [{
					name: "id"
					type: "string"
				}, {
					name: "source"
					type: "string"
				}, {
					name: "time"
					type: "long"
					// logicalType: "timestamp-micros"
				}]
			}
		}]
	}
	inData: {
		Metadata: {
			id:     "id1"
			source: "source1"
			time:   12345
		}
		other: "some other data"
	}
	outData: Metadata: inData.Metadata
}
