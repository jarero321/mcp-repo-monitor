package port

import (
	"context"

	"github.com/carlos/mcp-repo-monitor/internal/domain/entity"
)

type GitHubClient interface {
	ListRepositories(ctx context.Context, filter string, includeArchived bool) ([]entity.Repository, error)
	GetRepository(ctx context.Context, owner, repo string) (*entity.Repository, error)

	ListPullRequests(ctx context.Context, filter entity.PRFilter) ([]entity.PullRequest, error)
	GetPullRequest(ctx context.Context, owner, repo string, number int) (*entity.PullRequest, error)
	CreatePullRequest(ctx context.Context, owner, repo, title, body, head, base string) (*entity.PullRequest, error)

	ListCommits(ctx context.Context, filter entity.CommitFilter) ([]entity.Commit, error)

	ListWorkflowRuns(ctx context.Context, filter entity.CIFilter) ([]entity.WorkflowRun, error)
	RerunWorkflow(ctx context.Context, owner, repo string, runID int64) error
	TriggerWorkflow(ctx context.Context, owner, repo, workflowID, ref string) error

	CompareBranches(ctx context.Context, owner, repo, base, head string) (*entity.BranchComparison, error)

	GetCurrentUser(ctx context.Context) (string, error)
}
