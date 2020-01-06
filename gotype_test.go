package avro_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/heetch/avro"
)

func TestSimpleGoType(t *testing.T) {
	c := qt.New(t)
	wSchema := TestRecord{}.AvroRecord().Schema
	data, err := avro.Marshal(TestRecord{
		A: 1,
		B: 2,
	})
	c.Assert(err, qt.Equals, nil)
	type TestRecord struct {
		B int
		A int
	}
	wType, err := avro.ParseType(wSchema)
	c.Assert(err, qt.Equals, nil)
	var x TestRecord
	_, err = avro.Unmarshal(data, &x, wType)
	c.Assert(err, qt.Equals, nil)
	c.Assert(x, qt.Equals, TestRecord{
		A: 1,
		B: 2,
	})
}
