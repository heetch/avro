package avro

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

type codecShemaPair struct {
	reader   reflect.Type
	writerID int64
}

// SchemaGetter is used by a Codec to find information
// about the schemas used to encode a messages.
// One notable implementation is avroregistry.Registry.
type SchemaGetter interface {
	// SchemaID returns the schema ID of the message
	// and the bare message without schema information.
	// A schema ID is specific to the SchemaGetter instance - within
	// a given SchemaGetter instance (only), a given schema ID
	// must always correspond to the same schema.
	//
	// If the message isn't valid, SchemaID should return (0, nil).
	SchemaID(msg []byte) (int64, []byte)

	// AppendWithSchemaID appends the message encoded along with the
	// given schema ID to the given buffer.
	AppendWithSchemaID(buf []byte, msg []byte, id int64) []byte

	// SchemaForID returns the schema for the given ID.
	SchemaForID(ctx context.Context, id int64) (string, error)
}

type codecSchemaPair struct {
	t        reflect.Type
	schemaID int64
}

// Codec encodes and decodes messages in Avro binary format.
// Each message includes a header or wrapper that indicates the schema
// used to encode the message.
//
// A SchemaGetter is used to retrieve the schema for a given message
// or to find the encoding for a given schema.
//
// To encode or decode a stream of messages that all use the same
// schema, use Encoder or Decoder instead.
//
// TODO implement Codec.Marshal.
type Codec struct {
	getter SchemaGetter

	// mu protects the fields below.
	// We might be better off with a couple of sync.Maps here, but this is a bit easier on the brain.
	mu sync.RWMutex

	// writerTypes holds a cache of the schemas previously encountered when
	// decoding messages.
	writerTypes map[int64]*Type

	// programs holds the programs previously created when decoding.
	programs map[codecSchemaPair]*decodeProgram
}

// NewCodec returns a new Codec
// that uses g to determine the schema of each
// message that's marshaled or unmarshaled.
func NewCodec(g SchemaGetter) *Codec {
	return &Codec{
		getter:      g,
		writerTypes: make(map[int64]*Type),
		programs:    make(map[codecSchemaPair]*decodeProgram),
	}
}

// Unmarshal unmarshals the given message into x. The body
// of the message is unmarshaled as with the Unmarshal function.
//
// It needs the context argument because it might end up
// fetching schema data over the network via the Codec's
// associated SchemaGetter.
func (c *Codec) Unmarshal(ctx context.Context, data []byte, x interface{}) (*Type, error) {
	v := reflect.ValueOf(x)
	if v.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("cannot decode into non-pointer value %T", x)
	}
	v = v.Elem()
	vt := v.Type()
	wID, body := c.getter.SchemaID(data)
	if wID == 0 && body == nil {
		return nil, fmt.Errorf("cannot get schema ID from message")
	}
	prog, err := c.getProgram(ctx, vt, wID)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal: %v", err)
	}
	return unmarshal(nil, body, prog, v)
}

func (c *Codec) getProgram(ctx context.Context, vt reflect.Type, wID int64) (*decodeProgram, error) {
	c.mu.RLock()
	if prog := c.programs[codecSchemaPair{vt, wID}]; prog != nil {
		c.mu.RUnlock()
		return prog, nil
	}
	wType := c.writerTypes[wID]
	c.mu.RUnlock()

	var err error
	if wType != nil {
		if es, ok := wType.avroType.(errorSchema); ok {
			return nil, es.err
		}
	} else {
		// We haven't seen the writer schema before, so try to fetch it.
		var s string
		s, err = c.getter.SchemaForID(ctx, wID)
		if err == nil {
			wType, err = ParseType(s)
		}
		// TODO look at the SchemaForID error
		// and return an error without caching it if it's temporary?
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if err != nil {
		c.writerTypes[wID] = &Type{
			avroType: errorSchema{err: err},
		}
		return nil, err
	}
	if prog := c.programs[codecSchemaPair{vt, wID}]; prog != nil {
		// Someone else got there first.
		return prog, nil
	}

	prog, err := compileDecoder(vt, wType)
	if err != nil {
		c.writerTypes[wID] = &Type{
			avroType: errorSchema{err: err},
		}
		return nil, err
	}
	return prog, nil
}
