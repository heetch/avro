package testutil

import (
	"encoding/json"
	"reflect"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/kr/pretty"
	"github.com/linkedin/goavro/v2"

	"github.com/heetch/avro"
)

type RoundTripTest struct {
	InSchema string
	GoType   interface{}
	Subtests []RoundTripSubtest
}

type ErrorType string

const (
	MarshalError   ErrorType = "marshal"
	UnmarshalError ErrorType = "unmarshal"
)

type RoundTripSubtest struct {
	TestName    string
	InDataJSON  string
	OutDataJSON string
	ExpectError map[ErrorType]string
}

func (test RoundTripTest) Test(t *testing.T) {
	c := qt.New(t)

	// Translate the JSON input data into binary using the input schema.
	inCodec, err := goavro.NewCodec(test.InSchema)
	c.Assert(err, qt.Equals, nil)
	for _, subtest := range test.Subtests {
		c.Run(subtest.TestName, func(c *qt.C) {
			subtest.runTest(c, test, inCodec)
		})
	}
}

func (subtest RoundTripSubtest) runTest(c *qt.C, test RoundTripTest, inCodec *goavro.Codec) {
	inNative, _, err := inCodec.NativeFromTextual([]byte(subtest.InDataJSON))
	c.Assert(err, qt.Equals, nil, qt.Commentf("inDataJSON: %q", subtest.InDataJSON))

	inData, err := inCodec.BinaryFromNative(nil, inNative)
	c.Assert(err, qt.Equals, nil)
	c.Logf("input data: %x", inData)

	sanity, _, err := inCodec.NativeFromBinary(inData)
	c.Assert(err, qt.Equals, nil)
	c.Logf("sanity: %s", pretty.Sprint(sanity))

	// Unmarshal the binary data into the Go type.
	x := reflect.New(reflect.TypeOf(test.GoType).Elem())
	inType, err := avro.ParseType(test.InSchema)
	c.Assert(err, qt.Equals, nil)
	_, err = avro.Unmarshal(inData, x.Interface(), inType)
	subtest.checkError(c, UnmarshalError, err, qt.Commentf("result data: %v", qt.Format(x.Interface())))
	c.Logf("unmarshaled: %s", pretty.Sprint(x.Interface()))

	// Marshal the data back into binary and then into
	// JSON, and check that it looks like we expect.
	outData, outSchema, err := avro.Marshal(x.Elem().Interface())
	subtest.checkError(c, MarshalError, err)
	c.Logf("output data: %x", outData)
	outCodec, err := goavro.NewCodec(outSchema.String())
	c.Assert(err, qt.Equals, nil, qt.Commentf("outSchema: %s", outSchema))
	native, remaining, err := outCodec.NativeFromBinary(outData)
	c.Assert(err, qt.Equals, nil)
	nativeJSON, err := outCodec.TextualFromNative(nil, native)
	c.Assert(err, qt.Equals, nil)
	c.Check(nativeJSON, qt.JSONEquals, json.RawMessage(subtest.OutDataJSON))
	c.Check(remaining, qt.HasLen, 0)
}

func (subtest RoundTripSubtest) checkError(c *qt.C, kind ErrorType, err error, extra ...interface{}) {
	if expectErr := subtest.ExpectError[kind]; expectErr != "" {
		args := append([]interface{}{expectErr}, extra...)
		c.Assert(err, qt.ErrorMatches, args...)
		c.SkipNow()
	}
	args := append([]interface{}{nil}, extra...)
	c.Assert(err, qt.Equals, args...)
}

func unmarshalJSON(c *qt.C, s string) interface{} {
	var x interface{}
	err := json.Unmarshal([]byte(s), &x)
	c.Assert(err, qt.Equals, nil)
	return x
}
