package avroCI

import (
	"tool/file"
	"encoding/yaml"
	workflow "github.com/heetch/cue-schema/github/workflow/go:workflow"
)

ci: workflow

ci: RunTest :: """
	set -ex
	tgz=$(mktemp)
	ARCH="$(uname -s)_$(uname -m)"
	mkdir -p ~/go/bin
	export PATH=$PATH:~/go/bin
	curl "https://github.com/cuelang/cue/releases/download/v0.0.15/cue_0.0.15_$ARCH.tar.gz" -L -o $tgz
	(cd ~/go/bin && tar xzf $tgz cue)
	go install ./cmd/... &&
	go generate . ./cmd/...  &&
	go test ./...
	"""

command: generateworkflow: {
	task: write: file.Create & {
		filename: ".github/workflows/test.yaml"
		contents: """
		# Code generated by github.com/heetch/cue-schema/github/workflow/generate. DO NOT EDIT.
		\(yaml.Marshal(ci))
		"""
	}
}
