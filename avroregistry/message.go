package avroregistry

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/heetch/avro"
)

type encodingRegistry struct {
	r       *Registry
	subject string
}

var _ avro.EncodingRegistry = encodingRegistry{}

// AppendSchemaID implements avro.EncodingRegistry.AppendSchemaID
// by appending the id.
// See https://docs.confluent.io/current/schema-registry/serializer-formatter.html#wire-format.
func (r encodingRegistry) AppendSchemaID(buf []byte, id int64) []byte {
	if id < 0 || id >= 1<<32-1 {
		panic("schema id out of range")
	}
	n := len(buf)
	// Magic zero byte, then 4 bytes of schema ID.
	buf = append(buf, 0, 0, 0, 0, 0)
	binary.BigEndian.PutUint32(buf[n+1:], uint32(id))
	return buf
}

// IDForSchema implements avro.EncodingRegistry.IDForSchema
// by fetching the schema ID from the registry server.
//
// See https://docs.confluent.io/current/schema-registry/develop/api.html#post--subjects-(string-%20subject).
func (r encodingRegistry) IDForSchema(ctx context.Context, schema *avro.Type) (int64, error) {
	data, err := json.Marshal(struct {
		Schema string `json:"schema"`
	}{canonical(schema)})
	if err != nil {
		return 0, err
	}
	req := r.r.newRequest(ctx, "POST", "/subjects/"+r.subject, bytes.NewReader(data))

	var resp Schema
	if err := r.r.doRequest(req, &resp); err != nil {
		return 0, err
	}
	// TODO could check that the subject is the same as r.params.Subject.
	return resp.ID, nil
}

type decodingRegistry struct {
	r *Registry
}

var _ avro.DecodingRegistry = decodingRegistry{}

// DecodeSchemaID implements avro.DecodingRegistry.DecodeSchemaID
// by stripping off the schema-identifier header.
//
// See https://docs.confluent.io/current/schema-registry/serializer-formatter.html#wire-format.
func (r decodingRegistry) DecodeSchemaID(msg []byte) (int64, []byte) {
	if len(msg) < 5 || msg[0] != 0 {
		return 0, nil
	}
	return int64(binary.BigEndian.Uint32(msg[1:5])), msg[5:]
}

// SchemaForID implements avro.DecodingRegistry.SchemaForID
// by fetching the schema from the registry server.
//
// See https://docs.confluent.io/current/schema-registry/develop/api.html#get--schemas-ids-int-%20id
func (r decodingRegistry) SchemaForID(ctx context.Context, id int64) (*avro.Type, error) {
	req := r.r.newRequest(ctx, "GET", fmt.Sprintf("/schemas/ids/%d", id), nil)
	var resp struct {
		Schema string `json:"schema"`
	}
	if err := r.r.doRequest(req, &resp); err != nil {
		return nil, err
	}
	t, err := avro.ParseType(resp.Schema)
	if err != nil {
		return nil, fmt.Errorf("invalid schema (%q) in response: %v", resp.Schema, err)
	}
	return t, nil
}
