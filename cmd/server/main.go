package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/carlos/mcp-repo-monitor/config"
	"github.com/carlos/mcp-repo-monitor/internal/application/usecase"
	"github.com/carlos/mcp-repo-monitor/internal/domain/service"
	"github.com/carlos/mcp-repo-monitor/internal/infrastructure/github"
	"github.com/carlos/mcp-repo-monitor/internal/infrastructure/mcp"
)

func main() {
	mode := flag.String("mode", "stdio", "Server mode: stdio or sse")
	addr := flag.String("addr", ":8080", "Address for SSE server")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if cfg.GitHubToken == "" {
		fmt.Fprintln(os.Stderr, "GITHUB_TOKEN environment variable is required")
		os.Exit(1)
	}

	ghClient := github.NewClient(cfg.GitHubToken)

	driftDetector := service.NewDriftDetector()
	rollbackService := service.NewRollbackService()

	listStatus := usecase.NewListStatusUseCase(ghClient)
	listPRs := usecase.NewListPRsUseCase(ghClient)
	checkCI := usecase.NewCheckCIUseCase(ghClient)
	triggerRollback := usecase.NewTriggerRollbackUseCase(ghClient, rollbackService)
	recentCommits := usecase.NewRecentCommitsUseCase(ghClient)
	checkDrift := usecase.NewCheckDriftUseCase(ghClient, cfg, driftDetector)
	createSyncPR := usecase.NewCreateSyncPRUseCase(ghClient, cfg)

	presenter := mcp.NewPresenter()
	handler := mcp.NewHandler(
		listStatus,
		listPRs,
		checkCI,
		triggerRollback,
		recentCommits,
		checkDrift,
		createSyncPR,
		presenter,
	)
	server := mcp.NewServer(handler)

	switch *mode {
	case "stdio":
		err = server.ServeStdio()
	case "sse":
		err = server.ServeSSE(*addr)
	default:
		fmt.Fprintf(os.Stderr, "Unknown mode: %s\n", *mode)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
