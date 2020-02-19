package goTypeExternal

import (
	"reflect"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/heetch/avro/internal/testtypes"
)

var (
	messageType    = reflect.TypeOf(testtypes.Message{})
	cloudEventType = reflect.TypeOf(testtypes.CloudEvent{})
)

func TestCorrectTypes(t *testing.T) {
	c := qt.New(t)
	var r R
	c.Assert(reflect.TypeOf(r.F), qt.Equals, messageType)
	c.Assert(reflect.TypeOf(r.G), qt.Equals, cloudEventType)
}
