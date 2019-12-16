// Code generated by generatetestcode.go; DO NOT EDIT.

package fixedDefault

import (
	"testing"

	"github.com/rogpeppe/avro/avro-generate-go/internal/testutil"
)

var test = testutil.RoundTripTest{
	InDataJSON: `{}`,
	OutDataJSON: `{
                "fixedField": "hello"
            }`,
	InSchema: `{
                "name": "R",
                "type": "record",
                "fields": [
                    {
                        "name": "_",
                        "type": "int",
                        "default": 0
                    }
                ]
            }`,
	OutSchema: `{
                "name": "R",
                "type": "record",
                "fields": [
                    {
                        "name": "fixedField",
                        "type": {
                            "name": "five",
                            "type": "fixed",
                            "size": 5
                        },
                        "default": "hello"
                    }
                ]
            }`,
	GoType: new(R),
}

func TestGeneratedCode(t *testing.T) {
	test.Test(t)
}
