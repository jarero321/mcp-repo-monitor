package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/carlos/mcp-repo-monitor/internal/application/port"
	"github.com/carlos/mcp-repo-monitor/internal/domain/entity"
	"github.com/carlos/mcp-repo-monitor/internal/domain/service"
)

type TriggerRollbackUseCase struct {
	client          port.GitHubClient
	rollbackService *service.RollbackService
}

func NewTriggerRollbackUseCase(client port.GitHubClient, rollbackService *service.RollbackService) *TriggerRollbackUseCase {
	return &TriggerRollbackUseCase{
		client:          client,
		rollbackService: rollbackService,
	}
}

type TriggerRollbackInput struct {
	Repository string
	Strategy   string
	WorkflowID int64
	RunID      int64
	CommitSHA  string
	DryRun     bool
}

func (uc *TriggerRollbackUseCase) Execute(ctx context.Context, input TriggerRollbackInput) (*entity.RollbackResult, error) {
	parts := strings.Split(input.Repository, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository format, expected owner/repo")
	}
	owner, repo := parts[0], parts[1]

	strategy := entity.RollbackStrategy(input.Strategy)

	if input.DryRun {
		return &entity.RollbackResult{
			Success:  true,
			Strategy: strategy,
			Message:  fmt.Sprintf("[DRY RUN] Would execute %s rollback on %s", strategy, input.Repository),
		}, nil
	}

	switch strategy {
	case entity.RollbackRerun:
		if input.RunID == 0 {
			runs, err := uc.client.ListWorkflowRuns(ctx, entity.CIFilter{
				Repository: input.Repository,
				Limit:      1,
			})
			if err != nil {
				return nil, err
			}
			if len(runs) == 0 {
				return nil, fmt.Errorf("no workflow runs found")
			}
			input.RunID = runs[0].ID
		}

		err := uc.client.RerunWorkflow(ctx, owner, repo, input.RunID)
		if err != nil {
			return nil, err
		}

		return &entity.RollbackResult{
			Success:  true,
			Strategy: strategy,
			Message:  fmt.Sprintf("Rerun triggered for run ID %d", input.RunID),
			RunURL:   fmt.Sprintf("https://github.com/%s/actions/runs/%d", input.Repository, input.RunID),
		}, nil

	case entity.RollbackWorkflow:
		if input.WorkflowID == 0 {
			return nil, fmt.Errorf("workflow_id required for workflow strategy")
		}

		repoInfo, err := uc.client.GetRepository(ctx, owner, repo)
		if err != nil {
			return nil, err
		}

		workflowFile := fmt.Sprintf("%d", input.WorkflowID)
		err = uc.client.TriggerWorkflow(ctx, owner, repo, workflowFile, repoInfo.DefaultBranch)
		if err != nil {
			return nil, err
		}

		return &entity.RollbackResult{
			Success:  true,
			Strategy: strategy,
			Message:  fmt.Sprintf("Workflow %d triggered on %s", input.WorkflowID, repoInfo.DefaultBranch),
		}, nil

	case entity.RollbackRevert:
		return &entity.RollbackResult{
			Success:  false,
			Strategy: strategy,
			Message:  "Revert strategy requires manual intervention - create a revert commit through the GitHub UI or CLI",
		}, nil

	default:
		return nil, fmt.Errorf("unknown rollback strategy: %s", strategy)
	}
}
