package service

import (
	"reflect"
	"testing"

	"github.com/carlos/mcp-repo-monitor/internal/domain/entity"
)

func TestRollbackService_DetermineStrategy(t *testing.T) {
	service := NewRollbackService()

	tests := []struct {
		name string
		runs []entity.WorkflowRun
		want entity.RollbackStrategy
	}{
		{
			name: "revert when no runs",
			runs: []entity.WorkflowRun{},
			want: entity.RollbackRevert,
		},
		{
			name: "rerun when last run failed",
			runs: []entity.WorkflowRun{
				{Conclusion: "failure"},
			},
			want: entity.RollbackRerun,
		},
		{
			name: "revert when last run succeeded",
			runs: []entity.WorkflowRun{
				{Conclusion: "success"},
			},
			want: entity.RollbackRevert,
		},
		{
			name: "revert when last run cancelled",
			runs: []entity.WorkflowRun{
				{Conclusion: "cancelled"},
			},
			want: entity.RollbackRevert,
		},
		{
			name: "uses first run in list (most recent)",
			runs: []entity.WorkflowRun{
				{Conclusion: "failure"},
				{Conclusion: "success"},
			},
			want: entity.RollbackRerun,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.DetermineStrategy(tt.runs)
			if got != tt.want {
				t.Errorf("DetermineStrategy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRollbackService_GetRecommendedActions(t *testing.T) {
	service := NewRollbackService()

	tests := []struct {
		name   string
		status entity.DriftStatus
		want   []string
	}{
		{
			name:   "no action for synced branches",
			status: entity.DriftNone,
			want:   []string{"No action needed - branches are synced"},
		},
		{
			name:   "sync options for prod ahead",
			status: entity.DriftProdAhead,
			want: []string{
				"Create sync PR to merge prod into dev",
				"Cherry-pick specific commits to dev",
			},
		},
		{
			name:   "merge options for dev ahead",
			status: entity.DriftDevAhead,
			want: []string{
				"Create PR to merge dev into prod",
				"Review and merge pending PRs",
			},
		},
		{
			name:   "rebase options for diverged",
			status: entity.DriftDiverged,
			want: []string{
				"Review diverged commits carefully",
				"Consider rebasing dev on prod",
				"Create sync PR with manual conflict resolution",
			},
		},
		{
			name:   "unknown status",
			status: entity.DriftStatus("unknown"),
			want:   []string{"Unknown drift status"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.GetRecommendedActions(tt.status)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRecommendedActions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRollbackService_ValidateRequest(t *testing.T) {
	service := NewRollbackService()

	// Current implementation always returns nil
	err := service.ValidateRequest(entity.RollbackRequest{})
	if err != nil {
		t.Errorf("ValidateRequest() error = %v, want nil", err)
	}
}
