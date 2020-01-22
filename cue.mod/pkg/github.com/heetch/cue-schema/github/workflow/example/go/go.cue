package go
import (
	goworkflow "github.com/heetch/cue-schema/github/workflow/go:workflow"
)

Workflow :: goworkflow
Workflow :: {
	name: "My CI caboodle"
	Versions :: ["v1.12", "v1.13"]
	Services :: ["postgres", "kafka"]
}
