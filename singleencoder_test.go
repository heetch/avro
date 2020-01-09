package avro_test

import (
	"context"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/heetch/avro"
)

func TestSingleEncoder(t *testing.T) {
	c := qt.New(t)
	avroType, err := avro.TypeOf(TestRecord{})
	c.Assert(err, qt.Equals, nil)
	registry := memRegistry{
		1: avroType.String(),
	}
	enc := avro.NewSingleEncoder(registry)
	data, err := enc.Marshal(context.Background(), TestRecord{A: 20, B: 34})
	c.Assert(err, qt.Equals, nil)
	c.Assert(data, qt.DeepEquals, []byte{1, 40, 68})

	// Check that we can decode it again.
	var x TestRecord
	dec := avro.NewSingleDecoder(registry)
	_, err = dec.Unmarshal(context.Background(), data, &x)
	c.Assert(err, qt.Equals, nil)
	c.Assert(x, qt.DeepEquals, TestRecord{A: 20, B: 34})
}
