package roundtrip

import (
	"github.com/heetch/cue-schema/avro"
)

tests: [_]: roundTripTest

tests: [name=_]: testName: name

roundTripTest :: {
	testName:  string
	inSchema:  avro.Schema
	outSchema: avro.Schema
	goType:    string | *outSchema.name
	inData:    _
	outData:   _
}
