// Package avroregistry provides avro.*Registry implementations
// that consult an Avro registry through its REST API.
package avroregistry

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/heetch/avro"
	"gopkg.in/httprequest.v1"
	retry "gopkg.in/retry.v1"
)

// Registry represents an Avro registry server. It implements avro.EncodingRegistry
// and avro.DecodingRegistry.
type Registry struct {
	params Params
}

type Params struct {
	// ServerURL holds the URL of the Avro registry server, for example "http://localhost:8084".
	ServerURL string

	// RetryStrategy is used when requests are retried after HTTP errors.
	// If this is nil, a default exponential-backoff strategy is used.
	RetryStrategy retry.Strategy

	// Username and Password hold the basic auth credentials to use.
	// If Userame is empty, no authentication will be sent.
	Username string
	Password string
}

// Schema holds the schema metadata and actual schema stored in a Schema registry
type Schema struct {
	// Subject defines the name this schema is registered under
	Subject string `json:"subject"`
	// ID globally unique schema identifier
	ID int64 `json:"id"`
	// Version is the version of the schema
	Version int `json:"version"`
	// Schema is the actual schema in Avro format
	Schema string `json:"schema"`
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

// Encoder returns an avro.EncodingRegistry implementation that can be
// used to encode messages with schemas associated with the given
// subject, which must be non-empty.
func (r *Registry) Encoder(subject string) avro.EncodingRegistry {
	return encodingRegistry{
		r:       r,
		subject: subject,
	}
}

// Decoder returns an avro.DecodingRegistry implementation
// that can be used to decode messages from the registry.
func (r *Registry) Decoder() avro.DecodingRegistry {
	return decodingRegistry{
		r: r,
	}
}

// Register registers a schema with the registry associated
// with the given subject and returns its id.
//
// See https://docs.confluent.io/current/schema-registry/develop/api.html#post--subjects-(string-%20subject)-versions
func (r *Registry) Register(ctx context.Context, subject string, schema *avro.Type) (_ int64, err error) {
	// Note: because of https://github.com/confluentinc/schema-registry/issues/1348
	// we need to strip metadata from the schema when registering.
	data, err := json.Marshal(struct {
		Schema string `json:"schema"`
	}{canonical(schema)})
	if err != nil {
		return 0, err
	}
	req := r.newRequest(ctx, "POST", fmt.Errorf("/subjects/%s/versions", subject).Error(), bytes.NewReader(data))
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
func (r *Registry) SetCompatibility(ctx context.Context, subject string, mode avro.CompatMode) error {
	data, err := json.Marshal(struct {
		Compatibility string `json:"compatibility"`
	}{mode.String()})
	if err != nil {
		return err
	}
	return r.doRequest(r.newRequest(ctx, "PUT", "/config/"+subject, bytes.NewReader(data)), nil)
}

// DeleteSubject deletes the  given subject from the registry.
//
// See https://docs.confluent.io/current/schema-registry/develop/api.html#delete--subjects-(string-%20subject)
func (r *Registry) DeleteSubject(ctx context.Context, subject string) error {
	return r.doRequest(r.newRequest(ctx, "DELETE", "/subjects/"+subject, nil), nil)
}

// Schema gets a specific version of the schema registered under this subject
//
// See https://docs.confluent.io/platform/current/schema-registry/develop/api.html#get--subjects-(string-%20subject)-versions-(versionId-%20version)
func (r *Registry) Schema(ctx context.Context, subject, version string) (*Schema, error) {
	// validate version
	if err := validateVersion(version); err != nil {
		return nil, err
	}

	req := r.newRequest(ctx, http.MethodGet, fmt.Errorf("/subjects/%s/versions/%s", subject, version).Error(), nil)
	schema := new(Schema)
	if err := r.doRequest(req, schema); err != nil {
		return nil, err
	}

	return schema, nil
}

func validateVersion(version string) error {
	msg := `Invalid version. It should be between 1 and 2^31-1 or "latest"`
	i, err := strconv.Atoi(version)
	if err != nil {
		if version == "latest" {
			return nil
		}
		return fmt.Errorf("%s: %w", msg, err)
	} else if i < 1 || i > 1<<31-1 {
		return fmt.Errorf("%s: %d provided", msg, i)
	}

	return nil
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
	if req.Body != nil {
		req.Header.Set("Content-Type", "application/vnd.schemaregistry.v1+json")
	}
	if r.params.Username != "" {
		req.SetBasicAuth(r.params.Username, r.params.Password)
	}
	ctx := req.Context()
	attempt := retry.StartWithCancel(r.params.RetryStrategy, nil, ctx.Done())
	for attempt.Next() {
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			if !attempt.More() || !isTemporaryError(err) {
				return err
			}
			continue
		}
		err = unmarshalResponse(req, resp, result)
		if err == nil {
			return nil
		}
		if !attempt.More() {
			return err
		}
		if err, ok := err.(*apiError); ok && err.StatusCode/100 != 5 {
			// It's not a 5xx error. We want to retry on 5xx
			// errors, because the Confluent Avro registry
			// can occasionally return them as a matter of
			// course (and there could also be an
			// unavailable service that we're reaching
			// through a proxy).
			return err
		}
	}
	if attempt.Stopped() {
		return ctx.Err()
	}
	panic("unreachable")
}

func isTemporaryError(err error) bool {
	err1, ok := err.(interface {
		Temporary() bool
	})
	return ok && err1.Temporary()
}

func unmarshalResponse(req *http.Request, resp *http.Response, result interface{}) error {
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		if err := httprequest.UnmarshalJSONResponse(resp, result); err != nil {
			return fmt.Errorf("cannot unmarshal JSON response from %v: %v", req.URL, err)
		}
		return nil
	}
	var apiErr apiError
	if err := httprequest.UnmarshalJSONResponse(resp, &apiErr); err != nil {
		return fmt.Errorf("cannot unmarshal JSON error response from %v: %v", req.URL, err)
	}
	apiErr.StatusCode = resp.StatusCode
	return &apiErr
}

// https://docs.confluent.io/current/schema-registry/develop/api.html#errors
type apiError struct {
	ErrorCode  int    `json:"error_code"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
}

func (e *apiError) Error() string {
	if e.StatusCode != e.ErrorCode {
		return fmt.Errorf("avro registry error (code %d; HTTP status %d): %v", e.ErrorCode, e.StatusCode, e.Message).Error()
	}
	return fmt.Errorf("avro registry error (HTTP status %d): %v", e.ErrorCode, e.Message).Error()
}

func canonical(schema *avro.Type) string {
	return schema.CanonicalString(avro.RetainDefaults | avro.RetainLogicalTypes)
}
