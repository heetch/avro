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

tests: uuid: {
       inSchema: {
                 type: "record"
                 name: "R"
                 fields: [{
                         name: "T"
                         type: {
                               type:        "string"
                               logicalType: "uuid"
                         }
                 }]
       }
       outSchema: inSchema
       inData: T: "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
       outData: inData
}

tests: invalidUUID: {
       inSchema: {
                 type: "record"
                 name: "R"
                 fields: [{
                         name: "T"
                         type: {
                               type:        "string"
                               logicalType: "uuid"
                         }
                 }]
       }
       outSchema: inSchema
       inData: T: "invalid_uuid"
       outData: null
       expectError: unmarshal: "invalid UUID in Avro encoding: invalid UUID length: 12"
}

tests: durationNanos: {
	inSchema: {
		type: "record"
		name: "R"
		fields: [{
			name: "D"
			type: {
				type:        "long"
				logicalType: "duration-nanos"
			}
		}]
	}
	outSchema: inSchema
	inData: D: 15000000000
	outData: inData
}
