package usecase

import (
	"context"

	"github.com/carlos/mcp-repo-monitor/internal/application/port"
	"github.com/carlos/mcp-repo-monitor/internal/domain/entity"
)

type ListStatusUseCase struct {
	client port.GitHubClient
}

func NewListStatusUseCase(client port.GitHubClient) *ListStatusUseCase {
	return &ListStatusUseCase{client: client}
}

type ListStatusInput struct {
	Filter   string
	Archived bool
}

func (uc *ListStatusUseCase) Execute(ctx context.Context, input ListStatusInput) ([]entity.RepositoryStatus, error) {
	repos, err := uc.client.ListRepositories(ctx, input.Filter, input.Archived)
	if err != nil {
		return nil, err
	}

	var statuses []entity.RepositoryStatus
	for _, repo := range repos {
		status := entity.RepositoryStatus{
			Repository:   repo,
			LastCommitAt: repo.PushedAt,
		}

		prs, _ := uc.client.ListPullRequests(ctx, entity.PRFilter{
			Repository: repo.FullName,
			State:      "open",
			Limit:      100,
		})
		status.OpenPRs = len(prs)

		runs, _ := uc.client.ListWorkflowRuns(ctx, entity.CIFilter{
			Repository: repo.FullName,
			Limit:      1,
		})
		if len(runs) > 0 && runs[0].Conclusion == "failure" {
			status.FailedCI = true
		}

		statuses = append(statuses, status)
	}

	return statuses, nil
}
