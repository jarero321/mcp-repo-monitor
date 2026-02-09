package entity

import "time"

type PullRequest struct {
	ID          int64
	Number      int
	Title       string
	Body        string
	State       string
	Draft       bool
	HTMLURL     string
	User        string
	HeadBranch  string
	BaseBranch  string
	Mergeable   *bool
	Additions   int
	Deletions   int
	ChangedFiles int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	MergedAt    *time.Time
	ClosedAt    *time.Time
	Labels      []string
	Reviewers   []string
	Repository  string
}

type PRFilter struct {
	Repository string
	State      string
	Limit      int
}

type MergeMethod string

const (
	MergeMethodMerge  MergeMethod = "merge"
	MergeMethodSquash MergeMethod = "squash"
	MergeMethodRebase MergeMethod = "rebase"
)

type MergeResult struct {
	Success       bool
	SHA           string
	Message       string
	PRURL         string
	PRNumber      int
	MergeMethod   MergeMethod
	BranchDeleted bool
	BranchName    string
}
