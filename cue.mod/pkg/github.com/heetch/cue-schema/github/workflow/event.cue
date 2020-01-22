package workflow

Event ::
	"check_run" |
	"check_suite" |
	"commit_comment" |
	"create" |
	"delete" |
	"deployment" |
	"deployment_status" |
	"fork" |
	"gollum" |
	"issue_comment" |
	"issues" |
	"label" |
	"member" |
	"milestone" |
	"page_build" |
	"project" |
	"project_card" |
	"project_column" |
	"public" |
	"pull_request" |
	"pull_request_review" |
	"pull_request_review_comment" |
	"push" |
	"release" |
	"repository_dispatch" |
	"schedule" |
	"status" |
	"watch"

EventConfig :: {
	check_run?: null | {
		type =
		"created" |
		"rerequested" |
		"completed" |
		"requested_action"
		types?: [ ... type]
	}
	check_suite?: null | {
		type =
		"assigned" |
		"unassigned" |
		"labeled" |
		"unlabeled" |
		"opened" |
		"edited" |
		"closed" |
		"reopened" |
		"synchronize" |
		"ready_for_review" |
		"locked" |
		"unlocked " |
		"review_requested " |
		"review_request_removed"
		types?: [ ... type]
	}
	commit_comment?:    null | {}
	create?:            null | {}
	delete?:            null | {}
	deployment?:        null | {}
	deployment_status?: null | {}
	fork?:              null | {}
	gollum?:            null | {}
	issue_comment?:     null | {
		type =
		"created" |
		"edited" |
		"deleted"
		types?: [ ... type]
	}
	issues?: null | {
		type =
		"opened" |
		"edited" |
		"deleted" |
		"transferred" |
		"pinned" |
		"unpinned" |
		"closed" |
		"reopened" |
		"assigned" |
		"unassigned" |
		"labeled" |
		"unlabeled" |
		"locked" |
		"unlocked" |
		"milestoned" |
		"demilestoned"
		types?: [ ... type]
	}
	label?: null | {
		type =
		"created" |
		"edited" |
		"deleted"
		types?: [ ... type]
	}
	member?: null | {
		type =
		"added" |
		"edited" |
		"deleted"
		types?: [ ... type]
	}
	milestone?: null | {
		type =
		"created" |
		"closed" |
		"opened" |
		"edited" |
		"deleted"
		types?: [ ... type]
	}
	page_build?: null | {}
	project?:    null | {
		type =
		"created" |
		"closed" |
		"opened" |
		"edited" |
		"deleted"
		types?: [ ... type]
	}
	project_card?: null | {
		type =
		"created" |
		"moved" |
		"converted" |
		"edited" |
		"deleted"
		types?: [ ... type]
	}
	project_column?: null | {
		type =
		"created" |
		"updated" |
		"moved" |
		"deleted"
		types?: [ ... type]
	}
	public?:       null | {}
	pull_request?: null | {
		PushPullEvent
		type =
		"assigned" |
		"unassigned" |
		"labeled" |
		"unlabeled" |
		"opened" |
		"edited" |
		"closed" |
		"reopened" |
		"synchronize" |
		"ready_for_review" |
		"locked" |
		"unlocked " |
		"review_requested " |
		"review_request_removed"
		types?: [ ... type]
	}
	pull_request_review?: null | {
		type =
		"submitted" |
		"edited" |
		"dismissed"
		types?: [ ... type]
	}
	pull_request_review_comment?: null | {
		type =
		"created" |
		"edited" |
		"deleted"
		types?: [ ... type]
	}
	push?: null | {
		PushPullEvent
	}
	release?: null | {
		type =
		"published " |
		"unpublished " |
		"created " |
		"edited " |
		"deleted " |
		"prereleased"
		types? : [ ... type]
	}
	repository_dispatch?: null | {}
	// schedule configures a workflow to run at specific UTC times using POSIX
	// cron syntax. Scheduled workflows run on the latest commit on the
	// default or base branch.
	schedule?: null | [{
		// cron specifies the time schedule in cron syntax.
		// See https://help.github.com/en/articles/events-that-trigger-workflows#scheduled-events
		// TODO regexp for cron syntax
		cron: string
	}]
	status?: null | {}
	watch?:  null | {
		types?: [ "started"]
	}
}

PushPullEvent :: {
	branches?: [... Glob]
	tags?: [... Glob]
	"branches-ignore"?: [... Glob]
	"tags-ignore"?: [... Glob]
	paths?: [... Glob]
}

// Glob represents a wildcard pattern.
// See https://help.github.com/en/articles/workflow-syntax-for-github-actions#filter-pattern-cheat-sheet
Glob :: string
