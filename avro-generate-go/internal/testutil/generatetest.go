package testutil

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/kr/pretty"
	"github.com/linkedin/goavro/v2"

	"github.com/rogpeppe/avro"
)

type RoundTripTest struct {
	InDataJSON  string
	OutDataJSON string
	InSchema    string
	OutSchema   string
	GoType      interface{}
}

func (test RoundTripTest) Test(t *testing.T) {
	c := qt.New(t)
	// Translate the JSON input data into binary using the input schema.
	inCodec, err := goavro.NewCodec(test.InSchema)
	c.Assert(err, qt.Equals, nil)
	inNative, _, err := inCodec.NativeFromTextual([]byte(test.InDataJSON))
	c.Assert(err, qt.Equals, nil)

	inData, err := inCodec.BinaryFromNative(nil, inNative)
	c.Assert(err, qt.Equals, nil)
	c.Logf("input data: %x", inData)

	sanity, _, err := inCodec.NativeFromBinary(inData)
	c.Assert(err, qt.Equals, nil)
	pretty.Println("sanity: ", sanity)

	// Unmarshal the binary data into the Go type.
	x := reflect.New(reflect.TypeOf(test.GoType).Elem())
	err = avro.Unmarshal(inData, x.Interface(), test.InSchema)
	c.Assert(err, qt.Equals, nil)
	pretty.Println("unmarshaled: ", x.Interface())

	// Marshal the data back into binary and then into
	// JSON, and check that it looks like we expect.
	outData, err := avro.Marshal(x.Elem().Interface())
	c.Assert(err, qt.Equals, nil)
	c.Logf("output data: %x", outData)
	outCodec, err := goavro.NewCodec(test.OutSchema)
	c.Assert(err, qt.Equals, nil)
	native, remaining, err := outCodec.NativeFromBinary(outData)
	c.Assert(err, qt.Equals, nil)
	// Marshal the native value to JSON so that we don't get type clashes on
	// numeric types (don't use json.Marshal because the goavro native encoding
	// encodes bytes values as []byte which doesn't encode the same as
	// the Avro JSON format.

	nativeJSON, err := marshalNative(native)
	c.Assert(err, qt.Equals, nil)
	c.Check(nativeJSON, qt.JSONEquals, json.RawMessage(test.OutDataJSON))
	c.Check(remaining, qt.HasLen, 0)
}

func unmarshalJSON(c *qt.C, s string) interface{} {
	var x interface{}
	err := json.Unmarshal([]byte(s), &x)
	c.Assert(err, qt.Equals, nil)
	return x
}

func marshalNative(x interface{}) ([]byte, error) {
	return json.Marshal(translateNative(x))
}

// translateNative translates from the "native" form used by goavro,
// to a form that marshals correctly as JSON. Specifically
// byte slices get transformed to the unicode mapping
// specified in the schema.
func translateNative(x interface{}) interface{} {
	switch x := x.(type) {
	case map[string]interface{}:
		x1 := make(map[string]interface{})
		for k, v := range x {
			x1[k] = translateNative(v)
		}
		return x1
	case []interface{}:
		x1 := make([]interface{}, len(x))
		for i, v := range x {
			x1[i] = translateNative(v)
		}
		return x1
	case []byte:
		var buf strings.Builder
		for _, b := range x {
			buf.WriteRune(rune(b))
		}
		return buf.String()
	case nil:
		return nil
	default:
		xv := reflect.TypeOf(x)
		switch xv.Kind() {
		case reflect.Array, reflect.Slice, reflect.Struct, reflect.Map:
			panic(fmt.Errorf("unexpected type in native object: %T", x))
		}
		return x
	}
}
