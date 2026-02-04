package port

import (
	"context"

	"github.com/carlos/mcp-repo-monitor/internal/domain/entity"
)

// MockGitHubClient is a mock implementation of GitHubClient for testing.
type MockGitHubClient struct {
	// Function stubs for each method
	ListRepositoriesFunc   func(ctx context.Context, filter string, includeArchived bool) ([]entity.Repository, error)
	GetRepositoryFunc      func(ctx context.Context, owner, repo string) (*entity.Repository, error)
	ListPullRequestsFunc   func(ctx context.Context, filter entity.PRFilter) ([]entity.PullRequest, error)
	GetPullRequestFunc     func(ctx context.Context, owner, repo string, number int) (*entity.PullRequest, error)
	CreatePullRequestFunc  func(ctx context.Context, owner, repo, title, body, head, base string) (*entity.PullRequest, error)
	ListCommitsFunc        func(ctx context.Context, filter entity.CommitFilter) ([]entity.Commit, error)
	ListWorkflowRunsFunc   func(ctx context.Context, filter entity.CIFilter) ([]entity.WorkflowRun, error)
	RerunWorkflowFunc      func(ctx context.Context, owner, repo string, runID int64) error
	TriggerWorkflowFunc    func(ctx context.Context, owner, repo, workflowID, ref string) error
	CompareBranchesFunc    func(ctx context.Context, owner, repo, base, head string) (*entity.BranchComparison, error)
	GetCurrentUserFunc     func(ctx context.Context) (string, error)

	// Call tracking
	CompareBranchesCalls []CompareBranchesCall
	CreatePRCalls        []CreatePRCall
}

type CompareBranchesCall struct {
	Owner, Repo, Base, Head string
}

type CreatePRCall struct {
	Owner, Repo, Title, Body, Head, Base string
}

func NewMockGitHubClient() *MockGitHubClient {
	return &MockGitHubClient{
		CompareBranchesCalls: []CompareBranchesCall{},
		CreatePRCalls:        []CreatePRCall{},
	}
}

func (m *MockGitHubClient) ListRepositories(ctx context.Context, filter string, includeArchived bool) ([]entity.Repository, error) {
	if m.ListRepositoriesFunc != nil {
		return m.ListRepositoriesFunc(ctx, filter, includeArchived)
	}
	return []entity.Repository{}, nil
}

func (m *MockGitHubClient) GetRepository(ctx context.Context, owner, repo string) (*entity.Repository, error) {
	if m.GetRepositoryFunc != nil {
		return m.GetRepositoryFunc(ctx, owner, repo)
	}
	return &entity.Repository{}, nil
}

func (m *MockGitHubClient) ListPullRequests(ctx context.Context, filter entity.PRFilter) ([]entity.PullRequest, error) {
	if m.ListPullRequestsFunc != nil {
		return m.ListPullRequestsFunc(ctx, filter)
	}
	return []entity.PullRequest{}, nil
}

func (m *MockGitHubClient) GetPullRequest(ctx context.Context, owner, repo string, number int) (*entity.PullRequest, error) {
	if m.GetPullRequestFunc != nil {
		return m.GetPullRequestFunc(ctx, owner, repo, number)
	}
	return &entity.PullRequest{}, nil
}

func (m *MockGitHubClient) CreatePullRequest(ctx context.Context, owner, repo, title, body, head, base string) (*entity.PullRequest, error) {
	m.CreatePRCalls = append(m.CreatePRCalls, CreatePRCall{
		Owner: owner,
		Repo:  repo,
		Title: title,
		Body:  body,
		Head:  head,
		Base:  base,
	})
	if m.CreatePullRequestFunc != nil {
		return m.CreatePullRequestFunc(ctx, owner, repo, title, body, head, base)
	}
	return &entity.PullRequest{Number: 1, HTMLURL: "https://github.com/test/repo/pull/1"}, nil
}

func (m *MockGitHubClient) ListCommits(ctx context.Context, filter entity.CommitFilter) ([]entity.Commit, error) {
	if m.ListCommitsFunc != nil {
		return m.ListCommitsFunc(ctx, filter)
	}
	return []entity.Commit{}, nil
}

func (m *MockGitHubClient) ListWorkflowRuns(ctx context.Context, filter entity.CIFilter) ([]entity.WorkflowRun, error) {
	if m.ListWorkflowRunsFunc != nil {
		return m.ListWorkflowRunsFunc(ctx, filter)
	}
	return []entity.WorkflowRun{}, nil
}

func (m *MockGitHubClient) RerunWorkflow(ctx context.Context, owner, repo string, runID int64) error {
	if m.RerunWorkflowFunc != nil {
		return m.RerunWorkflowFunc(ctx, owner, repo, runID)
	}
	return nil
}

func (m *MockGitHubClient) TriggerWorkflow(ctx context.Context, owner, repo, workflowID, ref string) error {
	if m.TriggerWorkflowFunc != nil {
		return m.TriggerWorkflowFunc(ctx, owner, repo, workflowID, ref)
	}
	return nil
}

func (m *MockGitHubClient) CompareBranches(ctx context.Context, owner, repo, base, head string) (*entity.BranchComparison, error) {
	m.CompareBranchesCalls = append(m.CompareBranchesCalls, CompareBranchesCall{
		Owner: owner,
		Repo:  repo,
		Base:  base,
		Head:  head,
	})
	if m.CompareBranchesFunc != nil {
		return m.CompareBranchesFunc(ctx, owner, repo, base, head)
	}
	return &entity.BranchComparison{}, nil
}

func (m *MockGitHubClient) GetCurrentUser(ctx context.Context) (string, error) {
	if m.GetCurrentUserFunc != nil {
		return m.GetCurrentUserFunc(ctx)
	}
	return "test-user", nil
}
