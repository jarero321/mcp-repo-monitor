package usecase

import (
	"context"

	"github.com/carlos/mcp-repo-monitor/internal/application/port"
	"github.com/carlos/mcp-repo-monitor/internal/domain/entity"
)

type CheckCIUseCase struct {
	client port.GitHubClient
}

func NewCheckCIUseCase(client port.GitHubClient) *CheckCIUseCase {
	return &CheckCIUseCase{client: client}
}

type CheckCIInput struct {
	Repository string
	Branch     string
	Workflow   string
	Limit      int
}

func (uc *CheckCIUseCase) Execute(ctx context.Context, input CheckCIInput) ([]entity.WorkflowRun, error) {
	filter := entity.CIFilter{
		Repository: input.Repository,
		Branch:     input.Branch,
		Workflow:   input.Workflow,
		Limit:      input.Limit,
	}

	if filter.Limit == 0 {
		filter.Limit = 10
	}

	return uc.client.ListWorkflowRuns(ctx, filter)
}
