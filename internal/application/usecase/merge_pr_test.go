package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/carlos/mcp-repo-monitor/internal/application/port"
	"github.com/carlos/mcp-repo-monitor/internal/domain/entity"
)

func boolPtr(b bool) *bool { return &b }

func TestMergePRUseCase_Execute_MergeSuccess(t *testing.T) {
	mockClient := port.NewMockGitHubClient()
	mockClient.GetPullRequestFunc = func(ctx context.Context, owner, repo string, number int) (*entity.PullRequest, error) {
		return &entity.PullRequest{
			Number:     42,
			State:      "open",
			Mergeable:  boolPtr(true),
			HeadBranch: "feature/test",
			HTMLURL:    "https://github.com/test/repo/pull/42",
		}, nil
	}
	mockClient.MergePullRequestFunc = func(ctx context.Context, owner, repo string, number int, method, commitTitle string) (*entity.MergeResult, error) {
		return &entity.MergeResult{
			Success:     true,
			SHA:         "abc123",
			Message:     "Pull Request successfully merged",
			PRURL:       "https://github.com/test/repo/pull/42",
			PRNumber:    42,
			MergeMethod: entity.MergeMethodMerge,
		}, nil
	}

	uc := NewMergePRUseCase(mockClient)

	result, err := uc.Execute(context.Background(), MergePRInput{
		Repository: "test/repo",
		PRNumber:   42,
	})

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !result.Success {
		t.Error("Success = false, want true")
	}

	if result.SHA != "abc123" {
		t.Errorf("SHA = %s, want abc123", result.SHA)
	}

	// Verify merge was called with default method
	if len(mockClient.MergePullRequestCalls) != 1 {
		t.Fatalf("MergePullRequest called %d times, want 1", len(mockClient.MergePullRequestCalls))
	}
	call := mockClient.MergePullRequestCalls[0]
	if call.Method != "merge" {
		t.Errorf("Method = %s, want merge", call.Method)
	}
}

func TestMergePRUseCase_Execute_SquashMethod(t *testing.T) {
	mockClient := port.NewMockGitHubClient()
	mockClient.GetPullRequestFunc = func(ctx context.Context, owner, repo string, number int) (*entity.PullRequest, error) {
		return &entity.PullRequest{Number: 1, State: "open", Mergeable: boolPtr(true), HeadBranch: "feat"}, nil
	}
	mockClient.MergePullRequestFunc = func(ctx context.Context, owner, repo string, number int, method, commitTitle string) (*entity.MergeResult, error) {
		return &entity.MergeResult{Success: true, SHA: "def456", MergeMethod: entity.MergeMethodSquash}, nil
	}

	uc := NewMergePRUseCase(mockClient)

	result, err := uc.Execute(context.Background(), MergePRInput{
		Repository: "test/repo",
		PRNumber:   1,
		Method:     "squash",
	})

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !result.Success {
		t.Error("Success = false, want true")
	}

	call := mockClient.MergePullRequestCalls[0]
	if call.Method != "squash" {
		t.Errorf("Method = %s, want squash", call.Method)
	}
}

func TestMergePRUseCase_Execute_RebaseMethod(t *testing.T) {
	mockClient := port.NewMockGitHubClient()
	mockClient.GetPullRequestFunc = func(ctx context.Context, owner, repo string, number int) (*entity.PullRequest, error) {
		return &entity.PullRequest{Number: 1, State: "open", Mergeable: boolPtr(true), HeadBranch: "feat"}, nil
	}
	mockClient.MergePullRequestFunc = func(ctx context.Context, owner, repo string, number int, method, commitTitle string) (*entity.MergeResult, error) {
		return &entity.MergeResult{Success: true, SHA: "ghi789", MergeMethod: entity.MergeMethodRebase}, nil
	}

	uc := NewMergePRUseCase(mockClient)

	result, err := uc.Execute(context.Background(), MergePRInput{
		Repository: "test/repo",
		PRNumber:   1,
		Method:     "rebase",
	})

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !result.Success {
		t.Error("Success = false, want true")
	}

	call := mockClient.MergePullRequestCalls[0]
	if call.Method != "rebase" {
		t.Errorf("Method = %s, want rebase", call.Method)
	}
}

func TestMergePRUseCase_Execute_WithBranchDeletion(t *testing.T) {
	mockClient := port.NewMockGitHubClient()
	mockClient.GetPullRequestFunc = func(ctx context.Context, owner, repo string, number int) (*entity.PullRequest, error) {
		return &entity.PullRequest{
			Number:     1,
			State:      "open",
			Mergeable:  boolPtr(true),
			HeadBranch: "feature/test",
		}, nil
	}
	mockClient.MergePullRequestFunc = func(ctx context.Context, owner, repo string, number int, method, commitTitle string) (*entity.MergeResult, error) {
		return &entity.MergeResult{Success: true, SHA: "abc123", Message: "Merged"}, nil
	}

	uc := NewMergePRUseCase(mockClient)

	result, err := uc.Execute(context.Background(), MergePRInput{
		Repository:   "test/repo",
		PRNumber:     1,
		DeleteBranch: true,
	})

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !result.BranchDeleted {
		t.Error("BranchDeleted = false, want true")
	}

	if len(mockClient.DeleteBranchCalls) != 1 {
		t.Fatalf("DeleteBranch called %d times, want 1", len(mockClient.DeleteBranchCalls))
	}

	deleteCall := mockClient.DeleteBranchCalls[0]
	if deleteCall.Branch != "feature/test" {
		t.Errorf("Branch = %s, want feature/test", deleteCall.Branch)
	}
}

func TestMergePRUseCase_Execute_BranchDeletionFails_NonFatal(t *testing.T) {
	mockClient := port.NewMockGitHubClient()
	mockClient.GetPullRequestFunc = func(ctx context.Context, owner, repo string, number int) (*entity.PullRequest, error) {
		return &entity.PullRequest{
			Number:     1,
			State:      "open",
			Mergeable:  boolPtr(true),
			HeadBranch: "feature/test",
		}, nil
	}
	mockClient.MergePullRequestFunc = func(ctx context.Context, owner, repo string, number int, method, commitTitle string) (*entity.MergeResult, error) {
		return &entity.MergeResult{Success: true, SHA: "abc123", Message: "Merged"}, nil
	}
	mockClient.DeleteBranchFunc = func(ctx context.Context, owner, repo, branch string) error {
		return errors.New("branch deletion failed")
	}

	uc := NewMergePRUseCase(mockClient)

	result, err := uc.Execute(context.Background(), MergePRInput{
		Repository:   "test/repo",
		PRNumber:     1,
		DeleteBranch: true,
	})

	// Merge should succeed even if branch deletion fails
	if err != nil {
		t.Fatalf("Execute() error = %v, want nil (branch deletion is non-fatal)", err)
	}

	if !result.Success {
		t.Error("Success = false, want true")
	}

	if result.BranchDeleted {
		t.Error("BranchDeleted = true, want false (deletion failed)")
	}

	if !strings.Contains(result.Message, "warning") {
		t.Errorf("Message = %s, want to contain 'warning'", result.Message)
	}
}

func TestMergePRUseCase_Execute_DryRun(t *testing.T) {
	mockClient := port.NewMockGitHubClient()
	mockClient.GetPullRequestFunc = func(ctx context.Context, owner, repo string, number int) (*entity.PullRequest, error) {
		return &entity.PullRequest{
			Number:     42,
			State:      "open",
			Mergeable:  boolPtr(true),
			HeadBranch: "feature/test",
			HTMLURL:    "https://github.com/test/repo/pull/42",
		}, nil
	}

	uc := NewMergePRUseCase(mockClient)

	result, err := uc.Execute(context.Background(), MergePRInput{
		Repository:   "test/repo",
		PRNumber:     42,
		DeleteBranch: true,
		DryRun:       true,
	})

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !strings.Contains(result.Message, "DRY RUN") {
		t.Errorf("Message = %s, want to contain 'DRY RUN'", result.Message)
	}

	if !strings.Contains(result.Message, "feature/test") {
		t.Errorf("Message = %s, want to contain branch name", result.Message)
	}

	// Should not call merge or delete in dry run
	if len(mockClient.MergePullRequestCalls) != 0 {
		t.Errorf("MergePullRequest called %d times, want 0", len(mockClient.MergePullRequestCalls))
	}
	if len(mockClient.DeleteBranchCalls) != 0 {
		t.Errorf("DeleteBranch called %d times, want 0", len(mockClient.DeleteBranchCalls))
	}
}

func TestMergePRUseCase_Execute_PRNotOpen(t *testing.T) {
	mockClient := port.NewMockGitHubClient()
	mockClient.GetPullRequestFunc = func(ctx context.Context, owner, repo string, number int) (*entity.PullRequest, error) {
		return &entity.PullRequest{Number: 1, State: "closed"}, nil
	}

	uc := NewMergePRUseCase(mockClient)

	_, err := uc.Execute(context.Background(), MergePRInput{
		Repository: "test/repo",
		PRNumber:   1,
	})

	if err == nil {
		t.Fatal("Execute() error = nil, want error for closed PR")
	}

	if !strings.Contains(err.Error(), "not open") {
		t.Errorf("error = %v, want to contain 'not open'", err)
	}
}

func TestMergePRUseCase_Execute_PRHasConflicts(t *testing.T) {
	mockClient := port.NewMockGitHubClient()
	mockClient.GetPullRequestFunc = func(ctx context.Context, owner, repo string, number int) (*entity.PullRequest, error) {
		return &entity.PullRequest{Number: 1, State: "open", Mergeable: boolPtr(false)}, nil
	}

	uc := NewMergePRUseCase(mockClient)

	_, err := uc.Execute(context.Background(), MergePRInput{
		Repository: "test/repo",
		PRNumber:   1,
	})

	if err == nil {
		t.Fatal("Execute() error = nil, want error for merge conflicts")
	}

	if !strings.Contains(err.Error(), "conflicts") {
		t.Errorf("error = %v, want to contain 'conflicts'", err)
	}
}

func TestMergePRUseCase_Execute_InvalidRepoFormat(t *testing.T) {
	mockClient := port.NewMockGitHubClient()
	uc := NewMergePRUseCase(mockClient)

	_, err := uc.Execute(context.Background(), MergePRInput{
		Repository: "invalid",
		PRNumber:   1,
	})

	if err == nil {
		t.Fatal("Execute() error = nil, want error")
	}
}

func TestMergePRUseCase_Execute_InvalidMethod(t *testing.T) {
	mockClient := port.NewMockGitHubClient()
	uc := NewMergePRUseCase(mockClient)

	_, err := uc.Execute(context.Background(), MergePRInput{
		Repository: "test/repo",
		PRNumber:   1,
		Method:     "invalid",
	})

	if err == nil {
		t.Fatal("Execute() error = nil, want error for invalid method")
	}

	if !strings.Contains(err.Error(), "invalid merge method") {
		t.Errorf("error = %v, want to contain 'invalid merge method'", err)
	}
}
