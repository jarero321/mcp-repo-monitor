<div align="center">

```
                                                    _ _
 _ __ ___ _ __   ___        _ __ ___   ___  _ __ (_) |_ ___  _ __
| '_ ` _ \ '_ \ / __|_____ | '_ ` _ \ / _ \| '_ \| | __/ _ \| '__|
| | | | | | |_) | (_|_____|| | | | | | (_) | | | | | || (_) | |
|_| |_| |_| .__/ \___|     |_| |_| |_|\___/|_| |_|_|\__\___/|_|
          |_|
```

### I wanted Claude to manage my GitHub repos. So I built an MCP server for it.

![Go](https://img.shields.io/badge/Go-00ADD8?logo=go&logoColor=white)
![MCP](https://img.shields.io/badge/MCP-Protocol-7c3aed?style=flat-square)
[![License](https://img.shields.io/badge/license-MIT-brightgreen?style=flat-square)](LICENSE)

**7 tools for GitHub automation via Claude Code**

[Tools](#tools) · [Setup](#setup) · [Configuration](#configuration) · [Claude Integration](#claude-code-integration) · [Troubleshooting](#troubleshooting)

</div>

---

## Why I Built This

I manage multiple repos and got tired of:
- Checking PR status manually
- Forgetting which branches drifted
- Context-switching to GitHub for CI status

Now I ask Claude: "Any failing CI runs?" and get instant answers.

---

## Tools

| Tool | What it does |
|------|--------------|
| `repo_list_status` | All repos with open PRs, CI status, last activity |
| `repo_list_prs` | Open pull requests across repos |
| `repo_check_ci` | GitHub Actions status for any branch |
| `repo_trigger_rollback` | Execute rollback strategies (rerun, revert, workflow) |
| `repo_recent_commits` | Recent commits with filters |
| `repo_check_drift` | Detect prod/dev branch differences |
| `repo_create_sync_pr` | Create PR to sync branches |

### Tool Examples

#### List Repository Status
```
"Show me the status of all my repositories"
"Which repos have failing CI?"
"List repos with open PRs"
```

#### Check Pull Requests
```
"Show me all open PRs"
"List PRs in my-org/api"
"Any PRs waiting for review?"
```

#### CI/CD Monitoring
```
"Any failing CI runs on main?"
"Check CI status for my-org/api"
"Show me recent workflow runs for feature-branch"
```

#### Drift Detection
```
"Check drift between main and develop"
"Are any of my repos out of sync?"
"Show drift for my-org/api"
```

#### Sync PRs
```
"Create a sync PR for my-org/api"
"Preview sync PR for my-repo (dry run)"
"Sync production into develop for all drifted repos"
```

#### Rollback Operations
```
"Rerun the last failed workflow on my-org/api"
"What rollback options are available?"
"Trigger rollback workflow for production"
```

---

## Setup

### 1. Get a GitHub Token

Create a [Personal Access Token](https://github.com/settings/tokens) with these scopes:
- `repo` - Full access to repositories
- `workflow` - Execute workflows
- `read:user` - Read user profile

### 2. Create `.env`

```bash
cp .env.example .env
# Add your token
GITHUB_TOKEN=ghp_xxxx
```

Alternatively, create `~/.mcp-repo-monitor/.env` for user-wide configuration.

---

## Configuration

### repos.json

Create `~/.mcp-repo-monitor/repos.json` to customize branch names per repository:

```json
{
  "default": {
    "prod_branch": "main",
    "dev_branch": "develop"
  },
  "repositories": {
    "my-org/api": {
      "prod_branch": "production",
      "dev_branch": "main"
    },
    "my-org/frontend": {
      "prod_branch": "release",
      "dev_branch": "develop"
    }
  }
}
```

| Field | Description |
|-------|-------------|
| `default.prod_branch` | Production branch name (default: `main`) |
| `default.dev_branch` | Development branch name (default: `develop`) |
| `repositories.<repo>` | Override branches for specific `owner/repo` |

### Command Line Flags

```bash
./mcp-repo-monitor --help

  --mode string      Server mode: stdio or sse (default "stdio")
  --addr string      Address for SSE server (default ":8080")
  --log-level string Log level: debug, info, warn, error (default "info")
```

---

## Usage

### Native

```bash
make build      # Build binary
make run        # Run stdio mode
make inspect    # MCP Inspector UI
```

### Docker

```bash
make docker-build
make docker-run
```

### Debug Mode

Run with verbose logging:

```bash
./bin/mcp-repo-monitor --log-level=debug
```

---

## Claude Code Integration

### Option 1: Native binary

Add to your Claude Code settings (`~/.claude/settings.json`):

```json
{
  "mcpServers": {
    "repo-monitor": {
      "command": "/path/to/mcp-repo-monitor/bin/mcp-repo-monitor"
    }
  }
}
```

### Option 2: Docker

```json
{
  "mcpServers": {
    "repo-monitor": {
      "command": "docker",
      "args": ["run", "-i", "--rm", "--env-file", "/path/to/.env", "mcp-repo-monitor"]
    }
  }
}
```

### Option 3: With debug logging

```json
{
  "mcpServers": {
    "repo-monitor": {
      "command": "/path/to/mcp-repo-monitor/bin/mcp-repo-monitor",
      "args": ["--log-level=debug"]
    }
  }
}
```

---

## Architecture

Clean Architecture with Go:

```
internal/
├── domain/           # Entities & business logic
│   ├── entity/       # Repository, PR, Commit, Workflow
│   └── service/      # DriftDetector, RollbackService
├── application/      # Use cases
│   ├── port/         # GitHubClient interface
│   └── usecase/      # 7 use cases
└── infrastructure/   # External
    ├── cache/        # In-memory caching
    ├── github/       # go-github client with rate limiting & retry
    ├── logging/      # Structured logging (slog)
    └── mcp/          # MCP server
```

### Features

- **Rate Limiting**: Automatically respects GitHub API limits
- **Retry with Backoff**: Handles transient failures (429, 500, 502, 503, 504)
- **Caching**: In-memory cache for frequently accessed data (5 min TTL)
- **Structured Logging**: JSON logs to stderr for debugging

---

## Troubleshooting

### "GITHUB_TOKEN environment variable is required"

Ensure your token is set:
```bash
# Check current directory
cat .env

# Or user config
cat ~/.mcp-repo-monitor/.env
```

### "invalid JSON in repos.json"

Validate your JSON syntax:
```bash
cat ~/.mcp-repo-monitor/repos.json | jq .
```

Common issues:
- Trailing commas
- Missing quotes around keys
- Unclosed brackets

### "invalid branch name"

Branch names cannot:
- Be empty
- Contain whitespace

Check your `repos.json`:
```json
{
  "default": {
    "prod_branch": "main",
    "dev_branch": "develop"
  }
}
```

### Rate Limit Errors

The client automatically waits when rate limits are low. For heavy usage:

1. Use a token with higher limits (GitHub Pro/Enterprise)
2. The cache reduces API calls for repeated queries
3. Check logs for rate limit warnings: `--log-level=debug`

### "branches are already in sync"

This is expected when `prod_branch` and `dev_branch` have no differences. Use `repo_check_drift` to verify.

### MCP Connection Issues

1. Verify the binary runs standalone:
   ```bash
   echo '{"jsonrpc":"2.0","method":"initialize","id":1}' | ./bin/mcp-repo-monitor
   ```

2. Check Claude Code logs for errors

3. Ensure the path in settings is absolute

---

## Post-Mortems

Documentamos incidentes y bugs críticos para aprender de ellos:

| ID | Fecha | Severidad | Título |
|----|-------|-----------|--------|
| [PM-001](docs/postmortem-001-drift-branch-comparison.md) | 2026-02-03 | CRÍTICA | Bug de comparación de ramas invertida en sync PR |

---

## License

MIT

---

<div align="center">

**[Report Bug](https://github.com/jarero321/mcp-repo-monitor/issues)** · **[Request Feature](https://github.com/jarero321/mcp-repo-monitor/issues)**

</div>
