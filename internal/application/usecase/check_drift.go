package usecase

import (
	"context"
	"strings"

	"github.com/carlos/mcp-repo-monitor/config"
	"github.com/carlos/mcp-repo-monitor/internal/application/port"
	"github.com/carlos/mcp-repo-monitor/internal/domain/entity"
	"github.com/carlos/mcp-repo-monitor/internal/domain/service"
)

type CheckDriftUseCase struct {
	client        port.GitHubClient
	config        *config.Config
	driftDetector *service.DriftDetector
}

func NewCheckDriftUseCase(client port.GitHubClient, cfg *config.Config, driftDetector *service.DriftDetector) *CheckDriftUseCase {
	return &CheckDriftUseCase{
		client:        client,
		config:        cfg,
		driftDetector: driftDetector,
	}
}

type CheckDriftInput struct {
	Repository string
}

type DriftResult struct {
	Comparison entity.BranchComparison
	Severity   string
	Actions    []string
}

func (uc *CheckDriftUseCase) Execute(ctx context.Context, input CheckDriftInput) ([]DriftResult, error) {
	var results []DriftResult

	if input.Repository != "" {
		result, err := uc.checkSingleRepo(ctx, input.Repository)
		if err != nil {
			return nil, err
		}
		results = append(results, *result)
	} else {
		repos, err := uc.client.ListRepositories(ctx, "", false)
		if err != nil {
			return nil, err
		}

		for _, repo := range repos {
			result, err := uc.checkSingleRepo(ctx, repo.FullName)
			if err != nil {
				continue
			}
			if result.Comparison.TotalCommits > 0 {
				results = append(results, *result)
			}
		}
	}

	return results, nil
}

func (uc *CheckDriftUseCase) checkSingleRepo(ctx context.Context, repoFullName string) (*DriftResult, error) {
	parts := strings.Split(repoFullName, "/")
	if len(parts) != 2 {
		return nil, nil
	}
	owner, repo := parts[0], parts[1]

	branchConfig := uc.config.GetBranchConfig(repoFullName)

	comparison, err := uc.client.CompareBranches(ctx, owner, repo, branchConfig.ProdBranch, branchConfig.DevBranch)
	if err != nil {
		return nil, err
	}

	comparison.Status = uc.driftDetector.AnalyzeDrift(*comparison)

	rollbackSvc := service.NewRollbackService()

	return &DriftResult{
		Comparison: *comparison,
		Severity:   uc.driftDetector.GetDriftSeverity(*comparison),
		Actions:    rollbackSvc.GetRecommendedActions(comparison.Status),
	}, nil
}
