package mcp

import (
	"fmt"
	"strings"
	"time"

	"github.com/carlos/mcp-repo-monitor/internal/application/usecase"
	"github.com/carlos/mcp-repo-monitor/internal/domain/entity"
)

type Presenter struct{}

func NewPresenter() *Presenter {
	return &Presenter{}
}

func (p *Presenter) FormatRepositoryStatuses(statuses []entity.RepositoryStatus) string {
	if len(statuses) == 0 {
		return "No repositories found"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("┌─────────────────────────────────────────────────────────────────┐\n"))
	sb.WriteString(fmt.Sprintf("│ REPOSITORY STATUS                                               │\n"))
	sb.WriteString(fmt.Sprintf("├─────────────────────────────────────────────────────────────────┤\n"))

	for i, s := range statuses {
		ciStatus := "✓"
		if s.FailedCI {
			ciStatus = "✗"
		}

		sb.WriteString(fmt.Sprintf("│ %-40s                        │\n", truncate(s.Repository.FullName, 40)))
		sb.WriteString(fmt.Sprintf("│   PRs: %-3d │ CI: %s │ Updated: %-20s      │\n",
			s.OpenPRs,
			ciStatus,
			formatTime(s.LastCommitAt),
		))

		if s.Repository.Description != "" {
			sb.WriteString(fmt.Sprintf("│   %s\n", truncate(s.Repository.Description, 60)))
		}

		if i < len(statuses)-1 {
			sb.WriteString(fmt.Sprintf("├─────────────────────────────────────────────────────────────────┤\n"))
		}
	}

	sb.WriteString(fmt.Sprintf("└─────────────────────────────────────────────────────────────────┘\n"))
	sb.WriteString(fmt.Sprintf("Total: %d repositories\n", len(statuses)))

	return sb.String()
}

func (p *Presenter) FormatPullRequests(prs []entity.PullRequest) string {
	if len(prs) == 0 {
		return "No pull requests found"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("┌─────────────────────────────────────────────────────────────────┐\n"))
	sb.WriteString(fmt.Sprintf("│ PULL REQUESTS                                                   │\n"))
	sb.WriteString(fmt.Sprintf("├─────────────────────────────────────────────────────────────────┤\n"))

	for i, pr := range prs {
		state := "○"
		if pr.State == "open" {
			state = "●"
		}
		if pr.Draft {
			state = "◐"
		}

		sb.WriteString(fmt.Sprintf("│ %s #%-5d %-50s │\n", state, pr.Number, truncate(pr.Title, 50)))
		sb.WriteString(fmt.Sprintf("│   %s → %s │ @%-15s │ +%d/-%d   │\n",
			truncate(pr.HeadBranch, 15),
			truncate(pr.BaseBranch, 15),
			truncate(pr.User, 15),
			pr.Additions,
			pr.Deletions,
		))
		sb.WriteString(fmt.Sprintf("│   %s │ %s\n", pr.Repository, formatTime(pr.UpdatedAt)))

		if len(pr.Labels) > 0 {
			sb.WriteString(fmt.Sprintf("│   Labels: %s\n", strings.Join(pr.Labels, ", ")))
		}

		if i < len(prs)-1 {
			sb.WriteString(fmt.Sprintf("├─────────────────────────────────────────────────────────────────┤\n"))
		}
	}

	sb.WriteString(fmt.Sprintf("└─────────────────────────────────────────────────────────────────┘\n"))
	sb.WriteString(fmt.Sprintf("Total: %d pull requests\n", len(prs)))

	return sb.String()
}

func (p *Presenter) FormatWorkflowRuns(runs []entity.WorkflowRun) string {
	if len(runs) == 0 {
		return "No workflow runs found"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("┌─────────────────────────────────────────────────────────────────┐\n"))
	sb.WriteString(fmt.Sprintf("│ CI/CD STATUS                                                    │\n"))
	sb.WriteString(fmt.Sprintf("├─────────────────────────────────────────────────────────────────┤\n"))

	for i, run := range runs {
		status := getStatusIcon(run.Conclusion)

		sb.WriteString(fmt.Sprintf("│ %s %-50s │\n", status, truncate(run.Name, 50)))
		sb.WriteString(fmt.Sprintf("│   Branch: %-15s │ Run #%-5d │ %s   │\n",
			truncate(run.HeadBranch, 15),
			run.RunNumber,
			formatTime(run.CreatedAt),
		))
		sb.WriteString(fmt.Sprintf("│   Status: %-10s │ Conclusion: %-10s         │\n",
			run.Status,
			run.Conclusion,
		))
		sb.WriteString(fmt.Sprintf("│   %s\n", run.HTMLURL))

		if i < len(runs)-1 {
			sb.WriteString(fmt.Sprintf("├─────────────────────────────────────────────────────────────────┤\n"))
		}
	}

	sb.WriteString(fmt.Sprintf("└─────────────────────────────────────────────────────────────────┘\n"))
	sb.WriteString(fmt.Sprintf("Total: %d runs\n", len(runs)))

	return sb.String()
}

func (p *Presenter) FormatCommits(commits []entity.Commit) string {
	if len(commits) == 0 {
		return "No commits found"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("┌─────────────────────────────────────────────────────────────────┐\n"))
	sb.WriteString(fmt.Sprintf("│ RECENT COMMITS                                                  │\n"))
	sb.WriteString(fmt.Sprintf("├─────────────────────────────────────────────────────────────────┤\n"))

	for i, commit := range commits {
		firstLine := strings.Split(commit.Message, "\n")[0]

		sb.WriteString(fmt.Sprintf("│ %.7s %-54s │\n", commit.SHA, truncate(firstLine, 54)))
		sb.WriteString(fmt.Sprintf("│   @%-20s │ %s               │\n",
			truncate(commit.Author, 20),
			formatTime(commit.Date),
		))

		if commit.Repository != "" {
			sb.WriteString(fmt.Sprintf("│   %s\n", commit.Repository))
		}

		if i < len(commits)-1 {
			sb.WriteString(fmt.Sprintf("├─────────────────────────────────────────────────────────────────┤\n"))
		}
	}

	sb.WriteString(fmt.Sprintf("└─────────────────────────────────────────────────────────────────┘\n"))
	sb.WriteString(fmt.Sprintf("Total: %d commits\n", len(commits)))

	return sb.String()
}

func (p *Presenter) FormatDriftResults(results []usecase.DriftResult) string {
	if len(results) == 0 {
		return "No drift detected - all branches are in sync"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("┌─────────────────────────────────────────────────────────────────┐\n"))
	sb.WriteString(fmt.Sprintf("│ BRANCH DRIFT ANALYSIS                                           │\n"))
	sb.WriteString(fmt.Sprintf("├─────────────────────────────────────────────────────────────────┤\n"))

	for i, r := range results {
		statusIcon := getDriftStatusIcon(r.Comparison.Status)
		severityIcon := getSeverityIcon(r.Severity)

		sb.WriteString(fmt.Sprintf("│ %s %-50s │\n", statusIcon, r.Comparison.Repository))
		sb.WriteString(fmt.Sprintf("│   %s → %s                                         │\n",
			r.Comparison.ProdBranch,
			r.Comparison.DevBranch,
		))
		sb.WriteString(fmt.Sprintf("│   Status: %-12s │ Severity: %s %-8s          │\n",
			r.Comparison.Status,
			severityIcon,
			r.Severity,
		))
		sb.WriteString(fmt.Sprintf("│   Ahead: %-3d │ Behind: %-3d │ Files: %-3d              │\n",
			r.Comparison.AheadBy,
			r.Comparison.BehindBy,
			len(r.Comparison.Files),
		))

		if len(r.Actions) > 0 {
			sb.WriteString(fmt.Sprintf("│   Recommended Actions:                                          │\n"))
			for _, action := range r.Actions {
				sb.WriteString(fmt.Sprintf("│     → %s\n", truncate(action, 55)))
			}
		}

		if i < len(results)-1 {
			sb.WriteString(fmt.Sprintf("├─────────────────────────────────────────────────────────────────┤\n"))
		}
	}

	sb.WriteString(fmt.Sprintf("└─────────────────────────────────────────────────────────────────┘\n"))
	sb.WriteString(fmt.Sprintf("Total: %d repositories with drift\n", len(results)))

	return sb.String()
}

func (p *Presenter) FormatRollbackResult(result *entity.RollbackResult) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("┌─────────────────────────────────────────────────────────────────┐\n"))
	sb.WriteString(fmt.Sprintf("│ ROLLBACK RESULT                                                 │\n"))
	sb.WriteString(fmt.Sprintf("├─────────────────────────────────────────────────────────────────┤\n"))

	status := "✓"
	if !result.Success {
		status = "✗"
	}

	sb.WriteString(fmt.Sprintf("│ %s Strategy: %-52s │\n", status, result.Strategy))
	sb.WriteString(fmt.Sprintf("│   %s\n", result.Message))

	if result.RunURL != "" {
		sb.WriteString(fmt.Sprintf("│   Run: %s\n", result.RunURL))
	}
	if result.PRURL != "" {
		sb.WriteString(fmt.Sprintf("│   PR: %s\n", result.PRURL))
	}

	sb.WriteString(fmt.Sprintf("└─────────────────────────────────────────────────────────────────┘\n"))

	return sb.String()
}

func (p *Presenter) FormatSyncPRResult(result *entity.SyncPRResult) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("┌─────────────────────────────────────────────────────────────────┐\n"))
	sb.WriteString(fmt.Sprintf("│ SYNC PR RESULT                                                  │\n"))
	sb.WriteString(fmt.Sprintf("├─────────────────────────────────────────────────────────────────┤\n"))

	status := "✓"
	if !result.Success {
		status = "✗"
	}

	sb.WriteString(fmt.Sprintf("│ %s %s\n", status, result.Message))

	if result.PRNumber > 0 {
		sb.WriteString(fmt.Sprintf("│   PR #%d                                                         │\n", result.PRNumber))
	}
	if result.PRURL != "" {
		sb.WriteString(fmt.Sprintf("│   %s\n", result.PRURL))
	}
	if result.FilesChanged > 0 || result.Commits > 0 {
		sb.WriteString(fmt.Sprintf("│   Files: %d │ Commits: %d                                       │\n",
			result.FilesChanged,
			result.Commits,
		))
	}

	sb.WriteString(fmt.Sprintf("└─────────────────────────────────────────────────────────────────┘\n"))

	return sb.String()
}

func (p *Presenter) FormatCreatePRResult(result *usecase.CreatePRResult) string {
	var sb strings.Builder
	sb.WriteString("┌─────────────────────────────────────────────────────────────────┐\n")
	sb.WriteString("│ CREATE PR RESULT                                                │\n")
	sb.WriteString("├─────────────────────────────────────────────────────────────────┤\n")

	status := "✓"
	if !result.Success {
		status = "✗"
	}

	sb.WriteString(fmt.Sprintf("│ %s %s\n", status, result.Message))

	if result.PR != nil {
		draftLabel := ""
		if result.PR.Draft {
			draftLabel = " (draft)"
		}
		sb.WriteString(fmt.Sprintf("│   PR #%d%s                                                      │\n", result.PR.Number, draftLabel))
		sb.WriteString(fmt.Sprintf("│   %s → %s\n", result.PR.HeadBranch, result.PR.BaseBranch))
		if result.PR.HTMLURL != "" {
			sb.WriteString(fmt.Sprintf("│   %s\n", result.PR.HTMLURL))
		}
	}

	if result.FilesChanged > 0 || result.Commits > 0 {
		sb.WriteString(fmt.Sprintf("│   Files: %d │ Commits: %d                                       │\n",
			result.FilesChanged,
			result.Commits,
		))
	}

	sb.WriteString("└─────────────────────────────────────────────────────────────────┘\n")

	return sb.String()
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04")
}

func getStatusIcon(conclusion string) string {
	switch conclusion {
	case "success":
		return "✓"
	case "failure":
		return "✗"
	case "cancelled":
		return "○"
	case "skipped":
		return "⊘"
	default:
		return "●"
	}
}

func getDriftStatusIcon(status entity.DriftStatus) string {
	switch status {
	case entity.DriftNone:
		return "✓"
	case entity.DriftProdAhead:
		return "↑"
	case entity.DriftDevAhead:
		return "↓"
	case entity.DriftDiverged:
		return "⇅"
	default:
		return "?"
	}
}

func getSeverityIcon(severity string) string {
	switch severity {
	case "none":
		return "○"
	case "low":
		return "●"
	case "medium":
		return "◉"
	case "high":
		return "◈"
	default:
		return "?"
	}
}
