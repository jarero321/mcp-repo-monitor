package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/carlos/mcp-repo-monitor/internal/application/port"
)

type DeleteBranchUseCase struct {
	client port.GitHubClient
}

func NewDeleteBranchUseCase(client port.GitHubClient) *DeleteBranchUseCase {
	return &DeleteBranchUseCase{client: client}
}

type DeleteBranchInput struct {
	Repository string
	Branch     string
	DryRun     bool
}

type DeleteBranchResult struct {
	Success bool
	Message string
}

var protectedBranches = map[string]bool{
	"main":       true,
	"master":     true,
	"develop":    true,
	"production": true,
}

func (uc *DeleteBranchUseCase) Execute(ctx context.Context, input DeleteBranchInput) (*DeleteBranchResult, error) {
	parts := strings.Split(input.Repository, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository format, expected owner/repo")
	}
	owner, repo := parts[0], parts[1]

	if input.Branch == "" {
		return nil, fmt.Errorf("branch name is required")
	}

	if protectedBranches[input.Branch] {
		return nil, fmt.Errorf("refusing to delete protected branch '%s'", input.Branch)
	}

	if input.DryRun {
		return &DeleteBranchResult{
			Success: true,
			Message: fmt.Sprintf("[DRY RUN] Would delete branch '%s' from %s", input.Branch, input.Repository),
		}, nil
	}

	err := uc.client.DeleteBranch(ctx, owner, repo, input.Branch)
	if err != nil {
		return nil, fmt.Errorf("failed to delete branch '%s': %w", input.Branch, err)
	}

	return &DeleteBranchResult{
		Success: true,
		Message: fmt.Sprintf("Deleted branch '%s' from %s", input.Branch, input.Repository),
	}, nil
}
