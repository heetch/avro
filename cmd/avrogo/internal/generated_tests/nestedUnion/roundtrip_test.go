// Code generated by generatetestcode.go; DO NOT EDIT.

package nestedUnion

import (
	"testing"

	"github.com/heetch/avro/cmd/avrogo/internal/testutil"
)

var tests = testutil.RoundTripTest{
	InSchema: `{
                "name": "R",
                "type": "record",
                "fields": [
                    {
                        "name": "F",
                        "type": {
                            "type": "array",
                            "items": [
                                "int",
                                {
                                    "type": "array",
                                    "items": [
                                        "null",
                                        "string"
                                    ]
                                }
                            ]
                        }
                    }
                ]
            }`,
	GoType: new(R),
	Subtests: []testutil.RoundTripSubtest{{
		TestName: "main",
		InDataJSON: `{
                        "F": [
                            {
                                "int": 1
                            },
                            {
                                "array": [
                                    null,
                                    {
                                        "string": "hello"
                                    }
                                ]
                            }
                        ]
                    }`,
		OutDataJSON: `{
                        "F": [
                            {
                                "int": 1
                            },
                            {
                                "array": [
                                    null,
                                    {
                                        "string": "hello"
                                    }
                                ]
                            }
                        ]
                    }`,
	}},
}

func TestGeneratedCode(t *testing.T) {
	tests.Test(t)
}
