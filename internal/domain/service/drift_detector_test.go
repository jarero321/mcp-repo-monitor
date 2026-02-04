package service

import (
	"testing"

	"github.com/carlos/mcp-repo-monitor/internal/domain/entity"
)

func TestDriftDetector_AnalyzeDrift(t *testing.T) {
	detector := NewDriftDetector()

	tests := []struct {
		name       string
		comparison entity.BranchComparison
		want       entity.DriftStatus
	}{
		{
			name: "no drift when branches are synced",
			comparison: entity.BranchComparison{
				AheadBy:  0,
				BehindBy: 0,
			},
			want: entity.DriftNone,
		},
		{
			name: "prod ahead when ahead > 0 and behind = 0",
			comparison: entity.BranchComparison{
				AheadBy:  5,
				BehindBy: 0,
			},
			want: entity.DriftProdAhead,
		},
		{
			name: "dev ahead when ahead = 0 and behind > 0",
			comparison: entity.BranchComparison{
				AheadBy:  0,
				BehindBy: 3,
			},
			want: entity.DriftDevAhead,
		},
		{
			name: "diverged when both ahead and behind > 0",
			comparison: entity.BranchComparison{
				AheadBy:  2,
				BehindBy: 4,
			},
			want: entity.DriftDiverged,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detector.AnalyzeDrift(tt.comparison)
			if got != tt.want {
				t.Errorf("AnalyzeDrift() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDriftDetector_HasSignificantDrift(t *testing.T) {
	detector := NewDriftDetector()

	tests := []struct {
		name       string
		comparison entity.BranchComparison
		want       bool
	}{
		{
			name: "no drift when no commits or files",
			comparison: entity.BranchComparison{
				TotalCommits: 0,
				Files:        nil,
			},
			want: false,
		},
		{
			name: "has drift when commits > 0",
			comparison: entity.BranchComparison{
				TotalCommits: 5,
				Files:        nil,
			},
			want: true,
		},
		{
			name: "has drift when files changed",
			comparison: entity.BranchComparison{
				TotalCommits: 0,
				Files:        []entity.ChangedFile{{Filename: "test.go"}},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detector.HasSignificantDrift(tt.comparison)
			if got != tt.want {
				t.Errorf("HasSignificantDrift() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDriftDetector_GetDriftSeverity(t *testing.T) {
	detector := NewDriftDetector()

	tests := []struct {
		name       string
		comparison entity.BranchComparison
		want       string
	}{
		{
			name: "none when no changes",
			comparison: entity.BranchComparison{
				AheadBy:  0,
				BehindBy: 0,
			},
			want: "none",
		},
		{
			name: "low when total <= 5",
			comparison: entity.BranchComparison{
				AheadBy:  3,
				BehindBy: 2,
			},
			want: "low",
		},
		{
			name: "medium when total <= 20",
			comparison: entity.BranchComparison{
				AheadBy:  10,
				BehindBy: 5,
			},
			want: "medium",
		},
		{
			name: "high when total > 20",
			comparison: entity.BranchComparison{
				AheadBy:  15,
				BehindBy: 10,
			},
			want: "high",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detector.GetDriftSeverity(tt.comparison)
			if got != tt.want {
				t.Errorf("GetDriftSeverity() = %v, want %v", got, tt.want)
			}
		})
	}
}
