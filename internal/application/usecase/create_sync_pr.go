package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/carlos/mcp-repo-monitor/config"
	"github.com/carlos/mcp-repo-monitor/internal/application/port"
	"github.com/carlos/mcp-repo-monitor/internal/domain/entity"
)

type CreateSyncPRUseCase struct {
	client port.GitHubClient
	config *config.Config
}

func NewCreateSyncPRUseCase(client port.GitHubClient, cfg *config.Config) *CreateSyncPRUseCase {
	return &CreateSyncPRUseCase{
		client: client,
		config: cfg,
	}
}

type CreateSyncPRInput struct {
	Repository string
	Title      string
	Body       string
	DryRun     bool
}

func (uc *CreateSyncPRUseCase) Execute(ctx context.Context, input CreateSyncPRInput) (*entity.SyncPRResult, error) {
	parts := strings.Split(input.Repository, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository format, expected owner/repo")
	}
	owner, repo := parts[0], parts[1]

	branchConfig := uc.config.GetBranchConfig(input.Repository)

	comparison, err := uc.client.CompareBranches(ctx, owner, repo, branchConfig.ProdBranch, branchConfig.DevBranch)
	if err != nil {
		return nil, fmt.Errorf("failed to compare branches: %w", err)
	}

	if comparison.TotalCommits == 0 {
		return &entity.SyncPRResult{
			Success:  true,
			Message:  "Branches are already in sync, no PR needed",
			Commits:  0,
		}, nil
	}

	title := input.Title
	if title == "" {
		title = fmt.Sprintf("sync: merge %s into %s", branchConfig.ProdBranch, branchConfig.DevBranch)
	}

	body := input.Body
	if body == "" {
		body = fmt.Sprintf("## Sync PR\n\nThis PR syncs `%s` into `%s`.\n\n### Changes\n- **Commits**: %d\n- **Files changed**: %d\n\n---\n_Created by mcp-repo-monitor_",
			branchConfig.ProdBranch,
			branchConfig.DevBranch,
			comparison.TotalCommits,
			len(comparison.Files),
		)
	}

	if input.DryRun {
		return &entity.SyncPRResult{
			Success:      true,
			Message:      fmt.Sprintf("[DRY RUN] Would create PR: %s -> %s", branchConfig.ProdBranch, branchConfig.DevBranch),
			FilesChanged: len(comparison.Files),
			Commits:      comparison.TotalCommits,
		}, nil
	}

	pr, err := uc.client.CreatePullRequest(ctx, owner, repo, title, body, branchConfig.ProdBranch, branchConfig.DevBranch, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create PR: %w", err)
	}

	return &entity.SyncPRResult{
		Success:      true,
		PRURL:        pr.HTMLURL,
		PRNumber:     pr.Number,
		Message:      fmt.Sprintf("Created sync PR #%d", pr.Number),
		FilesChanged: len(comparison.Files),
		Commits:      comparison.TotalCommits,
	}, nil
}
