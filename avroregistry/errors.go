package avroregistry

import (
	"fmt"
)

// UnavailableError reports an error when the schema registry is unavailable.
type UnavailableError struct {
	Cause error
}

// Error implements the error interface.
func (m *UnavailableError) Error() string {
	return fmt.Sprintf("schema registry unavailability caused by: %v", m.Cause)
}
