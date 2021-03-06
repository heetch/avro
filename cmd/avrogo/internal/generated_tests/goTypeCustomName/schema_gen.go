// Code generated by avrogen. DO NOT EDIT.

package goTypeCustomName

import (
	"fmt"
	"github.com/heetch/avro/avrotypegen"
	"strconv"
)

type customName struct {
	E customEnum
	F customFixed
}

// AvroRecord implements the avro.AvroRecord interface.
func (customName) AvroRecord() avrotypegen.RecordInfo {
	return avrotypegen.RecordInfo{
		Schema: `{"fields":[{"name":"E","type":{"go.name":"customEnum","name":"e","symbols":["a","b"],"type":"enum"}},{"name":"F","type":{"go.name":"customFixed","name":"f","size":2,"type":"fixed"}}],"go.name":"customName","name":"M","type":"record"}`,
		Required: []bool{
			0: true,
			1: true,
		},
	}
}

type customEnum int

const (
	customEnumA customEnum = iota
	customEnumB
)

var _customEnum_strings = []string{
	"a",
	"b",
}

// String returns the textual representation of customEnum.
func (e customEnum) String() string {
	if e < 0 || int(e) >= len(_customEnum_strings) {
		return "customEnum(" + strconv.FormatInt(int64(e), 10) + ")"
	}
	return _customEnum_strings[e]
}

// MarshalText implements encoding.TextMarshaler
// by returning the textual representation of customEnum.
func (e customEnum) MarshalText() ([]byte, error) {
	if e < 0 || int(e) >= len(_customEnum_strings) {
		return nil, fmt.Errorf("customEnum value %d is out of bounds", e)
	}
	return []byte(_customEnum_strings[e]), nil
}

// UnmarshalText implements encoding.TextUnmarshaler
// by expecting the textual representation of E.
func (e *customEnum) UnmarshalText(data []byte) error {
	// Note for future: this could be more efficient.
	for i, s := range _customEnum_strings {
		if string(data) == s {
			*e = customEnum(i)
			return nil
		}
	}
	return fmt.Errorf("unknown value %q for customEnum", data)
}

type customFixed [2]byte
