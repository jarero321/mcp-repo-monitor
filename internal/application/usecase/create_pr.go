package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/carlos/mcp-repo-monitor/internal/application/port"
	"github.com/carlos/mcp-repo-monitor/internal/domain/entity"
)

type CreatePRUseCase struct {
	client port.GitHubClient
}

func NewCreatePRUseCase(client port.GitHubClient) *CreatePRUseCase {
	return &CreatePRUseCase{client: client}
}

type CreatePRInput struct {
	Repository string
	Title      string
	Head       string
	Base       string
	Body       string
	Draft      bool
	DryRun     bool
}

type CreatePRResult struct {
	Success      bool
	Message      string
	PR           *entity.PullRequest
	FilesChanged int
	Commits      int
}

func (uc *CreatePRUseCase) Execute(ctx context.Context, input CreatePRInput) (*CreatePRResult, error) {
	parts := strings.Split(input.Repository, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository format, expected owner/repo")
	}
	owner, repo := parts[0], parts[1]

	if input.Title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if input.Head == "" {
		return nil, fmt.Errorf("head branch is required")
	}
	if input.Base == "" {
		return nil, fmt.Errorf("base branch is required")
	}

	if input.DryRun {
		comparison, err := uc.client.CompareBranches(ctx, owner, repo, input.Base, input.Head)
		if err != nil {
			return nil, fmt.Errorf("failed to compare branches: %w", err)
		}

		draftLabel := ""
		if input.Draft {
			draftLabel = " (draft)"
		}

		return &CreatePRResult{
			Success:      true,
			Message:      fmt.Sprintf("[DRY RUN] Would create PR%s: %s → %s", draftLabel, input.Head, input.Base),
			FilesChanged: len(comparison.Files),
			Commits:      comparison.TotalCommits,
		}, nil
	}

	pr, err := uc.client.CreatePullRequest(ctx, owner, repo, input.Title, input.Body, input.Head, input.Base, input.Draft)
	if err != nil {
		return nil, fmt.Errorf("failed to create PR: %w", err)
	}

	return &CreatePRResult{
		Success: true,
		Message: fmt.Sprintf("Created PR #%d", pr.Number),
		PR:      pr,
	}, nil
}
