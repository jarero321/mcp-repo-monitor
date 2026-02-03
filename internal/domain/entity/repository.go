package entity

import "time"

type Repository struct {
	ID          int64
	Name        string
	FullName    string
	Description string
	Private     bool
	Archived    bool
	Fork        bool
	DefaultBranch string
	Language    string
	Stars       int
	Forks       int
	OpenIssues  int
	HTMLURL     string
	CloneURL    string
	UpdatedAt   time.Time
	PushedAt    time.Time
}

type RepositoryStatus struct {
	Repository    Repository
	OpenPRs       int
	FailedCI      bool
	LastCommitAt  time.Time
	HasDrift      bool
}
