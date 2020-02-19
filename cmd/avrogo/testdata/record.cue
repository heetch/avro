package roundtrip

tests: largeRecord: {
	inSchema: {
		type:      "record"
		name:      "sample"
		namespace: "com.avro.test"
		doc:       "GoGen test"
		fields: [{
			name: "header"
			type: [
				"null",
				{
					type:      "record"
					name:      "Data0"
					namespace: "headerworks"
					doc:       "Common information related to the event which must be included in any clean event"
					fields: [{
						name: "uuid"
						type: [
							"null",
							{
								type:      "record"
								name:      "UUID0"
								namespace: "headerworks.datatype"
								doc:       "A Universally Unique Identifier, in canonical form in lowercase. Example: de305d54-75b4-431b-adb2-eb6b9e546014"
								fields: [
									{
										name:    "uuid"
										type:    "string"
										default: ""
									},
								]
							},
						]
						doc:     "Unique identifier for the event used for de-duplication and tracing."
						default: null
					}, {
						name: "hostname"
						type: [
							"null",
							"string",
						]
						doc:     "Fully qualified name of the host that generated the event that generated the data."
						default: null
					}, {
						name: "trace"
						type: [
							"null",
							{
								type: "record"
								name: "Trace0"
								doc:  "Trace0"
								fields: [
									{
										name: "traceId"
										type: [
											"null",
											"headerworks.datatype.UUID0",
										]
										doc:     "Trace Identifier"
										default: null
									},
								]
							},
						]
						doc:     "Trace information not redundant with this object"
						default: null
					}]
				},
			]
			doc:     "Core data information required for any event"
			default: null
		}, {
			name: "body"
			type: [
				"null",
				{
					type:      "record"
					name:      "Data1"
					namespace: "bodyworks"
					doc:       "Common information related to the event which must be included in any clean event"
					fields: [{
						name: "uuid"
						type: [
							"null",
							{
								type:      "record"
								name:      "UUID1"
								namespace: "bodyworks.datatype"
								doc:       "A Universally Unique Identifier, in canonical form in lowercase. Example: de305d54-75b4-431b-adb2-eb6b9e546014"
								fields: [{
									name:    "uuid"
									type:    "string"
									default: ""
								}]
							},
						]
						doc:     "Unique identifier for the event used for de-duplication and tracing."
						default: null
					}, {
						name: "hostname"
						type: [
							"null",
							"string",
						]
						doc:     "Fully qualified name of the host that generated the event that generated the data."
						default: null
					}, {
						name: "trace"
						type: [
							"null",
							{
								type: "record"
								name: "Trace1"
								doc:  "Trace1"
								fields: [{
									name: "traceId"
									type: [
										"null",
										"headerworks.datatype.UUID0",
									]
									doc:     "Trace Identifier"
									default: null
								}]
							},
						]
						doc:     "Trace information not redundant with this object"
						default: null
					}]
				},
			]
			doc:     "Core data information required for any event"
			default: null
		}]
	}
	goType:    "Sample"
	outSchema: inSchema
	inData: {
		header: "headerworks.Data0": {
			hostname: string: "myhost.com"
			trace: null
			uuid:  null
		}
		body: null
	}
	outData: inData
}

tests: duplicateRecord: {
	inSchema: {
		name: "R1"
		type: "record"
		fields: [{
			name: "F"
			type: {
				type: "record"
				name: "R2"
				fields: [{
					name: "A"
					type: "string"
				}]
			}
		}, {
			name: "G"
			type: "R2"
		}, {
			name: "H"
			type: "int"
		}]
	}
	outSchema: inSchema
	inData: {
		F: A: "hello"
		G: A: "goodbye"
		H: 99
	}
	outData: inData
}
