package avrotestdata

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

type Test struct {
	TestName      string             `json:"testName"`
	InSchema      json.RawMessage    `json:"inSchema"`
	OutSchema     json.RawMessage    `json:"outSchema"`
	ExtraSchemas  []json.RawMessage  `json:"extraSchemas"`
	GoType        string             `json:"goType"`
	GoTypeBody    string             `json:"goTypeBody"`
	GenerateError string             `json:"generateError"`
	Subtests      map[string]Subtest `json:"subtests"`
	OtherTests    string             `json:"otherTests"`
}

type Subtest struct {
	TestName    string            `json:"testName"`
	InData      json.RawMessage   `json:"inData"`
	OutData     json.RawMessage   `json:"outData"`
	ExpectError map[string]string `json:"expectError"`
}

func Load(dir string) (map[string]Test, error) {
	var buf bytes.Buffer
	cmd := exec.Command("cue", "export")
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	cmd.Stdout = &buf
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("cue export failed: %v", err)
	}
	var exported struct {
		Tests map[string]Test `json:"tests"`
	}
	err = json.Unmarshal(buf.Bytes(), &exported)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal test data: %v", err)
	}
	return exported.Tests, nil
}
