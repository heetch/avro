package avroregistry_test

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
	"gopkg.in/retry.v1"

	"github.com/heetch/avro"
	"github.com/heetch/avro/avroregistry"
)

func TestRegister(t *testing.T) {
	c := qt.New(t)

	r, subject := newTestRegistry(c)

	type R struct {
		X int
	}
	ctx := context.Background()
	id, err := r.Register(ctx, subject, schemaOf(nil, R{}))
	c.Assert(err, qt.Equals, nil)

	id1, err := r.Encoder(subject).IDForSchema(ctx, schemaOf(nil, R{}))
	c.Assert(err, qt.Equals, nil)
	c.Assert(id1, qt.Equals, id)
}

func TestRegisterWithEmptyStruct(t *testing.T) {
	c := qt.New(t)

	r, subject := newTestRegistry(c)
	type Empty struct{}
	type R struct {
		X Empty
	}
	ctx := context.Background()
	_, err := r.Register(ctx, subject, schemaOf(nil, R{}))
	c.Assert(err, qt.Equals, nil)
}

func TestSchemaCompatibility(t *testing.T) {
	c := qt.New(t)

	r, subject := newTestRegistry(c)
	ctx := context.Background()
	err := r.SetCompatibility(ctx, subject, avro.BackwardTransitive)
	c.Assert(err, qt.Equals, nil)

	type R struct {
		X int
	}
	_, err = r.Register(ctx, subject, schemaOf(nil, R{}))
	c.Assert(err, qt.Equals, nil)

	// Try to register an incompatible type.
	type R1 struct {
		X string
	}
	names := new(avro.Names).RenameType(R1{}, "R")
	_, err = r.Register(ctx, subject, schemaOf(names, R1{}))
	c.Assert(err, qt.ErrorMatches, `Avro registry error \(HTTP status 409\): Schema being registered is incompatible with an earlier schema for subject "`+subject+`"`)

	// Check that we can't rename the schema.
	_, err = r.Register(ctx, subject, schemaOf(nil, R1{}))
	c.Assert(err, qt.ErrorMatches, `Avro registry error \(HTTP status 409\): Schema being registered is incompatible with an earlier schema for subject "`+subject+`"`)

	// Check that we can change the field to a compatible union.
	type R2 struct {
		X *int
	}
	names = new(avro.Names).RenameType(R2{}, "R")
	_, err = r.Register(ctx, subject, schemaOf(names, R2{}))
	c.Assert(err, qt.Equals, nil)

	// Check that we can't change it back again.
	type R3 struct {
		X int
		Y string
	}
	names = new(avro.Names).RenameType(R3{}, "R")
	_, err = r.Register(ctx, subject, schemaOf(names, R3{}))
	c.Assert(err, qt.ErrorMatches, `Avro registry error \(HTTP status 409\): Schema being registered is incompatible with an earlier schema for subject "`+subject+`"`)
}

func TestSchemasRetainLogicalTypes(t *testing.T) {
	c := qt.New(t)

	r, subject := newTestRegistry(c)
	ctx := context.Background()
	type R struct {
		T time.Time
	}
	id, err := r.Register(ctx, subject, schemaOf(nil, R{}))
	c.Assert(err, qt.Equals, nil)
	schema, err := r.Decoder().SchemaForID(ctx, id)
	c.Assert(err, qt.Equals, nil)
	c.Assert(schema.String(), qt.Equals, `{"type":"record","name":"R","fields":[{"name":"T","type":{"type":"long","logicalType":"timestamp-micros"},"default":0}]}`)
}

func TestSingleCodec(t *testing.T) {
	c := qt.New(t)

	r, subject := newTestRegistry(c)
	ctx := context.Background()
	err := r.SetCompatibility(ctx, subject, avro.BackwardTransitive)
	c.Assert(err, qt.Equals, nil)
	type R struct {
		X int
	}
	id1, err := r.Register(ctx, subject, schemaOf(nil, R{}))
	c.Assert(err, qt.Equals, nil)

	type R1 struct {
		X int
		Y int
	}
	names := new(avro.Names).RenameType(R1{}, "R")
	id2, err := r.Register(ctx, subject, schemaOf(names, R1{}))
	c.Assert(err, qt.Equals, nil)
	c.Assert(id2, qt.Not(qt.Equals), id1)

	enc := avro.NewSingleEncoder(r.Encoder(subject), names)
	data1, err := enc.Marshal(ctx, R{10})
	c.Assert(err, qt.Equals, nil)
	c.Assert(data1[0], qt.Equals, byte(0))
	c.Assert(int64(binary.BigEndian.Uint32(data1[1:5])), qt.Equals, id1)
	c.Assert(data1[5:], qt.DeepEquals, []byte{20})

	data2, err := enc.Marshal(ctx, R1{11, 30})
	c.Assert(err, qt.Equals, nil)
	c.Assert(data2[0], qt.Equals, byte(0))
	c.Assert(int64(binary.BigEndian.Uint32(data2[1:5])), qt.Equals, id2)
	c.Assert(data2[5:], qt.DeepEquals, []byte{22, 60})

	// Check that it round-trips back through the decoder OK.
	dec := avro.NewSingleDecoder(r.Decoder(), names)
	var x1 R
	_, err = dec.Unmarshal(ctx, data1, &x1)
	c.Assert(err, qt.Equals, nil)
	c.Assert(x1, qt.Equals, R{10})

	var x2 R1
	_, err = dec.Unmarshal(ctx, data2, &x2)
	c.Assert(err, qt.Equals, nil)
	c.Assert(x2, qt.Equals, R1{11, 30})
}

func TestSchema(t *testing.T) {
	c := qt.New(t)

	ctx := context.Background()
	r, subject := newTestRegistry(c)

	c.Run("OK", func(c *qt.C) {
		type R struct {
			X int
		}
		id1, err := r.Register(ctx, subject, schemaOf(nil, R{}))
		c.Assert(err, qt.IsNil)

		schema1, err := r.Schema(ctx, subject, "latest")
		c.Assert(err, qt.IsNil)
		c.Assert(schema1, qt.DeepEquals, &avroregistry.Schema{
			Subject: subject,
			ID:      id1,
			Version: 1,
			Schema:  `{"type":"record","name":"R","fields":[{"name":"X","type":"long","default":0}]}`,
		})

		type R1 struct {
			X int
			Y string
		}
		names := new(avro.Names).RenameType(R1{}, "R")
		id2, err := r.Register(ctx, subject, schemaOf(names, R1{}))
		c.Assert(err, qt.IsNil)
		c.Assert(id2, qt.Not(qt.Equals), id1)

		schema2, err := r.Schema(ctx, subject, "2")
		c.Assert(err, qt.IsNil)
		c.Assert(schema2, qt.DeepEquals, &avroregistry.Schema{
			Subject: subject,
			ID:      id2,
			Version: 2,
			Schema:  `{"type":"record","name":"R","fields":[{"name":"X","type":"long","default":0},{"name":"Y","type":"string","default":""}]}`,
		})

		testSchema, err := r.Schema(ctx, subject, "latest")
		c.Assert(err, qt.IsNil)
		c.Assert(testSchema, qt.DeepEquals, schema2)
		testSchema, err = r.Schema(ctx, subject, "1")
		c.Assert(err, qt.IsNil)
		c.Assert(testSchema, qt.DeepEquals, schema1)
		c.Assert(schema1, qt.Not(qt.DeepEquals), schema2)
	})

	c.Run("Validate version", func(c *qt.C) {
		tests := []struct {
			testName string
			in       string
			err      string
		}{
			{
				testName: "zero",
				in:       "0",
				err:      `Invalid version. It should be between 1 and 2\^31-1 or "latest": 0 provided`,
			}, {
				testName: "> 2^31 < 2^64",
				in:       strconv.FormatInt(1<<60, 10),
				err:      `Invalid version. It should be between 1 and 2\^31-1 or "latest": .* provided`,
			}, {
				testName: "empty",
				in:       "",
				err:      `Invalid version. It should be between 1 and 2\^31-1 or "latest": .* invalid syntax`,
			}, {
				testName: "> 2^64",
				in:       strconv.FormatUint(1<<63, 10),
				err:      `Invalid version. It should be between 1 and 2\^31-1 or "latest": .* value out of range`,
			},
		}

		for _, test := range tests {
			c.Run(test.testName, func(c *qt.C) {
				s, err := r.Schema(ctx, subject, test.in)
				c.Assert(err, qt.ErrorMatches, test.err)
				c.Assert(s, qt.IsNil)
			})
		}

	})
}

func TestRetryOnError(t *testing.T) {
	c := qt.New(t)

	c.Patch(&http.DefaultClient.Transport, errorTransport(tmpError(true)))
	registry, err := avroregistry.New(avroregistry.Params{
		ServerURL: "http://0.1.2.3",
		RetryStrategy: retry.LimitCount(4, retry.Regular{
			Total: time.Second,
			Delay: 10 * time.Millisecond,
		}),
	})
	c.Assert(err, qt.Equals, nil)
	t0 := time.Now()
	err = registry.SetCompatibility(context.Background(), "x", avro.BackwardTransitive)
	c.Assert(err, qt.ErrorMatches, `Put "?http://0.1.2.3/config/x"?: temporary test error true`)
	if d := time.Since(t0); d < 30*time.Millisecond {
		c.Errorf("retry duration too small, want >=30ms got %v", d)
	}
}

func TestCanceledRetry(t *testing.T) {
	c := qt.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()
	c.Patch(&http.DefaultClient.Transport, errorTransport(tmpError(true)))
	registry, err := avroregistry.New(avroregistry.Params{
		ServerURL: "http://0.1.2.3",
	})
	c.Assert(err, qt.Equals, nil)
	t0 := time.Now()
	err = registry.SetCompatibility(ctx, "x", avro.BackwardTransitive)
	c.Assert(err, qt.ErrorMatches, `context canceled`)
	if d := time.Since(t0); d > 500*time.Millisecond {
		c.Errorf("retry duration too large, want ~30ms got %v", d)
	}
}

func TestRetryOn500(t *testing.T) {
	c := qt.New(t)

	failCount := 3
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if failCount == 0 {
			return
		}
		failCount--
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		_, _ = w.Write([]byte(`{"error_code":50001,"message":"Failed to update compatibility level"}`))
	}))
	defer srv.Close()
	registry, err := avroregistry.New(avroregistry.Params{
		ServerURL: srv.URL,
		RetryStrategy: retry.LimitCount(4, retry.Regular{
			Total: time.Second,
			Delay: 10 * time.Millisecond,
		}),
	})
	c.Assert(err, qt.Equals, nil)
	t0 := time.Now()
	err = registry.SetCompatibility(context.Background(), "x", avro.BackwardTransitive)
	c.Assert(err, qt.Equals, nil)
	if d := time.Since(t0); d < 30*time.Millisecond {
		c.Errorf("retry duration too small, want >=30ms got %v", d)
	}

	// If it fails more times than the retry limit, we should get
	// an error.
	failCount = 5
	err = registry.SetCompatibility(context.Background(), "x", avro.BackwardTransitive)
	c.Assert(err, qt.ErrorMatches, `Avro registry error \(code 50001; HTTP status 500\): Failed to update compatibility level`)
}

func TestNoRetryOnNon5XXStatus(t *testing.T) {
	c := qt.New(t)

	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(409)
		_, _ = w.Write([]byte(`{"error_code":409,"message":"incompatible wotsit"}`))
	}))
	defer srv.Close()
	registry, err := avroregistry.New(avroregistry.Params{
		ServerURL: srv.URL,
		RetryStrategy: retry.LimitCount(4, retry.Regular{
			Total: time.Second,
			Delay: 10 * time.Millisecond,
		}),
	})
	c.Assert(err, qt.IsNil)
	err = registry.SetCompatibility(context.Background(), "x", avro.BackwardTransitive)
	c.Assert(err, qt.ErrorMatches, `Avro registry error \(HTTP status 409\): incompatible wotsit`)
	c.Assert(calls, qt.Equals, 1)
}

func TestUnavailableError(t *testing.T) {
	c := qt.New(t)

	// When the service in unavailable, the response is probably not
	// formatted as JSON.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")
		w.WriteHeader(500)
		_, _ = w.Write([]byte(`
<!DOCTYPE HTML PUBLIC "-//IETF//DTD HTML 2.0//EN">
<html><head>
<title>502 Proxy Error</title>
</head><body>
<h1>Proxy Error</h1>
<p>The whole world is bogus
</body>
`))
	}))
	defer srv.Close()
	registry, err := avroregistry.New(avroregistry.Params{
		ServerURL:     srv.URL,
		RetryStrategy: retry.Regular{},
	})
	c.Assert(err, qt.Equals, nil)
	err = registry.SetCompatibility(context.Background(), "x", avro.BackwardTransitive)
	c.Assert(err, qt.ErrorMatches, `cannot unmarshal JSON error response from .*/config/x: unexpected content type text/html; want application/json; content: 502 Proxy Error; Proxy Error; The whole world is bogus`)
}

var schemaEquivalenceTests = []struct {
	testName string
	register string
	fetch    string
}{{
	testName: "ignore_whitespace",
	register: `     "string"    `,
	fetch:    `"string" `,
}, {
	testName: "namespace_normalization",
	register: `{
		"type": "record",
		"name": "com.example.Foo",
		"fields": [{
			 "name": "a",
			 "type": {
			 	"type": "enum",
			 	"name": "Bar",
			 	"symbols": ["a", "b"]
			 }
		}]
	}`,
	fetch: `{
		"type": "record",
		"name": "com.example.Foo",
		"fields": [{
			 "name": "a",
			 "type": {
			 	"type": "enum",
			 	"name": "com.example.Bar",
			 	"symbols": ["a", "b"]
			 }
		}]
	}`,
}, {
	testName: "metadata_normalization#1",
	register: `{
		"type": "record",
		"name": "com.example.Foo",
		"fields": [{
			 "name": "a",
			 "type": "string"
		}],
		"extraMetadata": "hello"
	}`,
	fetch: `{
		"type": "record",
		"name": "com.example.Foo",
		"fields": [{
			 "name": "a",
			 "type": "string"
		}]
	}`,
}, {
	testName: "metadata_normalization#2",
	register: `{
		"type": "record",
		"name": "com.example.Foo",
		"fields": [{
			 "name": "a",
			 "type": "string"
		}]
	}`,
	fetch: `{
		"type": "record",
		"name": "com.example.Foo",
		"fields": [{
			 "name": "a",
			 "type": "string"
		}],
		"extraMetadata": "hello"
	}`,
}, {
	testName: "metadata_field_order",
	register: `{
		"type": "record",
		"name": "com.example.Foo",
		"fields": [{
			 "name": "a",
			 "type": "string"
		}],
		"extraMetadata": {
			"a": 1,
			"b": 1
		}
	}`,
	fetch: `{
		"type": "record",
		"name": "com.example.Foo",
		"fields": [{
			 "name": "a",
			 "type": "string"
		}],
		"extraMetadata": {
			"b": 1,
			"a": 1
		}
	}`,
}}

func TestSchemaEquivalence(t *testing.T) {
	c := qt.New(t)
	for _, test := range schemaEquivalenceTests {
		test := test
		c.Run(test.testName, func(c *qt.C) {
			ctx := context.Background()
			r, subject := newTestRegistry(c)
			// Sanity check it's not there already.
			_, err := r.Encoder(subject).IDForSchema(ctx, parseType(test.fetch))
			c.Assert(err, qt.Not(qt.IsNil))
			id, err := r.Register(ctx, subject, parseType(test.register))
			c.Assert(err, qt.Equals, nil)
			gotID, err := r.Encoder(subject).IDForSchema(ctx, parseType(test.fetch))
			c.Assert(err, qt.Equals, nil)
			c.Assert(gotID, qt.Equals, id)
		})
	}
}

func schemaOf(names *avro.Names, x interface{}) *avro.Type {
	if names == nil {
		names = new(avro.Names)
	}
	t, err := names.TypeOf(x)
	if err != nil {
		panic(err)
	}
	return t
}

func parseType(s string) *avro.Type {
	t, err := avro.ParseType(s)
	if err != nil {
		panic(err)
	}
	return t
}

// newTestRegistry returns a registry instance connected server
// pointed by KAFKA_REGISTRY_ADDR env var with a random subject to use.
func newTestRegistry(c *qt.C) (*avroregistry.Registry, string) {
	ctx := context.Background()
	serverAddr := os.Getenv("KAFKA_REGISTRY_ADDR")
	if serverAddr == "" {
		c.Skip("no KAFKA_REGISTRY_ADDR configured")
	}
	serverURL := "http://" + serverAddr
	subject := randomString()
	registry, err := avroregistry.New(avroregistry.Params{
		ServerURL:     serverURL,
		RetryStrategy: noRetry,
	})
	c.Assert(err, qt.Equals, nil)
	c.Cleanup(func() {
		err := registry.DeleteSubject(ctx, subject)
		c.Check(err, qt.Equals, nil)
	})
	return registry, subject
}

var noRetry = retry.Regular{}

func randomString() string {
	buf := make([]byte, 16)
	_, err := rand.Read(buf)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("test-%x", buf)
}

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func errorTransport(err error) http.RoundTripper {
	return transportFunc(func(*http.Request) (*http.Response, error) {
		return nil, err
	})
}

type tmpError bool

func (e tmpError) Error() string {
	return fmt.Sprintf("temporary test error %t", e)
}

func (e tmpError) Temporary() bool {
	return bool(e)
}
