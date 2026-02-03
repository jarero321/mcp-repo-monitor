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
