// Code generated by avrogen. DO NOT EDIT.

package avro_test

import (
	"fmt"
	"github.com/heetch/avro/avrotypegen"
	"strconv"
)

type EnumC int

const (
	EnumCX EnumC = iota
	EnumCY
	EnumCZ
)

var _EnumC_strings = []string{
	"x",
	"y",
	"z",
}

// String returns the textual representation of EnumC.
func (e EnumC) String() string {
	if e < 0 || int(e) >= len(_EnumC_strings) {
		return "EnumC(" + strconv.FormatInt(int64(e), 10) + ")"
	}
	return _EnumC_strings[e]
}

// MarshalText implements encoding.TextMarshaler
// by returning the textual representation of EnumC.
func (e EnumC) MarshalText() ([]byte, error) {
	if e < 0 || int(e) >= len(_EnumC_strings) {
		return nil, fmt.Errorf("EnumC value %d is out of bounds", e)
	}
	return []byte(_EnumC_strings[e]), nil
}

// UnmarshalText implements encoding.TextUnmarshaler
// by expecting the textual representation of EnumC.
func (e *EnumC) UnmarshalText(data []byte) error {
	// Note for future: this could be more efficient.
	for i, s := range _EnumC_strings {
		if string(data) == s {
			*e = EnumC(i)
			return nil
		}
	}
	return fmt.Errorf("unknown value %q for EnumC", data)
}

type TestNewRecord struct {
	A int
	B int
	C *EnumC
}

// AvroRecord implements the avro.AvroRecord interface.
func (TestNewRecord) AvroRecord() avrotypegen.RecordInfo {
	return avrotypegen.RecordInfo{
		Schema: `{"fields":[{"default":42,"name":"A","type":{"type":"int"}},{"name":"B","type":{"type":"int"}},{"default":null,"name":"C","type":["null",{"name":"EnumC","symbols":["x","y","z"],"type":"enum"}]}],"name":"TestNewRecord","type":"record"}`,
		Required: []bool{
			1: true,
		},
		Defaults: []func() interface{}{
			0: func() interface{} {
				return 42
			},
		},
	}
}
