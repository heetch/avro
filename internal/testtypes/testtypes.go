// Package testtypes defines types for testing the avro package
// that aren't easily defined in the test package there.
//
// Note: stringer doesn't work on test files, and we want to make sure
// that enum detection works ok on stringer-generated String methods,
// hence the reason for this package.
package testtypes

//go:generate stringer -type Enum -trimprefix Enum

type Enum int

const (
	EnumOne Enum = iota
	EnumTwo
	EnumThree
)
