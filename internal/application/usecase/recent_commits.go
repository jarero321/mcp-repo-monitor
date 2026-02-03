package usecase

import (
	"context"
	"time"

	"github.com/carlos/mcp-repo-monitor/internal/application/port"
	"github.com/carlos/mcp-repo-monitor/internal/domain/entity"
)

type RecentCommitsUseCase struct {
	client port.GitHubClient
}

func NewRecentCommitsUseCase(client port.GitHubClient) *RecentCommitsUseCase {
	return &RecentCommitsUseCase{client: client}
}

type RecentCommitsInput struct {
	Repository string
	Branch     string
	Since      string
	Limit      int
}

func (uc *RecentCommitsUseCase) Execute(ctx context.Context, input RecentCommitsInput) ([]entity.Commit, error) {
	filter := entity.CommitFilter{
		Repository: input.Repository,
		Branch:     input.Branch,
		Limit:      input.Limit,
	}

	if filter.Limit == 0 {
		filter.Limit = 30
	}

	if input.Since != "" {
		t, err := time.Parse(time.RFC3339, input.Since)
		if err == nil {
			filter.Since = &t
		} else {
			duration, err := time.ParseDuration(input.Since)
			if err == nil {
				t := time.Now().Add(-duration)
				filter.Since = &t
			}
		}
	}

	return uc.client.ListCommits(ctx, filter)
}
