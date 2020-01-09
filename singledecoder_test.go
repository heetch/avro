package avro_test

import (
	"context"
	"fmt"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/heetch/avro"
)

//go:generate avro-generate-go -t -p avro_test testschema1.avsc

func TestSingleDecoder(t *testing.T) {
	c := qt.New(t)
	dec := avro.NewSingleDecoder(memRegistry{
		1: `{
	"name": "TestRecord",
	"type": "record",
	"fields": [{
		"name": "B",
		"type": {
		    "type": "int"
		}
	}, {
		"name": "A",
		"type": {
		    "type": "int"
		}
	}]
}`,
		2: `{
	"name": "TestRecord",
	"type": "record",
	"fields": [{
		"name": "B",
		"type": {
		    "type": "int"
		}
	}]
}`,
		3: `{
	"name": "TestRecord",
	"type": "record",
	"fields": [{
		"name": "A",
		"type": {
		    "type": "int"
		}
	}]
}`,
		13: `{
	"name": "TestRecord",
	"type": "record",
	"fields": [{
		"name": "A",
		"type": {
		    "type": "string"
		}
	}]
}`,
	})
	data, _, err := avro.Marshal(TestRecord{A: 40, B: 20})
	c.Assert(err, qt.Equals, nil)
	c.Logf("data: %d", data)
	var x TestRecord
	// In the byte slice below:
	// 	1: the schema id
	//	40: B=20 (zig-zag encoded)
	//	80: A=40 (ditto)
	_, err = dec.Unmarshal(context.Background(), []byte{1, 40, 80}, &x)
	c.Assert(err, qt.Equals, nil)
	c.Assert(x, qt.Equals, TestRecord{A: 40, B: 20})

	// Check the record compatibility stuff is working by reading from a
	// record written with less fields (note: the default value for A is 42).
	var x1 TestRecord
	_, err = dec.Unmarshal(context.Background(), []byte{2, 80}, &x1)
	c.Assert(err, qt.Equals, nil)
	c.Assert(x1, qt.Equals, TestRecord{A: 42, B: 40})

	// There's no default value for A, so it doesn't work that way around.
	var x2 TestRecord
	_, err = dec.Unmarshal(context.Background(), []byte{3, 80}, &x2)
	c.Assert(err, qt.ErrorMatches, `cannot unmarshal: cannot create decoder: Incompatible schemas: field B in reader is not present in writer and has no default value`)
}

// memRegistry implements DecodingRegistry and EncodingRegistry by associating a single-byte
// schema ID with schemas.
type memRegistry map[int64]string

func (m memRegistry) DecodeSchemaID(msg []byte) (int64, []byte) {
	if len(msg) < 1 {
		return 0, nil
	}
	return int64(msg[0]), msg[1:]
}

func (m memRegistry) SchemaForID(ctx context.Context, id int64) (string, error) {
	s, ok := m[id]
	if !ok {
		return "", fmt.Errorf("schema not found for id %d", id)
	}
	return s, nil
}

func (m memRegistry) AppendSchemaID(buf []byte, id int64) []byte {
	if id < 0 || id > 256 {
		panic("schema ID out of range")
	}
	return append(buf, byte(id))
}

func (m memRegistry) IDForSchema(ctx context.Context, schema string) (int64, error) {
	for id, s := range m {
		if s == schema {
			return id, nil
		}
	}
	return 0, fmt.Errorf("schema not found")
}
