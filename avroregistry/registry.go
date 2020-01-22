// Package avroregistry provides avro.*Registry implementations
// that consult an Avro registry through its REST API.
package avroregistry

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/heetch/avro"
	"gopkg.in/retry.v1"
)

// Registry represents an Avro registry server. It implements avro.EncodingRegistry
// and avro.DecodingRegistry.
type Registry struct {
	params Params
}

var (
	_ avro.EncodingRegistry = (*Registry)(nil)
	_ avro.DecodingRegistry = (*Registry)(nil)
)

type Params struct {
	// ServerURL holds the URL of the Avro registry server, for example "http://localhost:8084".
	ServerURL string

	// Subject holds the registry subject to use. This must be non-empty, and
	// is usually the name of the Kafka topic.
	Subject string

	// RetryStrategy is used when requests are retried after HTTP errors.
	// If this is nil, a default exponential-backoff strategy is used.
	RetryStrategy retry.Strategy

	// Username and Password hold the basic auth credentials to use.
	// If Userame is empty, no authentication will be sent.
	Username string
	Password string
}

var defaultRetryStrategy = retry.LimitTime(5*time.Second, retry.Exponential{
	Initial:  time.Millisecond,
	MaxDelay: time.Second,
	Jitter:   true,
})

func New(p Params) (*Registry, error) {
	if p.RetryStrategy == nil {
		p.RetryStrategy = defaultRetryStrategy
	}
	if p.Subject == "" {
		return nil, fmt.Errorf("no subject found for Avro registry")
	}
	if p.ServerURL == "" {
		return nil, fmt.Errorf("no server address found for Avro registry")
	}
	if u, err := url.Parse(p.ServerURL); err != nil || u.Scheme == "" {
		return nil, fmt.Errorf("invalid server address %q", p.ServerURL)
	}
	return &Registry{
		params: p,
	}, nil
}

// AppendSchemaID implements avro.EncodingRegistry.AppendSchemaID
// by appending the id.
// See https://docs.confluent.io/current/schema-registry/serializer-formatter.html#wire-format.
func (r *Registry) AppendSchemaID(buf []byte, id int64) []byte {
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
func (r *Registry) IDForSchema(ctx context.Context, schema string) (int64, error) {
	data, err := json.Marshal(struct {
		Schema string `json:"schema"`
	}{schema})
	if err != nil {
		return 0, err
	}
	req := r.newRequest(ctx, "POST", "/subjects/"+r.params.Subject, bytes.NewReader(data))

	var resp struct {
		Subject string `json:"subject"`
		ID      int64  `json:"id"`
		Version int    `json:"version"`
		Schema  string `json:"schema"`
	}
	if err := r.doRequest(req, &resp); err != nil {
		return 0, err
	}
	// TODO could check that the subject is the same as r.params.Subject.
	return resp.ID, nil
}

// DecodeSchemaID implements avro.DecodingRegistry.DecodeSchemaID
// by stripping off the schema-identifier header.
//
// See https://docs.confluent.io/current/schema-registry/serializer-formatter.html#wire-format.
func (r *Registry) DecodeSchemaID(msg []byte) (int64, []byte) {
	if len(msg) < 5 || msg[0] != 0 {
		return 0, nil
	}
	return int64(binary.BigEndian.Uint32(msg[1:5])), msg[5:]
}

// SchemaForID implements avro.DecodingRegistry.SchemaForID
// by fetching the schema from the registry server.
//
// See https://docs.confluent.io/current/schema-registry/develop/api.html#get--schemas-ids-int-%20id
func (r *Registry) SchemaForID(ctx context.Context, id int64) (string, error) {
	req := r.newRequest(ctx, "GET", fmt.Sprintf("/schemas/ids/%d", id), nil)
	var resp struct {
		Schema string `json:"schema"`
	}
	if err := r.doRequest(req, &resp); err != nil {
		return "", err
	}
	return resp.Schema, nil
}

// Register registers a schema with the registry and returns its id.
//
// See https://docs.confluent.io/current/schema-registry/develop/api.html#post--subjects-(string-%20subject)-versions
func (r *Registry) Register(ctx context.Context, schema string) (int64, error) {
	data, err := json.Marshal(struct {
		Schema string `json:"schema"`
	}{schema})
	if err != nil {
		return 0, err
	}
	req := r.newRequest(ctx, "POST", fmt.Sprintf("/subjects/%s/versions", r.params.Subject), bytes.NewReader(data))
	var resp struct {
		ID int64 `json:"id"`
	}
	if err := r.doRequest(req, &resp); err != nil {
		return 0, err
	}
	return resp.ID, nil
}

// SetCompatibility sets the compatibility mode for the registry's subject to mode.
//
// See https://docs.confluent.io/current/schema-registry/develop/api.html#put--config-(string-%20subject)
func (r *Registry) SetCompatibility(ctx context.Context, mode avro.CompatMode) error {
	data, err := json.Marshal(struct {
		Compatibility string `json:"compatibility"`
	}{mode.String()})
	if err != nil {
		return err
	}
	return r.doRequest(r.newRequest(ctx, "PUT", "/config/"+r.params.Subject, bytes.NewReader(data)), nil)
}

// DeleteSubject deletes the registry's subject from the registry.
//
// See https://docs.confluent.io/current/schema-registry/develop/api.html#delete--subjects-(string-%20subject)
func (r *Registry) DeleteSubject(ctx context.Context) error {
	return r.doRequest(r.newRequest(ctx, "DELETE", "/subjects/"+r.params.Subject, nil), nil)
}

func (r *Registry) newRequest(ctx context.Context, method string, urlStr string, body io.Reader) *http.Request {
	req, err := http.NewRequestWithContext(ctx, method, r.params.ServerURL+urlStr, body)
	if err != nil {
		// Should never happen, as we've checked the URL for validity when
		// creating the registry instance.
		panic(err)
	}
	return req
}

func (r *Registry) doRequest(req *http.Request, result interface{}) error {
	// TODO should we specificy a version number of the API to accept?
	req.Header.Set("Accept", "application/vnd.schemaregistry.v1+json")
	if r.params.Username != "" {
		req.SetBasicAuth(r.params.Username, r.params.Password)
	}
	ctx := req.Context()
	var resp *http.Response
	attempt := retry.StartWithCancel(r.params.RetryStrategy, nil, ctx.Done())
	for attempt.Next() {
		resp1, err := http.DefaultClient.Do(req)
		if err == nil {
			resp = resp1
			break
		}
		// TODO only retry if error is temporary?
		if !attempt.More() {
			return err
		}
	}
	if attempt.Stopped() {
		return ctx.Err()
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("cannot read response body: %v", err)
	}
	if resp.StatusCode == http.StatusOK {
		if result != nil {
			if err := json.Unmarshal(data, result); err != nil {
				return fmt.Errorf("cannot unmarshal JSON response from %v: %v", req.URL, err)
			}
		}
		return nil
	}
	var apiErr apiError
	if err := json.Unmarshal(data, &apiErr); err != nil {
		return fmt.Errorf("cannot unmarshal JSON error response from %v: %v", req.URL, err)
	}
	return &apiErr
}

// https://docs.confluent.io/current/schema-registry/develop/api.html#errors
type apiError struct {
	ErrorCode int    `json:"error_code"`
	Message   string `json:"message"`
}

func (e *apiError) Error() string {
	return fmt.Sprintf("Avro registry error (code %d): %v", e.ErrorCode, e.Message)
}
