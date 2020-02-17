package avro_test

import (
	"context"
	"sync"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/heetch/avro"
)

func TestSingleEncoder(t *testing.T) {
	c := qt.New(t)
	avroType := mustTypeOf(TestRecord{})
	registry := memRegistry{
		1: avroType.String(),
	}
	enc := avro.NewSingleEncoder(registry, nil)
	data, err := enc.Marshal(context.Background(), TestRecord{A: 20, B: 34})
	c.Assert(err, qt.Equals, nil)
	c.Assert(data, qt.DeepEquals, []byte{1, 40, 68})

	// Check that we can decode it again.
	var x TestRecord
	dec := avro.NewSingleDecoder(registry, nil)
	_, err = dec.Unmarshal(context.Background(), data, &x)
	c.Assert(err, qt.Equals, nil)
	c.Assert(x, qt.DeepEquals, TestRecord{A: 20, B: 34})
}

func TestSingleEncoderCheckMarshalTypeBadType(t *testing.T) {
	c := qt.New(t)
	enc := avro.NewSingleEncoder(memRegistry{}, nil)
	err := enc.CheckMarshalType(context.Background(), struct{ C chan int }{})
	c.Assert(err, qt.ErrorMatches, `cannot use unnamed type struct .*`)
}

func TestSingleEncoderCheckMarshalTypeNotFound(t *testing.T) {
	c := qt.New(t)
	enc := avro.NewSingleEncoder(memRegistry{}, nil)
	err := enc.CheckMarshalType(context.Background(), TestRecord{})
	c.Assert(err, qt.ErrorMatches, `schema not found`)
}

func TestSingleEncoderCachesTypes(t *testing.T) {
	c := qt.New(t)
	registry := &statsRegistry{
		memRegistry: memRegistry{
			1: mustTypeOf(TestRecord{}).String(),
		},
	}
	enc := avro.NewSingleEncoder(registry, nil)
	data, err := enc.Marshal(context.Background(), TestRecord{A: 20, B: 34})
	c.Assert(err, qt.Equals, nil)
	c.Assert(data, qt.DeepEquals, []byte{1, 40, 68})

	// Check that when we marshal it again that we don't get another
	// call to the registry.
	data, err = enc.Marshal(context.Background(), TestRecord{A: 22, B: 35})
	c.Assert(err, qt.Equals, nil)
	c.Assert(data, qt.DeepEquals, []byte{1, 44, 70})
	c.Assert(registry.idForSchemaCount, qt.Equals, 1)
}

func TestSingleEncoderRace(t *testing.T) {
	// Note: this test is designed to be run with the
	// race detector enabled.

	c := qt.New(t)

	type T1 struct {
		A int
	}
	type T2 struct {
		B int
	}
	registry := memRegistry{
		1: mustTypeOf(T1{}).String(),
		2: mustTypeOf(T2{}).String(),
	}
	enc := avro.NewSingleEncoder(registry, nil)
	var wg sync.WaitGroup
	marshal := func(x interface{}) {
		defer wg.Done()
		_, err := enc.Marshal(context.Background(), x)
		c.Check(err, qt.Equals, nil)
	}
	wg.Add(3)
	go marshal(T1{10})
	go marshal(T1{11})
	go marshal(T2{12})
	wg.Wait()
}

// statsRegistry wraps a memRegistry instance and counts calls to some calls.
type statsRegistry struct {
	idForSchemaCount int
	schemaForIDCount int
	memRegistry
}

func (r *statsRegistry) IDForSchema(ctx context.Context, schema string) (int64, error) {
	r.idForSchemaCount++
	return r.memRegistry.IDForSchema(ctx, schema)
}

func (r *statsRegistry) SchemaForID(ctx context.Context, id int64) (string, error) {
	r.schemaForIDCount++
	return r.memRegistry.SchemaForID(ctx, id)
}

func mustTypeOf(x interface{}) *avro.Type {
	t, err := avro.TypeOf(x)
	if err != nil {
		panic(err)
	}
	return t
}
