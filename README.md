# mcp-repo-monitor

GitHub repository monitor for PRs, CI/CD status, and branch drift detection.

## Features

- **repo_list_status**: List all repositories with their status
- **repo_list_prs**: List open pull requests
- **repo_check_ci**: Check GitHub Actions status
- **repo_trigger_rollback**: Execute rollback strategies
- **repo_recent_commits**: View recent commits
- **repo_check_drift**: Detect differences between prod/dev branches
- **repo_create_sync_pr**: Create PR to sync branches

## Architecture

```
internal/
├── domain/           # Business logic & entities
│   ├── entity/       # Repository, PullRequest, Commit, Workflow, BranchComparison
│   └── service/      # DriftDetector, RollbackService
├── application/      # Use cases
│   ├── port/         # GitHubClient interface
│   └── usecase/      # 7 use cases for MCP tools
└── infrastructure/   # External interfaces
    ├── github/       # go-github implementation
    └── mcp/          # MCP server, handler, presenter
```

## Configuration

### 1. Secrets (.env)

Create `.env` file in the project root or `~/.mcp-repo-monitor/.env`:

```bash
cp .env.example .env
# Edit .env and add your GitHub token
```

```env
GITHUB_TOKEN=ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

**Required token scopes:**
- `repo` - Full access to repositories
- `workflow` - Execute workflows
- `read:user` - Read user profile

### 2. Branch Configuration (JSON)

Create `~/.mcp-repo-monitor/repos.json` for repository-specific branch mapping:

```json
{
  "default": {
    "prod_branch": "main",
    "dev_branch": "develop"
  },
  "repositories": {
    "my-org/webapp": {
      "prod_branch": "main",
      "dev_branch": "develop"
    },
    "my-org/api": {
      "prod_branch": "production",
      "dev_branch": "main"
    }
  }
}
```

The configuration loads in this order:
1. `.env` in current directory
2. `~/.mcp-repo-monitor/.env` (fallback)
3. `~/.mcp-repo-monitor/repos.json` (branch mapping)

## Usage

```bash
make build      # Build binary
make run        # Run stdio mode
make run-sse    # Run SSE mode
make inspect    # MCP Inspector UI
```

## Claude Integration

Add to `~/.claude/settings.json`:

```json
{
  "mcpServers": {
    "repo-monitor": {
      "command": "/path/to/mcp-repo-monitor/bin/mcp-repo-monitor"
    }
  }
}
```
