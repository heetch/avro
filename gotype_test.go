package avro_test

import (
	"encoding/json"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"

	"github.com/heetch/avro"
	"github.com/heetch/avro/internal/testtypes"
)

func TestSimpleGoType(t *testing.T) {
	test := func(t *testing.T) {
		c := qt.New(t)
		data, wType, err := avro.Marshal(TestRecord{
			A: 1,
			B: 2,
		})
		c.Assert(err, qt.Equals, nil)
		type TestRecord struct {
			B int
			A int
		}
		var x TestRecord
		_, err = avro.Unmarshal(data, &x, wType)
		c.Assert(err, qt.Equals, nil)
		c.Assert(x, qt.Equals, TestRecord{
			A: 1,
			B: 2,
		})
	}
	// Run the test twice to test caching.
	test(t)
	test(t)
}

func TestGoTypeWithTime(t *testing.T) {
	c := qt.New(t)
	type R struct {
		T time.Time
	}
	data, wType, err := avro.Marshal(R{
		T: time.Date(2020, 1, 15, 18, 47, 8, 888888777, time.UTC),
	})
	c.Assert(err, qt.Equals, nil)
	var x R
	_, err = avro.Unmarshal(data, &x, wType)
	c.Assert(err, qt.Equals, nil)
	c.Assert(x, qt.DeepEquals, R{
		T: time.Date(2020, 1, 15, 18, 47, 8, 888888000, time.UTC),
	})
}

func TestGoTypeWithZeroTime(t *testing.T) {
	c := qt.New(t)
	type R struct {
		T time.Time
	}
	// The zero time marshals as the zero unix time.
	data, wType, err := avro.Marshal(R{})
	c.Assert(err, qt.Equals, nil)
	{
		type R struct {
			T int
		}
		var x R
		_, err = avro.Unmarshal(data, &x, wType)
		c.Assert(err, qt.Equals, nil)
		c.Assert(x, qt.DeepEquals, R{})
	}
}

func TestGoTypeStringerEnum(t *testing.T) {
	c := qt.New(t)
	type R struct {
		E testtypes.Enum
	}
	at, err := avro.TypeOf(R{})
	c.Assert(err, qt.Equals, nil)
	c.Assert(at.String(), qt.JSONEquals, json.RawMessage(`{
		"type": "record",
		"name": "R",
		"fields": [{
			"name": "E",
			"default": "One",
			"type": {
				"type": "enum",
				"name": "Enum",
				"symbols": ["One", "Two", "Three"]
			}
		}]
	}`))
}

func TestGoTypeEnumOOBPanic(t *testing.T) {
	c := qt.New(t)
	type R struct {
		E OOBPanicEnum
	}
	at, err := avro.TypeOf(R{})
	c.Assert(err, qt.Equals, nil)
	c.Assert(at.String(), qt.JSONEquals, json.RawMessage(`{
		"type": "record",
		"name": "R",
		"fields": [{
			"name": "E",
			"default": "a",
			"type": {
				"type": "enum",
				"name": "OOBPanicEnum",
				"symbols": ["a", "b"]
			}
		}]
	}`))
}

func TestGoTypeEnumOOBEmpty(t *testing.T) {
	c := qt.New(t)
	type R struct {
		E OOBEmptyEnum
	}
	at, err := avro.TypeOf(R{})
	c.Assert(err, qt.Equals, nil)
	c.Assert(at.String(), qt.JSONEquals, json.RawMessage(`{
		"type": "record",
		"name": "R",
		"fields": [{
			"name": "E",
			"default": "a",
			"type": {
				"type": "enum",
				"name": "OOBEmptyEnum",
				"symbols": ["a", "b"]
			}
		}]
	}`))
}

type OOBPanicEnum int

var enumValues = []string{"a", "b"}

func (e OOBPanicEnum) String() string {
	return enumValues[e]
}

type OOBEmptyEnum int

func (e OOBEmptyEnum) String() string {
	if e < 0 || int(e) >= len(enumValues) {
		return ""
	}
	return enumValues[e]
}
