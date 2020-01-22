package avro

Protocol :: {
	protocol:   string
	namespace?: string
	doc?:       string
	types: [... Schema]
	messages?: {
		[string]: Message
	}
}

Message :: {
	doc?: string
	request: [... Field]
	response: Schema
	errors?: [... Schema]
	if (response & "null") != _|_ {
		"one-way"?: bool
	}
}
