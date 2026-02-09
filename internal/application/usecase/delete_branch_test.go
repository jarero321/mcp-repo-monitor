package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/carlos/mcp-repo-monitor/internal/application/port"
)

func TestDeleteBranchUseCase_Execute_Success(t *testing.T) {
	mockClient := port.NewMockGitHubClient()
	uc := NewDeleteBranchUseCase(mockClient)

	result, err := uc.Execute(context.Background(), DeleteBranchInput{
		Repository: "test/repo",
		Branch:     "feature/old-branch",
	})

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !result.Success {
		t.Error("Success = false, want true")
	}

	if !strings.Contains(result.Message, "Deleted branch") {
		t.Errorf("Message = %s, want to contain 'Deleted branch'", result.Message)
	}

	if len(mockClient.DeleteBranchCalls) != 1 {
		t.Fatalf("DeleteBranch called %d times, want 1", len(mockClient.DeleteBranchCalls))
	}

	call := mockClient.DeleteBranchCalls[0]
	if call.Branch != "feature/old-branch" {
		t.Errorf("Branch = %s, want feature/old-branch", call.Branch)
	}
}

func TestDeleteBranchUseCase_Execute_DryRun(t *testing.T) {
	mockClient := port.NewMockGitHubClient()
	uc := NewDeleteBranchUseCase(mockClient)

	result, err := uc.Execute(context.Background(), DeleteBranchInput{
		Repository: "test/repo",
		Branch:     "feature/test",
		DryRun:     true,
	})

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !strings.Contains(result.Message, "DRY RUN") {
		t.Errorf("Message = %s, want to contain 'DRY RUN'", result.Message)
	}

	if len(mockClient.DeleteBranchCalls) != 0 {
		t.Errorf("DeleteBranch called %d times, want 0 (dry run)", len(mockClient.DeleteBranchCalls))
	}
}

func TestDeleteBranchUseCase_Execute_InvalidRepoFormat(t *testing.T) {
	mockClient := port.NewMockGitHubClient()
	uc := NewDeleteBranchUseCase(mockClient)

	_, err := uc.Execute(context.Background(), DeleteBranchInput{
		Repository: "invalid",
		Branch:     "feature/test",
	})

	if err == nil {
		t.Fatal("Execute() error = nil, want error")
	}
}

func TestDeleteBranchUseCase_Execute_EmptyBranch(t *testing.T) {
	mockClient := port.NewMockGitHubClient()
	uc := NewDeleteBranchUseCase(mockClient)

	_, err := uc.Execute(context.Background(), DeleteBranchInput{
		Repository: "test/repo",
		Branch:     "",
	})

	if err == nil {
		t.Fatal("Execute() error = nil, want error for empty branch")
	}

	if !strings.Contains(err.Error(), "branch name is required") {
		t.Errorf("error = %v, want to contain 'branch name is required'", err)
	}
}

func TestDeleteBranchUseCase_Execute_ProtectedBranches(t *testing.T) {
	protected := []string{"main", "master", "develop", "production"}

	for _, branch := range protected {
		t.Run("protected_"+branch, func(t *testing.T) {
			mockClient := port.NewMockGitHubClient()
			uc := NewDeleteBranchUseCase(mockClient)

			_, err := uc.Execute(context.Background(), DeleteBranchInput{
				Repository: "test/repo",
				Branch:     branch,
			})

			if err == nil {
				t.Fatalf("Execute() error = nil, want error for protected branch '%s'", branch)
			}

			if !strings.Contains(err.Error(), "protected branch") {
				t.Errorf("error = %v, want to contain 'protected branch'", err)
			}

			if len(mockClient.DeleteBranchCalls) != 0 {
				t.Errorf("DeleteBranch called for protected branch '%s'", branch)
			}
		})
	}
}

func TestDeleteBranchUseCase_Execute_APIError(t *testing.T) {
	mockClient := port.NewMockGitHubClient()
	mockClient.DeleteBranchFunc = func(ctx context.Context, owner, repo, branch string) error {
		return errors.New("API error: not found")
	}

	uc := NewDeleteBranchUseCase(mockClient)

	_, err := uc.Execute(context.Background(), DeleteBranchInput{
		Repository: "test/repo",
		Branch:     "feature/test",
	})

	if err == nil {
		t.Fatal("Execute() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "failed to delete branch") {
		t.Errorf("error = %v, want to contain 'failed to delete branch'", err)
	}
}
