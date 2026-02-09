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
			name: "prod ahead when ahead > 0 and behind = 0 with file changes",
			comparison: entity.BranchComparison{
				AheadBy:  5,
				BehindBy: 0,
				Files:    []entity.ChangedFile{{Filename: "test.go"}},
			},
			want: entity.DriftProdAhead,
		},
		{
			name: "dev ahead when ahead = 0 and behind > 0 with file changes",
			comparison: entity.BranchComparison{
				AheadBy:  0,
				BehindBy: 3,
				Files:    []entity.ChangedFile{{Filename: "test.go"}},
			},
			want: entity.DriftDevAhead,
		},
		{
			name: "diverged when both ahead and behind > 0 with file changes",
			comparison: entity.BranchComparison{
				AheadBy:  2,
				BehindBy: 4,
				Files:    []entity.ChangedFile{{Filename: "test.go"}},
			},
			want: entity.DriftDiverged,
		},
		{
			name: "no drift when GitHub reports identical despite merge commits",
			comparison: entity.BranchComparison{
				AheadBy:      0,
				BehindBy:     1,
				TotalCommits: 1,
				GitHubStatus: "identical",
			},
			want: entity.DriftNone,
		},
		{
			name: "no drift when files empty despite commits ahead",
			comparison: entity.BranchComparison{
				AheadBy:      3,
				BehindBy:     0,
				TotalCommits: 3,
				Files:        []entity.ChangedFile{},
			},
			want: entity.DriftNone,
		},
		{
			name: "no drift when files nil despite commits ahead (merge commit scenario)",
			comparison: entity.BranchComparison{
				AheadBy:      0,
				BehindBy:     2,
				TotalCommits: 2,
				GitHubStatus: "behind",
				Files:        nil,
			},
			want: entity.DriftNone,
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
			name: "no drift when no files changed",
			comparison: entity.BranchComparison{
				TotalCommits: 0,
				Files:        nil,
			},
			want: false,
		},
		{
			name: "no drift when only merge commits (no file changes)",
			comparison: entity.BranchComparison{
				TotalCommits: 5,
				Files:        nil,
			},
			want: false,
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
			name: "none when files empty despite commits",
			comparison: entity.BranchComparison{
				AheadBy:      5,
				BehindBy:     3,
				TotalCommits: 8,
				Files:        nil,
			},
			want: "none",
		},
		{
			name: "low when total <= 5 with file changes",
			comparison: entity.BranchComparison{
				AheadBy:  3,
				BehindBy: 2,
				Files:    []entity.ChangedFile{{Filename: "test.go"}},
			},
			want: "low",
		},
		{
			name: "medium when total <= 20 with file changes",
			comparison: entity.BranchComparison{
				AheadBy:  10,
				BehindBy: 5,
				Files:    []entity.ChangedFile{{Filename: "test.go"}},
			},
			want: "medium",
		},
		{
			name: "high when total > 20 with file changes",
			comparison: entity.BranchComparison{
				AheadBy:  15,
				BehindBy: 10,
				Files:    []entity.ChangedFile{{Filename: "test.go"}},
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
