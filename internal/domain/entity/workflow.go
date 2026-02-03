package entity

import "time"

type WorkflowRun struct {
	ID           int64
	Name         string
	WorkflowID   int64
	WorkflowName string
	HeadBranch   string
	HeadSHA      string
	Status       string
	Conclusion   string
	HTMLURL      string
	RunNumber    int
	RunAttempt   int
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Repository   string
	Actor        string
	Event        string
}

type Workflow struct {
	ID        int64
	Name      string
	Path      string
	State     string
	HTMLURL   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CIFilter struct {
	Repository string
	Branch     string
	Workflow   string
	Limit      int
}

type RollbackStrategy string

const (
	RollbackRerun     RollbackStrategy = "rerun"
	RollbackRevert    RollbackStrategy = "revert"
	RollbackWorkflow  RollbackStrategy = "workflow"
)

type RollbackRequest struct {
	Repository  string
	Strategy    RollbackStrategy
	WorkflowID  int64
	RunID       int64
	CommitSHA   string
	DryRun      bool
}

type RollbackResult struct {
	Success    bool
	Strategy   RollbackStrategy
	Message    string
	RunURL     string
	PRURL      string
}
