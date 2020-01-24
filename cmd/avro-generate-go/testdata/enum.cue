package roundtrip

tests: simpleEnum: {
	inSchema: {
		type: "record"
		name: "R"
		fields: [{
			name: "E"
			type: {
				type: "enum"
				name: "MyEnum"
				symbols: ["a", "b", "c"]
			}
		}]
	}
	outSchema: inSchema
	inData: E: "b"
	outData: inData
}

tests: simpleEnum: otherTests: """
	package simpleEnum
	import (
		"encoding/json"
		"testing"

		qt "github.com/frankban/quicktest"

		"github.com/heetch/avro"
	)

	func TestString(t *testing.T) {
		c := qt.New(t)
		c.Assert(MyEnumA.String(), qt.Equals, "a")
		c.Assert(MyEnumB.String(), qt.Equals, "b")
		c.Assert(MyEnumC.String(), qt.Equals, "c")
		c.Assert(MyEnum(-1).String(), qt.Equals, "MyEnum(-1)")
		c.Assert(MyEnum(3).String(), qt.Equals, "MyEnum(3)")
	}

	func TestMarshalText(t *testing.T) {
		c := qt.New(t)
		data, err := MyEnumA.MarshalText()
		c.Assert(err, qt.Equals, nil)
		c.Assert(string(data), qt.Equals, "a")

		_, err = MyEnum(-1).MarshalText()
		c.Assert(err, qt.ErrorMatches, `MyEnum value -1 is out of bounds`)

		_, err = MyEnum(3).MarshalText()
		c.Assert(err, qt.ErrorMatches, `MyEnum value 3 is out of bounds`)
	}

	func TestUnmarshalText(t *testing.T) {
		c := qt.New(t)
		var e MyEnum
		err := e.UnmarshalText([]byte("b"))
		c.Assert(err, qt.Equals, nil)
		c.Assert(e, qt.Equals, MyEnumB)

		// Check that it works OK with encoding/json too.
		var x struct {
			E MyEnum
		}
		err = json.Unmarshal([]byte(`{"E": "c"}`), &x)
		c.Assert(err, qt.Equals, nil)
		c.Assert(x.E, qt.Equals, MyEnumC)

		x.E = 0
		err = json.Unmarshal([]byte(`{"E": "unknown"}`), &x)
		c.Assert(err, qt.ErrorMatches, `unknown value "unknown" for MyEnum`)
	}

	func TestSchema(t *testing.T) {
		c := qt.New(t)
		at, err := avro.TypeOf(MyEnumA)
		c.Assert(err, qt.Equals, nil)
		c.Assert(at.String(), qt.JSONEquals, json.RawMessage(`{
			"type": "enum",
			"name": "MyEnum",
			"symbols": ["a", "b", "c"]
		}`))
	}
	"""
