// Code generated by avrogen. DO NOT EDIT.

package simpleMap

import "github.com/rogpeppe/avro"

type R struct {
	M map[string]int
}

// AvroRecord implements the avro.AvroRecord interface.
func (R) AvroRecord() avro.RecordInfo {
	return avro.RecordInfo{
		Schema: `{"fields":[{"name":"M","type":{"type":"map","values":"int"}}],"name":"R","type":"record"}`,
	}
}

// TODO implement MarshalBinary and UnmarshalBinary methods?
