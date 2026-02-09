package service

import "github.com/carlos/mcp-repo-monitor/internal/domain/entity"

type DriftDetector struct{}

func NewDriftDetector() *DriftDetector {
	return &DriftDetector{}
}

func (d *DriftDetector) AnalyzeDrift(comparison entity.BranchComparison) entity.DriftStatus {
	// If GitHub reports branches as identical or there are no file changes,
	// the branches are effectively synced (merge commits don't count as real drift)
	if comparison.GitHubStatus == "identical" || len(comparison.Files) == 0 {
		return entity.DriftNone
	}

	if comparison.AheadBy > 0 && comparison.BehindBy > 0 {
		return entity.DriftDiverged
	}

	if comparison.AheadBy > 0 {
		return entity.DriftProdAhead
	}

	if comparison.BehindBy > 0 {
		return entity.DriftDevAhead
	}

	return entity.DriftNone
}

func (d *DriftDetector) HasSignificantDrift(comparison entity.BranchComparison) bool {
	return len(comparison.Files) > 0
}

func (d *DriftDetector) GetDriftSeverity(comparison entity.BranchComparison) string {
	if len(comparison.Files) == 0 {
		return "none"
	}

	totalChanges := comparison.AheadBy + comparison.BehindBy

	if totalChanges == 0 {
		return "none"
	}
	if totalChanges <= 5 {
		return "low"
	}
	if totalChanges <= 20 {
		return "medium"
	}
	return "high"
}
