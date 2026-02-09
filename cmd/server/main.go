package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/carlos/mcp-repo-monitor/config"
	"github.com/carlos/mcp-repo-monitor/internal/application/usecase"
	"github.com/carlos/mcp-repo-monitor/internal/domain/service"
	"github.com/carlos/mcp-repo-monitor/internal/infrastructure/cache"
	"github.com/carlos/mcp-repo-monitor/internal/infrastructure/github"
	"github.com/carlos/mcp-repo-monitor/internal/infrastructure/logging"
	"github.com/carlos/mcp-repo-monitor/internal/infrastructure/mcp"
)

func main() {
	mode := flag.String("mode", "stdio", "Server mode: stdio or sse")
	addr := flag.String("addr", ":8080", "Address for SSE server")
	logLevel := flag.String("log-level", "info", "Log level: debug, info, warn, error")
	flag.Parse()

	logger := logging.New(logging.ParseLevel(*logLevel))

	cfg, err := config.Load(logger)
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	if cfg.GitHubToken == "" {
		fmt.Fprintln(os.Stderr, "GITHUB_TOKEN environment variable is required")
		os.Exit(1)
	}

	ghClient := github.NewClient(cfg.GitHubToken, logger)

	// Create cache for frequently accessed data
	apiCache := cache.New(cache.DefaultConfig())
	cachedClient := github.NewCachedClient(ghClient, apiCache, logger)

	driftDetector := service.NewDriftDetector()
	rollbackService := service.NewRollbackService()

	// Use cached client for read-heavy use cases
	listStatus := usecase.NewListStatusUseCase(cachedClient)
	listPRs := usecase.NewListPRsUseCase(cachedClient)
	checkCI := usecase.NewCheckCIUseCase(ghClient) // CI status should be real-time
	triggerRollback := usecase.NewTriggerRollbackUseCase(ghClient, rollbackService)
	recentCommits := usecase.NewRecentCommitsUseCase(ghClient) // Commits should be real-time
	checkDrift := usecase.NewCheckDriftUseCase(ghClient, cfg, driftDetector) // Drift needs real-time
	createSyncPR := usecase.NewCreateSyncPRUseCase(ghClient, cfg)
	createPR := usecase.NewCreatePRUseCase(ghClient)
	mergePR := usecase.NewMergePRUseCase(ghClient)
	deleteBranch := usecase.NewDeleteBranchUseCase(ghClient)

	presenter := mcp.NewPresenter()
	handler := mcp.NewHandler(
		listStatus,
		listPRs,
		checkCI,
		triggerRollback,
		recentCommits,
		checkDrift,
		createSyncPR,
		createPR,
		mergePR,
		deleteBranch,
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
