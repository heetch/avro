package avro

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/actgardner/gogen-avro/compiler"
	"github.com/actgardner/gogen-avro/schema"
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

type Codec struct {
	getter SchemaGetter

	// mu protects the fields below.
	// We might be better off with a couple of sync.Maps here, but this is a bit easier on the brain.
	mu sync.RWMutex

	// writerSchemas holds a cache of the schemas previously encountered when
	// decoding messages.
	writerSchemas map[int64]schema.AvroType

	// programs holds the programs previously created when decoding.
	programs map[codecSchemaPair]*program
}

// errorSchema is a hack - it pretends to be an AvroType
// so that it can be put into the writerSchemas map.
// In fact it just holds an error so that we can cache SchemaGetter
// failures.
type errorSchema struct {
	schema.AvroType
	err error
}

func (c *Codec) Unmarshal(ctx context.Context, data []byte, x interface{}) error {
	v := reflect.ValueOf(x)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("cannot decode into non-pointer value %T", x)
	}
	v = v.Elem()
	vt := v.Type()
	wID, body := c.getter.SchemaID(data)
	if wID == 0 && body == nil {
		return fmt.Errorf("cannot get schema ID from message")
	}
	prog, err := c.getProgram(ctx, vt, wID)
	if err != nil {
		return fmt.Errorf("cannot unmarshal: %v", err)
	}
	return unmarshal(nil, data, prog, v)
}

func (c *Codec) getProgram(ctx context.Context, vt reflect.Type, wID int64) (*program, error) {
	c.mu.RLock()
	if prog := c.programs[codecSchemaPair{vt, wID}]; prog != nil {
		c.mu.RUnlock()
		return prog, nil
	}
	wSchema := c.writerSchemas[wID]
	c.mu.RUnlock()

	if es, ok := wSchema.(errorSchema); ok {
		return nil, es.err
	}
	var err error
	if wSchema == nil {
		// We haven't seen the writer schema before, so try to fetch it.
		var s string
		s, err = c.getter.SchemaForID(ctx, wID)
		if err == nil {
			wSchema, err = parseSchema(s)
		}
		// TODO look at the SchemaForID error
		// and return an error without caching it if it's temporary?
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if err != nil {
		c.writerSchemas[wID] = errorSchema{err: err}
		return nil, err
	}
	if prog := c.programs[codecSchemaPair{vt, wID}]; prog != nil {
		// Someone else got there first.
		return prog, nil
	}
	prog, err := compileProgram(vt, wSchema)
	if err != nil {
		c.writerSchemas[wID] = errorSchema{err: err}
		return nil, err
	}
	return prog, nil
}

func compileProgram(vt reflect.Type, wSchema schema.AvroType) (*program, error) {
	rSchema, err := schemaForGoType(vt)
	if err != nil {
		return nil, err
	}
	prog0, err := compiler.Compile(wSchema, rSchema)
	if err != nil {
		return nil, err
	}
	prog1, err := analyzeProgramTypes(prog0, vt)
	if err != nil {
		return nil, fmt.Errorf("analysis failed: %v", err)
	}
	return prog1, nil
}

func parseSchema(s string) (schema.AvroType, error) {
	ns := schema.NewNamespace(false)
	avroType, err := ns.TypeForSchema([]byte(s))
	if err != nil {
		return nil, err
	}
	if err := avroType.ResolveReferences(ns); err != nil {
		return nil, fmt.Errorf("cannot resolve references in schema: %v", err)
	}
	return avroType, nil
}

// map from reflect.Type fo schema.AvroType
var goTypeToSchema sync.Map

func schemaForGoType(t reflect.Type) (schema.AvroType, error) {
	if at, ok := goTypeToSchema.Load(t); ok {
		if es, ok := at.(errorSchema); ok {
			return nil, es.err
		}
		return at.(schema.AvroType), nil
	}
	x := reflect.Zero(t).Interface()
	var readerSchema string
	if x, ok := x.(AvroRecord); ok {
		readerSchema = x.AvroRecord().Schema
	} else {
		// TODO generate schema from type
		return nil, fmt.Errorf("cannot get schema info for non-Avro-generated type %T", x)
	}
	sc, err := parseSchema(readerSchema)
	if err != nil {
		goTypeToSchema.LoadOrStore(t, errorSchema{err: err})
		return nil, err
	}
	goTypeToSchema.LoadOrStore(t, sc)
	return sc, err
}
