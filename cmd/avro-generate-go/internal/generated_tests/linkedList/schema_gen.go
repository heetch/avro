// Code generated by avrogen. DO NOT EDIT.

package linkedList

import (
	"github.com/heetch/avro/avrotypegen"
)

type List struct {
	Item int
	Next *List
}

// AvroRecord implements the avro.AvroRecord interface.
func (List) AvroRecord() avrotypegen.RecordInfo {
	return avrotypegen.RecordInfo{
		Schema: `{"fields":[{"name":"Item","type":"int"},{"default":null,"name":"Next","type":["null","List"]}],"name":"List","type":"record"}`,
		Required: []bool{
			0: true,
		},
	}
}