package mcp

import (
	"context"
	"fmt"

	"github.com/carlos/mcp-repo-monitor/internal/application/usecase"
	"github.com/mark3labs/mcp-go/mcp"
)

type Handler struct {
	listStatus      *usecase.ListStatusUseCase
	listPRs         *usecase.ListPRsUseCase
	checkCI         *usecase.CheckCIUseCase
	triggerRollback *usecase.TriggerRollbackUseCase
	recentCommits   *usecase.RecentCommitsUseCase
	checkDrift      *usecase.CheckDriftUseCase
	createSyncPR    *usecase.CreateSyncPRUseCase
	createPR        *usecase.CreatePRUseCase
	presenter       *Presenter
}

func NewHandler(
	listStatus *usecase.ListStatusUseCase,
	listPRs *usecase.ListPRsUseCase,
	checkCI *usecase.CheckCIUseCase,
	triggerRollback *usecase.TriggerRollbackUseCase,
	recentCommits *usecase.RecentCommitsUseCase,
	checkDrift *usecase.CheckDriftUseCase,
	createSyncPR *usecase.CreateSyncPRUseCase,
	createPR *usecase.CreatePRUseCase,
	presenter *Presenter,
) *Handler {
	return &Handler{
		listStatus:      listStatus,
		listPRs:         listPRs,
		checkCI:         checkCI,
		triggerRollback: triggerRollback,
		recentCommits:   recentCommits,
		checkDrift:      checkDrift,
		createSyncPR:    createSyncPR,
		createPR:        createPR,
		presenter:       presenter,
	}
}

func (h *Handler) HandleListStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := getArgs(req)

	input := usecase.ListStatusInput{
		Filter:   getString(args, "filter"),
		Archived: getBool(args, "archived"),
	}

	statuses, err := h.listStatus.Execute(ctx, input)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list repositories: %v", err)), nil
	}

	return mcp.NewToolResultText(h.presenter.FormatRepositoryStatuses(statuses)), nil
}

func (h *Handler) HandleListPRs(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := getArgs(req)

	input := usecase.ListPRsInput{
		Repository: getString(args, "repo"),
		State:      getString(args, "state"),
		Limit:      getInt(args, "limit"),
	}

	prs, err := h.listPRs.Execute(ctx, input)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list PRs: %v", err)), nil
	}

	return mcp.NewToolResultText(h.presenter.FormatPullRequests(prs)), nil
}

func (h *Handler) HandleCheckCI(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := getArgs(req)

	repo := getString(args, "repo")
	if repo == "" {
		return mcp.NewToolResultError("repo parameter is required"), nil
	}

	input := usecase.CheckCIInput{
		Repository: repo,
		Branch:     getString(args, "branch"),
		Workflow:   getString(args, "workflow"),
		Limit:      getInt(args, "limit"),
	}

	runs, err := h.checkCI.Execute(ctx, input)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to check CI: %v", err)), nil
	}

	return mcp.NewToolResultText(h.presenter.FormatWorkflowRuns(runs)), nil
}

func (h *Handler) HandleTriggerRollback(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := getArgs(req)

	repo := getString(args, "repo")
	if repo == "" {
		return mcp.NewToolResultError("repo parameter is required"), nil
	}

	strategy := getString(args, "strategy")
	if strategy == "" {
		return mcp.NewToolResultError("strategy parameter is required (rerun, revert, or workflow)"), nil
	}

	input := usecase.TriggerRollbackInput{
		Repository: repo,
		Strategy:   strategy,
		WorkflowID: int64(getInt(args, "workflow_id")),
		RunID:      int64(getInt(args, "run_id")),
		CommitSHA:  getString(args, "commit_sha"),
		DryRun:     getBool(args, "dry_run"),
	}

	result, err := h.triggerRollback.Execute(ctx, input)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Rollback failed: %v", err)), nil
	}

	return mcp.NewToolResultText(h.presenter.FormatRollbackResult(result)), nil
}

func (h *Handler) HandleRecentCommits(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := getArgs(req)

	input := usecase.RecentCommitsInput{
		Repository: getString(args, "repo"),
		Branch:     getString(args, "branch"),
		Since:      getString(args, "since"),
		Limit:      getInt(args, "limit"),
	}

	commits, err := h.recentCommits.Execute(ctx, input)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list commits: %v", err)), nil
	}

	return mcp.NewToolResultText(h.presenter.FormatCommits(commits)), nil
}

func (h *Handler) HandleCheckDrift(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := getArgs(req)

	input := usecase.CheckDriftInput{
		Repository: getString(args, "repo"),
	}

	results, err := h.checkDrift.Execute(ctx, input)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to check drift: %v", err)), nil
	}

	return mcp.NewToolResultText(h.presenter.FormatDriftResults(results)), nil
}

func (h *Handler) HandleCreateSyncPR(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := getArgs(req)

	repo := getString(args, "repo")
	if repo == "" {
		return mcp.NewToolResultError("repo parameter is required"), nil
	}

	input := usecase.CreateSyncPRInput{
		Repository: repo,
		Title:      getString(args, "title"),
		Body:       getString(args, "body"),
		DryRun:     getBool(args, "dry_run"),
	}

	result, err := h.createSyncPR.Execute(ctx, input)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create sync PR: %v", err)), nil
	}

	return mcp.NewToolResultText(h.presenter.FormatSyncPRResult(result)), nil
}

func (h *Handler) HandleCreatePR(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := getArgs(req)

	repo := getString(args, "repo")
	if repo == "" {
		return mcp.NewToolResultError("repo parameter is required"), nil
	}

	title := getString(args, "title")
	if title == "" {
		return mcp.NewToolResultError("title parameter is required"), nil
	}

	head := getString(args, "head")
	if head == "" {
		return mcp.NewToolResultError("head parameter is required"), nil
	}

	base := getString(args, "base")
	if base == "" {
		return mcp.NewToolResultError("base parameter is required"), nil
	}

	input := usecase.CreatePRInput{
		Repository: repo,
		Title:      title,
		Head:       head,
		Base:       base,
		Body:       getString(args, "body"),
		Draft:      getBool(args, "draft"),
		DryRun:     getBool(args, "dry_run"),
	}

	result, err := h.createPR.Execute(ctx, input)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create PR: %v", err)), nil
	}

	return mcp.NewToolResultText(h.presenter.FormatCreatePRResult(result)), nil
}

func getArgs(req mcp.CallToolRequest) map[string]any {
	if args, ok := req.Params.Arguments.(map[string]any); ok {
		return args
	}
	return make(map[string]any)
}

func getString(args map[string]any, key string) string {
	if v, ok := args[key].(string); ok {
		return v
	}
	return ""
}

func getInt(args map[string]any, key string) int {
	if v, ok := args[key].(float64); ok {
		return int(v)
	}
	return 0
}

func getBool(args map[string]any, key string) bool {
	if v, ok := args[key].(bool); ok {
		return v
	}
	return false
}
