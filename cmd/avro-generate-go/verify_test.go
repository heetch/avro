package main

import (
	"bytes"
	"os"
	"os/exec"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/heetch/avro/cmd/avro-generate-go/internal/avrotestdata"
)

// Note: external command is called with three args:
//	in-schema, in-data, out-schema, all in JSON format
// It's expected to produce JSON output with the round-tripped data.

func TestExternalVerify(t *testing.T) {
	c := qt.New(t)
	verifyCmd := os.Getenv("AVRO_VERIFY")
	if verifyCmd == "" {
		c.Skip("$AVRO_VERIFY not set")
	}
	tests, err := avrotestdata.Load("./testdata")
	c.Assert(err, qt.Equals, nil)
	// Run tests in deterministic order by sorting keys.
	for _, test := range tests {
		test := test
		c.Run(test.TestName, func(c *qt.C) {
			if test.GoTypeBody != "" {
				c.Skip("test uses GoTypeBody")
			}
			if test.GenerateError != "" {
				c.Skip("test uses GenerateError")
			}
			inSchema := jsonMarshal(test.InSchema)
			outSchema := jsonMarshal(test.OutSchema)
			for _, subtest := range test.Subtests {
				subtest := subtest
				c.Run(subtest.TestName, func(c *qt.C) {
					if subtest.ExpectError != nil {
						c.Skip("subtest uses ExpectError")
					}
					inData := jsonMarshal(subtest.InData)
					c.Logf("run %s '%s' '%s' '%s'", verifyCmd, inSchema, outSchema, inData)
					cmd := exec.Command(verifyCmd, inSchema, outSchema, inData)
					var stdout, stderr bytes.Buffer
					cmd.Stdout = &stdout
					cmd.Stderr = &stderr
					err := cmd.Run()
					c.Assert(err, qt.Equals, nil, qt.Commentf("stderr:\n%s", &stderr))
					c.Assert(stdout.Bytes(), qt.JSONEquals, subtest.OutData)
				})
			}
		})
	}
}
