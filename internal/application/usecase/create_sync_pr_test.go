package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/carlos/mcp-repo-monitor/config"
	"github.com/carlos/mcp-repo-monitor/internal/application/port"
	"github.com/carlos/mcp-repo-monitor/internal/domain/entity"
)

func TestCreateSyncPRUseCase_Execute_Success(t *testing.T) {
	mockClient := port.NewMockGitHubClient()
	mockClient.CompareBranchesFunc = func(ctx context.Context, owner, repo, base, head string) (*entity.BranchComparison, error) {
		return &entity.BranchComparison{
			TotalCommits: 5,
			Files:        []entity.ChangedFile{{Filename: "test.go"}},
		}, nil
	}
	mockClient.CreatePullRequestFunc = func(ctx context.Context, owner, repo, title, body, head, base string, draft bool) (*entity.PullRequest, error) {
		return &entity.PullRequest{
			Number:  42,
			HTMLURL: "https://github.com/test/repo/pull/42",
		}, nil
	}

	cfg := &config.Config{
		ReposConfig: config.ReposConfig{
			Default: config.BranchConfig{
				ProdBranch: "main",
				DevBranch:  "develop",
			},
			Repositories: make(map[string]config.BranchConfig),
		},
	}

	uc := NewCreateSyncPRUseCase(mockClient, cfg)

	result, err := uc.Execute(context.Background(), CreateSyncPRInput{
		Repository: "test-owner/test-repo",
	})

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !result.Success {
		t.Error("Execute() Success = false, want true")
	}

	if result.PRNumber != 42 {
		t.Errorf("PRNumber = %d, want 42", result.PRNumber)
	}

	// Verify CompareBranches was called with correct order (ProdBranch first, DevBranch second)
	if len(mockClient.CompareBranchesCalls) != 1 {
		t.Fatalf("CompareBranches called %d times, want 1", len(mockClient.CompareBranchesCalls))
	}

	call := mockClient.CompareBranchesCalls[0]
	if call.Base != "main" || call.Head != "develop" {
		t.Errorf("CompareBranches called with base=%s head=%s, want base=main head=develop", call.Base, call.Head)
	}

	// Verify CreatePullRequest was called correctly
	if len(mockClient.CreatePRCalls) != 1 {
		t.Fatalf("CreatePullRequest called %d times, want 1", len(mockClient.CreatePRCalls))
	}

	prCall := mockClient.CreatePRCalls[0]
	if prCall.Head != "main" || prCall.Base != "develop" {
		t.Errorf("CreatePullRequest called with head=%s base=%s, want head=main base=develop", prCall.Head, prCall.Base)
	}
}

func TestCreateSyncPRUseCase_Execute_AlreadySynced(t *testing.T) {
	mockClient := port.NewMockGitHubClient()
	mockClient.CompareBranchesFunc = func(ctx context.Context, owner, repo, base, head string) (*entity.BranchComparison, error) {
		return &entity.BranchComparison{
			TotalCommits: 0,
		}, nil
	}

	cfg := &config.Config{
		ReposConfig: config.ReposConfig{
			Default: config.BranchConfig{
				ProdBranch: "main",
				DevBranch:  "develop",
			},
			Repositories: make(map[string]config.BranchConfig),
		},
	}

	uc := NewCreateSyncPRUseCase(mockClient, cfg)

	result, err := uc.Execute(context.Background(), CreateSyncPRInput{
		Repository: "test/repo",
	})

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !result.Success {
		t.Error("Execute() Success = false, want true")
	}

	if !strings.Contains(result.Message, "already in sync") {
		t.Errorf("Message = %s, want to contain 'already in sync'", result.Message)
	}

	// Should not create PR
	if len(mockClient.CreatePRCalls) != 0 {
		t.Errorf("CreatePullRequest called %d times, want 0", len(mockClient.CreatePRCalls))
	}
}

func TestCreateSyncPRUseCase_Execute_MergeCommitsNoFiles(t *testing.T) {
	mockClient := port.NewMockGitHubClient()
	mockClient.CompareBranchesFunc = func(ctx context.Context, owner, repo, base, head string) (*entity.BranchComparison, error) {
		// Merge commits but no file changes — should be treated as synced
		return &entity.BranchComparison{
			TotalCommits: 3,
			BehindBy:     3,
			GitHubStatus: "behind",
			Files:        nil,
		}, nil
	}

	cfg := &config.Config{
		ReposConfig: config.ReposConfig{
			Default: config.BranchConfig{
				ProdBranch: "main",
				DevBranch:  "develop",
			},
			Repositories: make(map[string]config.BranchConfig),
		},
	}

	uc := NewCreateSyncPRUseCase(mockClient, cfg)

	result, err := uc.Execute(context.Background(), CreateSyncPRInput{
		Repository: "test/repo",
	})

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !result.Success {
		t.Error("Execute() Success = false, want true")
	}

	if !strings.Contains(result.Message, "already in sync") {
		t.Errorf("Message = %s, want to contain 'already in sync'", result.Message)
	}

	// Should not create PR when only merge commits (no file changes)
	if len(mockClient.CreatePRCalls) != 0 {
		t.Errorf("CreatePullRequest called %d times, want 0", len(mockClient.CreatePRCalls))
	}
}

func TestCreateSyncPRUseCase_Execute_DryRun(t *testing.T) {
	mockClient := port.NewMockGitHubClient()
	mockClient.CompareBranchesFunc = func(ctx context.Context, owner, repo, base, head string) (*entity.BranchComparison, error) {
		return &entity.BranchComparison{
			TotalCommits: 3,
			Files:        []entity.ChangedFile{{Filename: "a.go"}, {Filename: "b.go"}},
		}, nil
	}

	cfg := &config.Config{
		ReposConfig: config.ReposConfig{
			Default: config.BranchConfig{
				ProdBranch: "main",
				DevBranch:  "develop",
			},
			Repositories: make(map[string]config.BranchConfig),
		},
	}

	uc := NewCreateSyncPRUseCase(mockClient, cfg)

	result, err := uc.Execute(context.Background(), CreateSyncPRInput{
		Repository: "test/repo",
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

	if result.Commits != 3 {
		t.Errorf("Commits = %d, want 3", result.Commits)
	}

	if result.FilesChanged != 2 {
		t.Errorf("FilesChanged = %d, want 2", result.FilesChanged)
	}

	// Should not create PR in dry run
	if len(mockClient.CreatePRCalls) != 0 {
		t.Errorf("CreatePullRequest called %d times, want 0", len(mockClient.CreatePRCalls))
	}
}

func TestCreateSyncPRUseCase_Execute_CustomTitle(t *testing.T) {
	mockClient := port.NewMockGitHubClient()
	mockClient.CompareBranchesFunc = func(ctx context.Context, owner, repo, base, head string) (*entity.BranchComparison, error) {
		return &entity.BranchComparison{TotalCommits: 1, Files: []entity.ChangedFile{{Filename: "a.go"}}}, nil
	}
	mockClient.CreatePullRequestFunc = func(ctx context.Context, owner, repo, title, body, head, base string, draft bool) (*entity.PullRequest, error) {
		return &entity.PullRequest{Number: 1}, nil
	}

	cfg := &config.Config{
		ReposConfig: config.ReposConfig{
			Default: config.BranchConfig{
				ProdBranch: "main",
				DevBranch:  "develop",
			},
			Repositories: make(map[string]config.BranchConfig),
		},
	}

	uc := NewCreateSyncPRUseCase(mockClient, cfg)

	_, err := uc.Execute(context.Background(), CreateSyncPRInput{
		Repository: "test/repo",
		Title:      "Custom PR Title",
		Body:       "Custom body",
	})

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	prCall := mockClient.CreatePRCalls[0]
	if prCall.Title != "Custom PR Title" {
		t.Errorf("Title = %s, want 'Custom PR Title'", prCall.Title)
	}
	if prCall.Body != "Custom body" {
		t.Errorf("Body = %s, want 'Custom body'", prCall.Body)
	}
}

func TestCreateSyncPRUseCase_Execute_InvalidRepoFormat(t *testing.T) {
	mockClient := port.NewMockGitHubClient()
	cfg := &config.Config{
		ReposConfig: config.ReposConfig{
			Default: config.BranchConfig{
				ProdBranch: "main",
				DevBranch:  "develop",
			},
			Repositories: make(map[string]config.BranchConfig),
		},
	}

	uc := NewCreateSyncPRUseCase(mockClient, cfg)

	_, err := uc.Execute(context.Background(), CreateSyncPRInput{
		Repository: "invalid-format",
	})

	if err == nil {
		t.Error("Execute() error = nil, want error for invalid repo format")
	}
}

func TestCreateSyncPRUseCase_Execute_CompareBranchesError(t *testing.T) {
	mockClient := port.NewMockGitHubClient()
	mockClient.CompareBranchesFunc = func(ctx context.Context, owner, repo, base, head string) (*entity.BranchComparison, error) {
		return nil, errors.New("API error")
	}

	cfg := &config.Config{
		ReposConfig: config.ReposConfig{
			Default: config.BranchConfig{
				ProdBranch: "main",
				DevBranch:  "develop",
			},
			Repositories: make(map[string]config.BranchConfig),
		},
	}

	uc := NewCreateSyncPRUseCase(mockClient, cfg)

	_, err := uc.Execute(context.Background(), CreateSyncPRInput{
		Repository: "test/repo",
	})

	if err == nil {
		t.Error("Execute() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "compare branches") {
		t.Errorf("error = %v, want to contain 'compare branches'", err)
	}
}

func TestCreateSyncPRUseCase_Execute_CreatePRError(t *testing.T) {
	mockClient := port.NewMockGitHubClient()
	mockClient.CompareBranchesFunc = func(ctx context.Context, owner, repo, base, head string) (*entity.BranchComparison, error) {
		return &entity.BranchComparison{TotalCommits: 1, Files: []entity.ChangedFile{{Filename: "a.go"}}}, nil
	}
	mockClient.CreatePullRequestFunc = func(ctx context.Context, owner, repo, title, body, head, base string, draft bool) (*entity.PullRequest, error) {
		return nil, errors.New("PR creation failed")
	}

	cfg := &config.Config{
		ReposConfig: config.ReposConfig{
			Default: config.BranchConfig{
				ProdBranch: "main",
				DevBranch:  "develop",
			},
			Repositories: make(map[string]config.BranchConfig),
		},
	}

	uc := NewCreateSyncPRUseCase(mockClient, cfg)

	_, err := uc.Execute(context.Background(), CreateSyncPRInput{
		Repository: "test/repo",
	})

	if err == nil {
		t.Error("Execute() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "create PR") {
		t.Errorf("error = %v, want to contain 'create PR'", err)
	}
}

func TestCreateSyncPRUseCase_Execute_CustomBranchConfig(t *testing.T) {
	mockClient := port.NewMockGitHubClient()
	mockClient.CompareBranchesFunc = func(ctx context.Context, owner, repo, base, head string) (*entity.BranchComparison, error) {
		return &entity.BranchComparison{TotalCommits: 1, Files: []entity.ChangedFile{{Filename: "a.go"}}}, nil
	}
	mockClient.CreatePullRequestFunc = func(ctx context.Context, owner, repo, title, body, head, base string, draft bool) (*entity.PullRequest, error) {
		return &entity.PullRequest{Number: 1}, nil
	}

	cfg := &config.Config{
		ReposConfig: config.ReposConfig{
			Default: config.BranchConfig{
				ProdBranch: "main",
				DevBranch:  "develop",
			},
			Repositories: map[string]config.BranchConfig{
				"custom/repo": {
					ProdBranch: "production",
					DevBranch:  "staging",
				},
			},
		},
	}

	uc := NewCreateSyncPRUseCase(mockClient, cfg)

	_, err := uc.Execute(context.Background(), CreateSyncPRInput{
		Repository: "custom/repo",
	})

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify custom branch config was used
	compareCall := mockClient.CompareBranchesCalls[0]
	if compareCall.Base != "production" || compareCall.Head != "staging" {
		t.Errorf("CompareBranches called with base=%s head=%s, want base=production head=staging", compareCall.Base, compareCall.Head)
	}

	prCall := mockClient.CreatePRCalls[0]
	if prCall.Head != "production" || prCall.Base != "staging" {
		t.Errorf("CreatePullRequest called with head=%s base=%s, want head=production base=staging", prCall.Head, prCall.Base)
	}
}
