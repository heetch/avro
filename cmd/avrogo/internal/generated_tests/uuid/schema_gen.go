// Code generated by avrogen. DO NOT EDIT.

package uuid

import (
	"github.com/heetch/avro/avrotypegen"
	uuid "github.com/satori/go.uuid"
)

type R struct {
	T uuid.UUID
}

// AvroRecord implements the avro.AvroRecord interface.
func (R) AvroRecord() avrotypegen.RecordInfo {
	return avrotypegen.RecordInfo{
		Schema: `{"fields":[{"name":"T","type":{"logicalType":"uuid","type":"string"}}],"name":"R","type":"record"}`,
		Required: []bool{
			0: true,
		},
	}
}