package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/carlos/mcp-repo-monitor/internal/application/port"
	"github.com/carlos/mcp-repo-monitor/internal/domain/entity"
)

func TestCreatePRUseCase_Execute_Success(t *testing.T) {
	mockClient := port.NewMockGitHubClient()
	mockClient.CreatePullRequestFunc = func(ctx context.Context, owner, repo, title, body, head, base string, draft bool) (*entity.PullRequest, error) {
		return &entity.PullRequest{
			Number:     42,
			HTMLURL:    "https://github.com/test/repo/pull/42",
			HeadBranch: head,
			BaseBranch: base,
			Draft:      draft,
		}, nil
	}

	uc := NewCreatePRUseCase(mockClient)

	result, err := uc.Execute(context.Background(), CreatePRInput{
		Repository: "test-owner/test-repo",
		Title:      "feat: add new feature",
		Head:       "feature/new",
		Base:       "main",
		Body:       "## Summary\nNew feature",
	})

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !result.Success {
		t.Error("Execute() Success = false, want true")
	}

	if result.PR.Number != 42 {
		t.Errorf("PR.Number = %d, want 42", result.PR.Number)
	}

	if len(mockClient.CreatePRCalls) != 1 {
		t.Fatalf("CreatePullRequest called %d times, want 1", len(mockClient.CreatePRCalls))
	}

	call := mockClient.CreatePRCalls[0]
	if call.Owner != "test-owner" || call.Repo != "test-repo" {
		t.Errorf("called with owner=%s repo=%s, want test-owner/test-repo", call.Owner, call.Repo)
	}
	if call.Head != "feature/new" || call.Base != "main" {
		t.Errorf("called with head=%s base=%s, want feature/new main", call.Head, call.Base)
	}
	if call.Draft != false {
		t.Error("called with draft=true, want false")
	}
}

func TestCreatePRUseCase_Execute_Draft(t *testing.T) {
	mockClient := port.NewMockGitHubClient()
	mockClient.CreatePullRequestFunc = func(ctx context.Context, owner, repo, title, body, head, base string, draft bool) (*entity.PullRequest, error) {
		return &entity.PullRequest{
			Number:  10,
			HTMLURL: "https://github.com/test/repo/pull/10",
			Draft:   draft,
		}, nil
	}

	uc := NewCreatePRUseCase(mockClient)

	result, err := uc.Execute(context.Background(), CreatePRInput{
		Repository: "test/repo",
		Title:      "wip: draft PR",
		Head:       "feature/wip",
		Base:       "main",
		Draft:      true,
	})

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !result.Success {
		t.Error("Execute() Success = false, want true")
	}

	call := mockClient.CreatePRCalls[0]
	if !call.Draft {
		t.Error("called with draft=false, want true")
	}
}

func TestCreatePRUseCase_Execute_DryRun(t *testing.T) {
	mockClient := port.NewMockGitHubClient()
	mockClient.CompareBranchesFunc = func(ctx context.Context, owner, repo, base, head string) (*entity.BranchComparison, error) {
		return &entity.BranchComparison{
			TotalCommits: 5,
			Files:        []entity.ChangedFile{{Filename: "a.go"}, {Filename: "b.go"}},
		}, nil
	}

	uc := NewCreatePRUseCase(mockClient)

	result, err := uc.Execute(context.Background(), CreatePRInput{
		Repository: "test/repo",
		Title:      "test PR",
		Head:       "feature/x",
		Base:       "main",
		DryRun:     true,
	})

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !result.Success {
		t.Error("Execute() Success = false, want true")
	}

	if !strings.Contains(result.Message, "DRY RUN") {
		t.Errorf("Message = %s, want to contain 'DRY RUN'", result.Message)
	}

	if result.Commits != 5 {
		t.Errorf("Commits = %d, want 5", result.Commits)
	}

	if result.FilesChanged != 2 {
		t.Errorf("FilesChanged = %d, want 2", result.FilesChanged)
	}

	if len(mockClient.CreatePRCalls) != 0 {
		t.Errorf("CreatePullRequest called %d times, want 0", len(mockClient.CreatePRCalls))
	}
}

func TestCreatePRUseCase_Execute_DryRunDraft(t *testing.T) {
	mockClient := port.NewMockGitHubClient()
	mockClient.CompareBranchesFunc = func(ctx context.Context, owner, repo, base, head string) (*entity.BranchComparison, error) {
		return &entity.BranchComparison{TotalCommits: 1}, nil
	}

	uc := NewCreatePRUseCase(mockClient)

	result, err := uc.Execute(context.Background(), CreatePRInput{
		Repository: "test/repo",
		Title:      "test PR",
		Head:       "feature/x",
		Base:       "main",
		Draft:      true,
		DryRun:     true,
	})

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !strings.Contains(result.Message, "(draft)") {
		t.Errorf("Message = %s, want to contain '(draft)'", result.Message)
	}
}

func TestCreatePRUseCase_Execute_InvalidRepoFormat(t *testing.T) {
	mockClient := port.NewMockGitHubClient()
	uc := NewCreatePRUseCase(mockClient)

	_, err := uc.Execute(context.Background(), CreatePRInput{
		Repository: "invalid-format",
		Title:      "test",
		Head:       "feature",
		Base:       "main",
	})

	if err == nil {
		t.Error("Execute() error = nil, want error for invalid repo format")
	}
}

func TestCreatePRUseCase_Execute_MissingTitle(t *testing.T) {
	mockClient := port.NewMockGitHubClient()
	uc := NewCreatePRUseCase(mockClient)

	_, err := uc.Execute(context.Background(), CreatePRInput{
		Repository: "test/repo",
		Head:       "feature",
		Base:       "main",
	})

	if err == nil {
		t.Error("Execute() error = nil, want error for missing title")
	}

	if !strings.Contains(err.Error(), "title") {
		t.Errorf("error = %v, want to contain 'title'", err)
	}
}

func TestCreatePRUseCase_Execute_MissingHead(t *testing.T) {
	mockClient := port.NewMockGitHubClient()
	uc := NewCreatePRUseCase(mockClient)

	_, err := uc.Execute(context.Background(), CreatePRInput{
		Repository: "test/repo",
		Title:      "test",
		Base:       "main",
	})

	if err == nil {
		t.Error("Execute() error = nil, want error for missing head")
	}

	if !strings.Contains(err.Error(), "head") {
		t.Errorf("error = %v, want to contain 'head'", err)
	}
}

func TestCreatePRUseCase_Execute_MissingBase(t *testing.T) {
	mockClient := port.NewMockGitHubClient()
	uc := NewCreatePRUseCase(mockClient)

	_, err := uc.Execute(context.Background(), CreatePRInput{
		Repository: "test/repo",
		Title:      "test",
		Head:       "feature",
	})

	if err == nil {
		t.Error("Execute() error = nil, want error for missing base")
	}

	if !strings.Contains(err.Error(), "base") {
		t.Errorf("error = %v, want to contain 'base'", err)
	}
}

func TestCreatePRUseCase_Execute_APIError(t *testing.T) {
	mockClient := port.NewMockGitHubClient()
	mockClient.CreatePullRequestFunc = func(ctx context.Context, owner, repo, title, body, head, base string, draft bool) (*entity.PullRequest, error) {
		return nil, errors.New("422 Validation Failed")
	}

	uc := NewCreatePRUseCase(mockClient)

	_, err := uc.Execute(context.Background(), CreatePRInput{
		Repository: "test/repo",
		Title:      "test PR",
		Head:       "feature/x",
		Base:       "main",
	})

	if err == nil {
		t.Error("Execute() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "create PR") {
		t.Errorf("error = %v, want to contain 'create PR'", err)
	}
}
