package roundtrip

tests: cloudEvent: {
	DomainName :: "someDomain"
	EventName ::  "someEvent"
	Version ::    "v9.9.99"
	inSchema: {
		type: "record"
		name: "com.heetch.\(DomainName).\(EventName)"
		heetchmeta: {
			commentary: "This Schema describes version \(Version) of the event \(EventName) from the domain \(DomainName)."
			topickey:   "\(DomainName).\(EventName).\(Version)"
		}
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
		}, {
			name: "other"
			type: "string"
		}]
	}
	goType: "Message"
	outSchema: {
		type: "record"
		name: "com.heetch.Message"
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
	inData: {
		Metadata: CloudEvent: {
			id:          "id1"
			source:      "source1"
			specversion: "someversion"
			time:        1580392724000000
		}
		other: "some other data"
	}
	outData: Metadata: inData.Metadata
}
