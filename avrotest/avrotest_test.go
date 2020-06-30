// +build go1.14

package avrotest

import (
	"context"
	"crypto/rand"
	"fmt"
	"testing"

	qt "github.com/frankban/quicktest"
)

type x struct {
	Int int
	Str string
}

func TestRegister(t *testing.T) {
	c := qt.New(t)

	c.Run("OK", func(c *qt.C) {
		err := Register(context.Background(), c, x{}, randomName("test-"))
		c.Assert(err, qt.IsNil)
	})

	c.Run("NOK - Wrong interface", func(c *qt.C) {
		err := Register(context.Background(), c, struct{}{}, randomName("test-"))
		c.Assert(err, qt.Not(qt.IsNil))
		c.Assert(err, qt.ErrorMatches, "cannot generate Avro schema for.*")
	})

	c.Run("NOK - Wrong addr", func(c *qt.C) {
		c.Setenv("KAFKA_REGISTRY_ADDR", "-host:1234")
		err := Register(context.Background(), c, x{}, randomName("test-"))
		c.Assert(err, qt.Not(qt.IsNil))
		c.Assert(err, qt.ErrorMatches, "cannot register.*")
	})
}

func randomName(prefix string) string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	return fmt.Sprintf("%s%x", prefix, buf)
}
