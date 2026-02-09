package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/carlos/mcp-repo-monitor/config"
	"github.com/carlos/mcp-repo-monitor/internal/application/port"
	"github.com/carlos/mcp-repo-monitor/internal/domain/entity"
	"github.com/carlos/mcp-repo-monitor/internal/domain/service"
)

func TestCheckDriftUseCase_Execute_SingleRepo(t *testing.T) {
	mockClient := port.NewMockGitHubClient()
	mockClient.CompareBranchesFunc = func(ctx context.Context, owner, repo, base, head string) (*entity.BranchComparison, error) {
		return &entity.BranchComparison{
			Repository:   owner + "/" + repo,
			ProdBranch:   base,
			DevBranch:    head,
			AheadBy:      5,
			BehindBy:     0,
			TotalCommits: 5,
			Files:        []entity.ChangedFile{{Filename: "main.go"}},
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

	driftDetector := service.NewDriftDetector()
	uc := NewCheckDriftUseCase(mockClient, cfg, driftDetector)

	results, err := uc.Execute(context.Background(), CheckDriftInput{
		Repository: "test-owner/test-repo",
	})

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Execute() returned %d results, want 1", len(results))
	}

	// Verify CompareBranches was called with correct order (ProdBranch, DevBranch)
	if len(mockClient.CompareBranchesCalls) != 1 {
		t.Fatalf("CompareBranches called %d times, want 1", len(mockClient.CompareBranchesCalls))
	}

	call := mockClient.CompareBranchesCalls[0]
	if call.Base != "main" || call.Head != "develop" {
		t.Errorf("CompareBranches called with base=%s head=%s, want base=main head=develop", call.Base, call.Head)
	}

	// Verify drift analysis
	if results[0].Severity != "low" {
		t.Errorf("Severity = %s, want low", results[0].Severity)
	}

	if results[0].Comparison.Status != entity.DriftProdAhead {
		t.Errorf("Status = %v, want %v", results[0].Comparison.Status, entity.DriftProdAhead)
	}
}

func TestCheckDriftUseCase_Execute_AllRepos(t *testing.T) {
	mockClient := port.NewMockGitHubClient()
	mockClient.ListRepositoriesFunc = func(ctx context.Context, filter string, includeArchived bool) ([]entity.Repository, error) {
		return []entity.Repository{
			{FullName: "org/repo1"},
			{FullName: "org/repo2"},
		}, nil
	}
	mockClient.CompareBranchesFunc = func(ctx context.Context, owner, repo, base, head string) (*entity.BranchComparison, error) {
		if repo == "repo1" {
			return &entity.BranchComparison{
				Repository:   owner + "/" + repo,
				TotalCommits: 3,
				AheadBy:      3,
				Files:        []entity.ChangedFile{{Filename: "a.go"}},
			}, nil
		}
		return &entity.BranchComparison{
			Repository:   owner + "/" + repo,
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

	driftDetector := service.NewDriftDetector()
	uc := NewCheckDriftUseCase(mockClient, cfg, driftDetector)

	results, err := uc.Execute(context.Background(), CheckDriftInput{})

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Only repo1 has drift (has file changes)
	if len(results) != 1 {
		t.Fatalf("Execute() returned %d results, want 1 (only repos with drift)", len(results))
	}

	if results[0].Comparison.Repository != "org/repo1" {
		t.Errorf("Repository = %s, want org/repo1", results[0].Comparison.Repository)
	}
}

func TestCheckDriftUseCase_Execute_AllRepos_MergeCommitsOnly(t *testing.T) {
	mockClient := port.NewMockGitHubClient()
	mockClient.ListRepositoriesFunc = func(ctx context.Context, filter string, includeArchived bool) ([]entity.Repository, error) {
		return []entity.Repository{
			{FullName: "org/repo1"},
		}, nil
	}
	mockClient.CompareBranchesFunc = func(ctx context.Context, owner, repo, base, head string) (*entity.BranchComparison, error) {
		// Merge commits but no file changes — false positive scenario
		return &entity.BranchComparison{
			Repository:   owner + "/" + repo,
			TotalCommits: 2,
			BehindBy:     2,
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

	driftDetector := service.NewDriftDetector()
	uc := NewCheckDriftUseCase(mockClient, cfg, driftDetector)

	results, err := uc.Execute(context.Background(), CheckDriftInput{})

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// No repos should have drift (merge commits only, no file changes)
	if len(results) != 0 {
		t.Fatalf("Execute() returned %d results, want 0 (merge commits only)", len(results))
	}
}

func TestCheckDriftUseCase_Execute_CustomBranchConfig(t *testing.T) {
	mockClient := port.NewMockGitHubClient()
	mockClient.CompareBranchesFunc = func(ctx context.Context, owner, repo, base, head string) (*entity.BranchComparison, error) {
		return &entity.BranchComparison{
			Repository:   owner + "/" + repo,
			ProdBranch:   base,
			DevBranch:    head,
			TotalCommits: 1,
			Files:        []entity.ChangedFile{{Filename: "test.go"}},
		}, nil
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

	driftDetector := service.NewDriftDetector()
	uc := NewCheckDriftUseCase(mockClient, cfg, driftDetector)

	_, err := uc.Execute(context.Background(), CheckDriftInput{
		Repository: "custom/repo",
	})

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	call := mockClient.CompareBranchesCalls[0]
	if call.Base != "production" || call.Head != "staging" {
		t.Errorf("CompareBranches called with base=%s head=%s, want base=production head=staging", call.Base, call.Head)
	}
}

func TestCheckDriftUseCase_Execute_InvalidRepoFormat(t *testing.T) {
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

	driftDetector := service.NewDriftDetector()
	uc := NewCheckDriftUseCase(mockClient, cfg, driftDetector)

	_, err := uc.Execute(context.Background(), CheckDriftInput{
		Repository: "invalid-format",
	})

	// Should now return error for invalid format
	if err == nil {
		t.Fatal("Execute() error = nil, want error for invalid repo format")
	}

	if !strings.Contains(err.Error(), "invalid repository format") {
		t.Errorf("error = %v, want to contain 'invalid repository format'", err)
	}
}

func TestCheckDriftUseCase_Execute_CompareBranchesError(t *testing.T) {
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

	driftDetector := service.NewDriftDetector()
	uc := NewCheckDriftUseCase(mockClient, cfg, driftDetector)

	_, err := uc.Execute(context.Background(), CheckDriftInput{
		Repository: "test/repo",
	})

	if err == nil {
		t.Error("Execute() error = nil, want error")
	}
}
