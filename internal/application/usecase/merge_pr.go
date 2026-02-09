package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/carlos/mcp-repo-monitor/internal/application/port"
	"github.com/carlos/mcp-repo-monitor/internal/domain/entity"
)

type MergePRUseCase struct {
	client port.GitHubClient
}

func NewMergePRUseCase(client port.GitHubClient) *MergePRUseCase {
	return &MergePRUseCase{client: client}
}

type MergePRInput struct {
	Repository   string
	PRNumber     int
	Method       string
	CommitTitle  string
	DeleteBranch bool
	DryRun       bool
}

func (uc *MergePRUseCase) Execute(ctx context.Context, input MergePRInput) (*entity.MergeResult, error) {
	parts := strings.Split(input.Repository, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository format, expected owner/repo")
	}
	owner, repo := parts[0], parts[1]

	if input.PRNumber <= 0 {
		return nil, fmt.Errorf("pr_number is required and must be positive")
	}

	method := input.Method
	if method == "" {
		method = "merge"
	}

	switch entity.MergeMethod(method) {
	case entity.MergeMethodMerge, entity.MergeMethodSquash, entity.MergeMethodRebase:
		// valid
	default:
		return nil, fmt.Errorf("invalid merge method '%s', must be merge, squash, or rebase", method)
	}

	// Get PR to verify state
	pr, err := uc.client.GetPullRequest(ctx, owner, repo, input.PRNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get PR #%d: %w", input.PRNumber, err)
	}

	if pr.State != "open" {
		return nil, fmt.Errorf("PR #%d is not open (state: %s)", input.PRNumber, pr.State)
	}

	if pr.Mergeable != nil && !*pr.Mergeable {
		return nil, fmt.Errorf("PR #%d has merge conflicts, resolve them before merging", input.PRNumber)
	}

	if input.DryRun {
		msg := fmt.Sprintf("[DRY RUN] Would merge PR #%d using %s method", input.PRNumber, method)
		if input.DeleteBranch {
			msg += fmt.Sprintf(" and delete branch '%s'", pr.HeadBranch)
		}
		return &entity.MergeResult{
			Success:     true,
			Message:     msg,
			PRURL:       pr.HTMLURL,
			PRNumber:    input.PRNumber,
			MergeMethod: entity.MergeMethod(method),
			BranchName:  pr.HeadBranch,
		}, nil
	}

	result, err := uc.client.MergePullRequest(ctx, owner, repo, input.PRNumber, method, input.CommitTitle)
	if err != nil {
		return nil, fmt.Errorf("failed to merge PR #%d: %w", input.PRNumber, err)
	}

	result.BranchName = pr.HeadBranch

	if input.DeleteBranch && pr.HeadBranch != "" {
		err := uc.client.DeleteBranch(ctx, owner, repo, pr.HeadBranch)
		if err != nil {
			// Non-fatal: merge succeeded but branch deletion failed
			result.Message += fmt.Sprintf(" (warning: failed to delete branch '%s': %v)", pr.HeadBranch, err)
		} else {
			result.BranchDeleted = true
		}
	}

	return result, nil
}
