package service

import "github.com/carlos/mcp-repo-monitor/internal/domain/entity"

type RollbackService struct{}

func NewRollbackService() *RollbackService {
	return &RollbackService{}
}

func (r *RollbackService) ValidateRequest(req entity.RollbackRequest) error {
	return nil
}

func (r *RollbackService) DetermineStrategy(runs []entity.WorkflowRun) entity.RollbackStrategy {
	if len(runs) == 0 {
		return entity.RollbackRevert
	}

	lastRun := runs[0]
	if lastRun.Conclusion == "failure" {
		return entity.RollbackRerun
	}

	return entity.RollbackRevert
}

func (r *RollbackService) GetRecommendedActions(status entity.DriftStatus) []string {
	switch status {
	case entity.DriftNone:
		return []string{"No action needed - branches are synced"}
	case entity.DriftProdAhead:
		return []string{
			"Create sync PR to merge prod into dev",
			"Cherry-pick specific commits to dev",
		}
	case entity.DriftDevAhead:
		return []string{
			"Create PR to merge dev into prod",
			"Review and merge pending PRs",
		}
	case entity.DriftDiverged:
		return []string{
			"Review diverged commits carefully",
			"Consider rebasing dev on prod",
			"Create sync PR with manual conflict resolution",
		}
	default:
		return []string{"Unknown drift status"}
	}
}
