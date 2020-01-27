package avro_test

import (
	"encoding/json"
	"sync"
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
		if !c.Check(err, qt.Equals, nil) {
			return
		}
		type TestRecord struct {
			B int
			A int
		}
		var x TestRecord
		_, err = avro.Unmarshal(data, &x, wType)
		if !c.Check(err, qt.Equals, nil) {
			return
		}
		c.Check(x, qt.Equals, TestRecord{
			A: 1,
			B: 2,
		})
	}
	// Run the test twice concurrently to test caching and potential race conditions.
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			test(t)
		}()
	}
	wg.Wait()
}

func TestGoTypeWithOmittedFields(t *testing.T) {
	c := qt.New(t)
	type R struct {
		omit1 int
		A     int
		omit2 int
		Omit3 int `json:"-"`
		B     string
	}
	data, wType, err := avro.Marshal(R{
		A: 1,
		B: "hello",
	})
	c.Assert(err, qt.Equals, nil)
	c.Assert(wType.String(), qt.JSONEquals, json.RawMessage(`{
		"type": "record",
		"name": "R",
		"fields": [{
			"default": 0,
			"name": "A",
			"type": "long"
		}, {
			"default": "",
			"name": "B",
			"type": "string"
		}]
	}`))

	var r R
	_, err = avro.Unmarshal(data, &r, wType)
	c.Assert(err, qt.Equals, nil)
	c.Assert(r, qt.Equals, R{A: 1, B: "hello"})
}

func TestGoTypeWithJSONTags(t *testing.T) {
	c := qt.New(t)
	type R struct {
		A int    `json:"something"`
		B string `json:"other,omitempty"`
	}
	data, wType, err := avro.Marshal(R{
		A: 1,
		B: "hello",
	})
	c.Assert(err, qt.Equals, nil)
	c.Assert(wType.String(), qt.JSONEquals, json.RawMessage(`{
		"type": "record",
		"name": "R",
		"fields": [{
			"default": 0,
			"name": "something",
			"type": "long"
		}, {
			"default": "",
			"name": "other",
			"type": "string"
		}]
	}`))

	var r R
	_, err = avro.Unmarshal(data, &r, wType)
	c.Assert(err, qt.Equals, nil)
	c.Assert(r, qt.Equals, R{A: 1, B: "hello"})
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

func TestProtobufGeneratedType(t *testing.T) {
	c := qt.New(t)
	at, err := avro.TypeOf(testtypes.MessageB{})
	c.Assert(err, qt.Equals, nil)
	c.Assert(at.String(), qt.JSONEquals, json.RawMessage(`{
		"type": "record",
		"name": "MessageB",
		"fields": [{
			"name": "arble",
			"default": null,
			"type": [
				"null", {
					"type": "record",
					"name": "MessageA",
					"fields": [{
						"name": "id",
						"default": "",
						"type": "string"
					}, {
						"name": "label",
						"default": "LABEL_FOR_ZERO",
						"type": {
							"type": "enum",
							"name": "LabelFor",
							"symbols": [
								"LABEL_FOR_ZERO",
								"LABEL_FOR_ONE",
								"LABEL_FOR_TWO",
								"LABEL_FOR_THREE"
							]
						}
					}, {
						"name": "foo_url",
						"type": "string",
						"default": ""
					}, {
						"name": "enabled",
						"default": false,
						"type": "boolean"
					}]
				}
			]
		}, {
			"name": "selected",
			"default": false,
			"type": "boolean"
		}]
	}`))
}

func TestUnmarshalDoesNotCorruptData(t *testing.T) {
	c := qt.New(t)
	type R struct {
		A *string
		B *string
	}
	type T struct {
		R R
	}
	x := T{
		R: R{
			A: newString("hello"),
			B: newString("goodbye"),
		},
	}
	data, at, err := avro.Marshal(x)
	c.Assert(err, qt.Equals, nil)
	origData := data
	var x1 T
	_, err = avro.Unmarshal(data, &x1, at)
	c.Assert(err, qt.Equals, nil)
	_, err = avro.Unmarshal(data, &x1, at)
	c.Assert(err, qt.Equals, nil)
	c.Assert(data, qt.DeepEquals, []byte(origData))
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
