package roundtrip

tests: multiSchema: {
	inSchema: {
		type: "record"
		name: "R"
		fields: [{
			name: "F"
			type: {
				type: "record"
				name: "S"
				fields: [{
					name: "G"
					type: "int"
				}]
			}
		}]
	}
	outSchema: {
		type: "record"
		name: "R"
		fields: [{
			name: "F"
			type: "S"
		}]
	}
	extraSchemas: [{
		type: "record"
		name: "S"
		fields: [{
			name: "G"
			type: "int"
		}]
	}]
	inData: F: G: 99
	outData: inData
}

tests: multiSchemaMutualRecursive: {
	inSchema: {
		type: "record"
		name: "R"
		fields: [{
			name: "F"
			default: []
			type: {
				type: "array"
				items: {
					type: "record"
					name: "S"
					fields: [{
						name:    "Data"
						type:    "string"
						default: ""
					}, {
						name: "Child"
						type: ["null", "R"]
						default: null
					}]
				}
			}
		}]
	}
	outSchema: {
		type: "record"
		name: "R"
		fields: [{
			name: "F"
			type: {
				type:  "array"
				items: "S"
			}
			default: []
		}]
	}
	extraSchemas: [{
		type: "record"
		name: "S"
		fields: [{
			name:    "Data"
			type:    "string"
			default: ""
		}, {
			name: "Child"
			type: ["null", "R"]
			default: null
		}]
	}]
	inData: F: [{
		Data: "hello"
		Child: R: F: [{
			Data: "goodbye"
			Child: null
		}]
	}, {
		Data: "whee"
		Child: null
	}]
	outData: inData
}

tests: multiSchemaExternalType: {
	inSchema: {
		name: "R"
		type: "record"
		fields: [{
			name: "F"
			type: {
				type:         "record"
				name:         "com.heetch.Message"
				"go.package": "github.com/heetch/avro/internal/testtypes"
				fields: [{
					name: "Metadata"
					type: {
						type: "record"
						name: "Metadata"
						fields: [{
							name: "CloudEvent"
							type: {
								type: "record"
								name: "CloudEvent"
								fields: [{
									name: "id"
									type: "string"
								}, {
									name: "source"
									type: "string"
								}, {
									name: "specversion"
									type: "string"
								}, {
									name: "time"
									type: {
										type:        "long"
										logicalType: "timestamp-micros"
									}
								}]
							}
						}]
					}
				}]
			}
		}, {
			name: "G"
			type: "com.heetch.CloudEvent"
		}, {
			name: "H"
			type: "string"
		}]
	}
	outSchema: {
		name: "R"
		type: "record"
		fields: [{
			name: "F"
			type: "com.heetch.Message"
		}, {
			name: "G"
			type: "com.heetch.CloudEvent"
		}, {
			name: "H"
			type: "string"
		}]
	}
	extraSchemas: [{
		type:         "record"
		name:         "com.heetch.Message"
		"go.package": "github.com/heetch/avro/internal/testtypes"
		fields: [{
			name: "Metadata"
			type: {
				type: "record"
				name: "Metadata"
				fields: [{
					name: "CloudEvent"
					type: {
						type: "record"
						name: "CloudEvent"
						fields: [{
							name: "id"
							type: "string"
						}, {
							name: "source"
							type: "string"
						}, {
							name: "specversion"
							type: "string"
						}, {
							name: "time"
							type: {
								type:        "long"
								logicalType: "timestamp-micros"
							}
						}]
					}
				}]
			}
		}]
	}]
	inData: {
		F: Metadata: CloudEvent: {
			id:          "xid"
			source:      "xsource"
			specversion: "xspecversion"
			time:        1580486871000000
		}
		G: {
			id:          "yd"
			source:      "ysource"
			specversion: "yspecversion"
			time:        1580495933000000
		}
		H: "xh"
	}
	outData: inData
}

tests: multiSchemaConflictingDefinitions: {
	inSchema: {
		type: "record"
		name: "R"
		fields: [{
			name: "F"
			type: {
				type: "record"
				name: "S"
				fields: [{
					name: "G"
					type: "int"
				}]
			}
		}]
	}
	outSchema: {
		type: "record"
		name: "R"
		fields: [{
			name: "F"
			type: {
				type: "record"
				name: "S"
				fields: [{
					name: "G"
					type: "int"
				}]
			}
		}]
	}
	extraSchemas: [{
		type: "record"
		name: "S"
		fields: [{
			name: "G"
			// Note: different type.
			type: "string"
		}]
	}]
	generateError: "avrogen: cannot parse schema in .*: Conflicting definitions for S"
}
