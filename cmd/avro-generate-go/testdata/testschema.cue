package roundtrip

import (
	avroPkg "github.com/heetch/cue-schema/avro"
)

avro :: avroPkg
avro :: Metadata :: {
	"go.package"?: string
	"go.name"?:    string
	heetchmeta?: {
		commentary: string
		status:     string | *"active"
		partitions: int | *1
		topickey:   string
	}
}

tests: [_]: roundTripTest

tests: [name=_]: testName: name

roundTripTest :: {
	testName:    string
	inSchema:    avro.Schema
	outSchema?:  avro.Schema
	goType:      *outSchema.name | string
	goTypeBody?: string
	// generateError holds the error expected from invoking avro-generate-go.
	// If this is specified, there will be no generated test package.
	generateError?: string
	inData?:        _
	outData?:       _
	expectError?: [errorKind]: string
	otherTests?: string
	subtests: [name=_]: {
		testName: name
		inData:   _
		outData:  _
		expectError?: [errorKind]: string
	}
	if inData != _|_ {
		subtests: main: {
			"inData":  inData
			"outData": outData
			if expectError != _|_ {
				"expectError": expectError
			}
		}
	}
}

errorKind :: "unmarshal" | "marshal"
