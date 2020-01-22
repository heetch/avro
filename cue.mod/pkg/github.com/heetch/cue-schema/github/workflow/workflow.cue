// Package workflow describes the contents of a file in the
// .github/workflows directory of a repository.
//
// This follows the API described here:
// https://help.github.com/en/actions/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions
package workflow

// TODO cross-verify against https://github.com/SchemaStore/schemastore/blob/master/src/schemas/json/github-workflow.json

import "regexp"

// name holds the name of your workflow. GitHub displays the
// names of your workflows on your repository's actions page. If
// you omit this field, GitHub sets the name to the workflow's
// filename.
name?: string

// On describes the event or events that trigger the workflow.
on: Event | [... Event] | EventConfig

env?: Env
jobs: {
	// TODO must start with a letter or _ and contain only alphanumeric characters, -, or _.
	<jobID>: Job
}

// Job represents one of the jobs to do as part of the action.
Job :: {
	name?: string
	// TODO restrict JobID to names mentioned in the workflow jobs?
	needs?:     JobID | [...JobID]
	"runs-on"?: string | [string, ...string]
	env?:       Env
	if?:        Condition
	steps?: [... JobStep]
	"timeout-minutes"?: number
	container?: {
		image?: string
		env?:   Env
		ports?: [ ... int]
		volumes?: [ ... string]
		options?: _ // TODO more specific type here
	}
	strategy?: {
		matrix: {
			<var>: [ ...]
		}
		"fail-fast"?:    bool
		"max-parallel"?: int & >0
	}
	services: {
		[_]: Service
	}
}

Service :: {
	image: string
	env?:  Env
	ports?: [... string]
	volumes?: [ ... string]
	options?: _ // TODO more specific type here

}

JobStep :: {
	id?:                  string
	if?:                  Condition
	name?:                string
	uses?:                string
	run?:                 string
	"working-directory"?: string
	shell?:
		"bash" |
		"pwsh" |
		"python" |
		"sh" |
		"cmd" |
		"powershell" |
		regexp.Match(#"\{0\}"#)

	// with holds a map of the input parameters defined by the action.
	// Each input parameter is a key/value pair. Input parameters are
	// set as environment variables. The variable is prefixed with INPUT_
	// and converted to upper case.
	with?: {<_>: string}
	env?:                 Env
	"continue-on-error"?: bool
	"timeout-minutes"?:   number
}

// JobID represents one the of jobs specified in Workflow.
JobID :: string

// Condition represents a condition to evaluate.
// TODO link to syntax.
Condition :: string

// Env represents a set of environment variables and their values.
Env :: {
	<_>: string
}
