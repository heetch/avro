package avrotest

import (
	"context"
	"fmt"
	"os"

	"github.com/heetch/avro"
	"github.com/heetch/avro/avroregistry"
)

// Register an avro type in a registry for a particular topic and then delete the subject at the end of the test.
//
// The following environment variables can be used to
// configure the connection parameters:
//
//	- $KAFKA_REGISTRY_ADDR
//		The Kafka registry address in host:port
//		form. If this is empty, localhost:8084 will be used.
//
// This requires go1.14 or higher
func Register(ctx context.Context, t T, x interface{}, topic string) error {
	registryAddr := os.Getenv("KAFKA_REGISTRY_ADDR")
	if registryAddr == "" {
		registryAddr = "localhost:8084"
	}

	registry, err := avroregistry.New(avroregistry.Params{
		ServerURL: "http://" + registryAddr,
	})
	if err != nil {
		return fmt.Errorf("cannot connect to registry: %w", err)
	}

	avroType, err := avro.TypeOf(x)
	if err != nil {
		return fmt.Errorf("cannot generate Avro schema for %T: %w", x, err)
	}

	_, err = registry.Register(ctx, topic, avroType)
	if err != nil {
		return fmt.Errorf("cannot register %T in %v: %w", x, topic, err)
	}

	t.Cleanup(func() {
		err := registry.DeleteSubject(ctx, topic)
		if err != nil {
			t.Errorf("cannot delete subject: %w", err)
		}
	})

	return nil
}

// T represents a test (the usual implementation being *testing.T).
type T interface {
	Cleanup(f func())
	Errorf(f string, a ...interface{})
}
