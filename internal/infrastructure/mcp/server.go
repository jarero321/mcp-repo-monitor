package mcp

import (
	"fmt"
	"net/http"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type Server struct {
	mcpServer *server.MCPServer
	handler   *Handler
}

func NewServer(handler *Handler) *Server {
	s := server.NewMCPServer(
		"mcp-repo-monitor",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	srv := &Server{mcpServer: s, handler: handler}
	srv.registerTools()
	return srv
}

func (s *Server) registerTools() {
	s.mcpServer.AddTool(
		mcp.NewTool("repo_list_status",
			mcp.WithDescription("List all repositories with their status including open PRs, CI status, and last activity"),
			mcp.WithString("filter",
				mcp.Description("Filter repositories by name (partial match)"),
			),
			mcp.WithBoolean("archived",
				mcp.Description("Include archived repositories"),
			),
		),
		s.handler.HandleListStatus,
	)

	s.mcpServer.AddTool(
		mcp.NewTool("repo_list_prs",
			mcp.WithDescription("List open pull requests across repositories"),
			mcp.WithString("repo",
				mcp.Description("Filter by repository (owner/repo format)"),
			),
			mcp.WithString("state",
				mcp.Description("PR state: open, closed, all (default: open)"),
			),
			mcp.WithNumber("limit",
				mcp.Description("Maximum number of PRs to return (default: 30)"),
			),
		),
		s.handler.HandleListPRs,
	)

	s.mcpServer.AddTool(
		mcp.NewTool("repo_check_ci",
			mcp.WithDescription("Check GitHub Actions CI/CD status for a repository"),
			mcp.WithString("repo",
				mcp.Description("Repository in owner/repo format"),
				mcp.Required(),
			),
			mcp.WithString("branch",
				mcp.Description("Filter by branch name"),
			),
			mcp.WithString("workflow",
				mcp.Description("Filter by workflow file name (e.g., ci.yml)"),
			),
			mcp.WithNumber("limit",
				mcp.Description("Maximum number of runs to return (default: 10)"),
			),
		),
		s.handler.HandleCheckCI,
	)

	s.mcpServer.AddTool(
		mcp.NewTool("repo_trigger_rollback",
			mcp.WithDescription("Trigger a rollback using specified strategy"),
			mcp.WithString("repo",
				mcp.Description("Repository in owner/repo format"),
				mcp.Required(),
			),
			mcp.WithString("strategy",
				mcp.Description("Rollback strategy: rerun (re-run failed workflow), revert (create revert commit), workflow (trigger rollback workflow)"),
				mcp.Required(),
			),
			mcp.WithNumber("workflow_id",
				mcp.Description("Workflow ID for 'workflow' strategy"),
			),
			mcp.WithNumber("run_id",
				mcp.Description("Run ID for 'rerun' strategy (uses latest if not specified)"),
			),
			mcp.WithString("commit_sha",
				mcp.Description("Commit SHA for 'revert' strategy"),
			),
			mcp.WithBoolean("dry_run",
				mcp.Description("Preview the rollback without executing"),
			),
		),
		s.handler.HandleTriggerRollback,
	)

	s.mcpServer.AddTool(
		mcp.NewTool("repo_recent_commits",
			mcp.WithDescription("List recent commits across repositories"),
			mcp.WithString("repo",
				mcp.Description("Filter by repository (owner/repo format)"),
			),
			mcp.WithString("branch",
				mcp.Description("Filter by branch name"),
			),
			mcp.WithString("since",
				mcp.Description("Show commits since (RFC3339 format or duration like '24h', '7d')"),
			),
			mcp.WithNumber("limit",
				mcp.Description("Maximum number of commits to return (default: 30)"),
			),
		),
		s.handler.HandleRecentCommits,
	)

	s.mcpServer.AddTool(
		mcp.NewTool("repo_check_drift",
			mcp.WithDescription("Detect differences between production and development branches"),
			mcp.WithString("repo",
				mcp.Description("Repository in owner/repo format (checks all repos if not specified)"),
			),
		),
		s.handler.HandleCheckDrift,
	)

	s.mcpServer.AddTool(
		mcp.NewTool("repo_create_sync_pr",
			mcp.WithDescription("Create a PR to sync production branch into development branch"),
			mcp.WithString("repo",
				mcp.Description("Repository in owner/repo format"),
				mcp.Required(),
			),
			mcp.WithString("title",
				mcp.Description("Custom PR title"),
			),
			mcp.WithString("body",
				mcp.Description("Custom PR body"),
			),
			mcp.WithBoolean("dry_run",
				mcp.Description("Preview the PR without creating it"),
			),
		),
		s.handler.HandleCreateSyncPR,
	)

	s.mcpServer.AddTool(
		mcp.NewTool("repo_create_pr",
			mcp.WithDescription("Create a pull request between any two branches"),
			mcp.WithString("repo",
				mcp.Description("Repository in owner/repo format"),
				mcp.Required(),
			),
			mcp.WithString("title",
				mcp.Description("PR title"),
				mcp.Required(),
			),
			mcp.WithString("head",
				mcp.Description("Source branch"),
				mcp.Required(),
			),
			mcp.WithString("base",
				mcp.Description("Target branch"),
				mcp.Required(),
			),
			mcp.WithString("body",
				mcp.Description("PR description (markdown)"),
			),
			mcp.WithBoolean("draft",
				mcp.Description("Create as draft PR"),
			),
			mcp.WithBoolean("dry_run",
				mcp.Description("Preview without creating"),
			),
		),
		s.handler.HandleCreatePR,
	)
}

func (s *Server) ServeStdio() error {
	return server.ServeStdio(s.mcpServer)
}

func (s *Server) ServeSSE(addr string) error {
	sseServer := server.NewSSEServer(s.mcpServer)
	fmt.Printf("SSE Server starting on %s\n", addr)
	return http.ListenAndServe(addr, sseServer)
}
