package usecase

import (
	"context"

	"github.com/carlos/mcp-repo-monitor/internal/application/port"
	"github.com/carlos/mcp-repo-monitor/internal/domain/entity"
)

type ListPRsUseCase struct {
	client port.GitHubClient
}

func NewListPRsUseCase(client port.GitHubClient) *ListPRsUseCase {
	return &ListPRsUseCase{client: client}
}

type ListPRsInput struct {
	Repository string
	State      string
	Limit      int
}

func (uc *ListPRsUseCase) Execute(ctx context.Context, input ListPRsInput) ([]entity.PullRequest, error) {
	filter := entity.PRFilter{
		Repository: input.Repository,
		State:      input.State,
		Limit:      input.Limit,
	}

	if filter.State == "" {
		filter.State = "open"
	}
	if filter.Limit == 0 {
		filter.Limit = 30
	}

	return uc.client.ListPullRequests(ctx, filter)
}
