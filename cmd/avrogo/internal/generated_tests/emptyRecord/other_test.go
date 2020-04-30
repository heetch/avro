package emptyRecord

import (
	"reflect"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestNoFieldsInGeneratedStruct(t *testing.T) {
	c := qt.New(t)
	c.Assert(reflect.TypeOf(R{}).NumField(), qt.Equals, 0)
}
