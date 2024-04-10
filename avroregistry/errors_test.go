package avroregistry_test

import (
	"errors"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/heetch/avro/avroregistry"
)

func TestUnavailableError_Unwrap(t *testing.T) {
	c := qt.New(t)
	var ErrExpect = errors.New("error")

	err := &avroregistry.UnavailableError{
		Cause: ErrExpect,
	}

	c.Assert(errors.Is(err, ErrExpect), qt.IsTrue)

	var newErr *avroregistry.UnavailableError
	c.Assert(errors.As(err, &newErr), qt.IsTrue)
}

func TestUnavailableError_Error(t *testing.T) {
	c := qt.New(t)

	err := &avroregistry.UnavailableError{
		Cause: errors.New("ECONNREFUSED"),
	}

	c.Assert(err.Error(), qt.Equals, "schema registry unavailability caused by: ECONNREFUSED")
}
