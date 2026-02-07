<img width="100%" src="https://capsule-render.vercel.app/api?type=waving&color=gradient&customColorList=0,2,5,30&height=180&section=header&text=mcp-repo-monitor&fontSize=36&fontColor=fff&animation=fadeIn&fontAlignY=32" />

<div align="center">

![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![MCP](https://img.shields.io/badge/MCP-Protocol-7c3aed?style=for-the-badge)
![Tools](https://img.shields.io/badge/tools-8-00d4ff?style=for-the-badge)
![License](https://img.shields.io/github/license/jarero321/mcp-repo-monitor?style=for-the-badge)

**MCP server for GitHub automation via Claude Code. 8 tools, zero context-switching.**

<a href="https://github.com/jarero321/mcp-repo-monitor">
  <img src="https://img.shields.io/badge/CODE-2ea44f?style=for-the-badge&logo=github&logoColor=white" alt="code" />
</a>

[Tools](#tools) •
[Setup](#setup) •
[Configuration](#configuration) •
[Claude Integration](#claude-code-integration) •
[Architecture](#architecture)

</div>

---

## Features

| Feature | Description |
|:--------|:------------|
| **8 MCP Tools** | PRs, CI, drift, rollback, commits, sync, and general PR creation |
| **Clean Architecture** | Domain, application, and infrastructure layers with DI |
| **Rate Limiting** | Automatic GitHub API rate limit tracking and throttling |
| **Retry with Backoff** | Exponential backoff for transient failures (429, 5xx) |
| **In-Memory Cache** | TTL-based caching (5 min) for frequently accessed data |
| **Dual Mode** | Run as stdio (Claude) or SSE (HTTP) server |
| **Docker Support** | Build and run with Docker or docker-compose |
| **Branch Config** | Per-repo branch customization via `repos.json` |

## Tech Stack

<div align="center">

**Languages & Frameworks**

<img src="https://skillicons.dev/icons?i=go&perline=8" alt="languages" />

**Infrastructure & Tools**

<img src="https://skillicons.dev/icons?i=docker,github,githubactions&perline=8" alt="infra" />

</div>

## Why I Built This

I manage multiple repos and got tired of:
- Checking PR status manually
- Forgetting which branches drifted
- Context-switching to GitHub for CI status

Now I ask Claude: "Any failing CI runs?" and get instant answers.

---

## Tools

| Tool | Description |
|:-----|:------------|
| `repo_list_status` | All repos with open PRs, CI status, last activity |
| `repo_list_prs` | Open pull requests across repos |
| `repo_check_ci` | GitHub Actions status for any branch |
| `repo_trigger_rollback` | Execute rollback strategies (rerun, revert, workflow) |
| `repo_recent_commits` | Recent commits with filters |
| `repo_check_drift` | Detect prod/dev branch differences |
| `repo_create_sync_pr` | Create PR to sync prod into dev |
| `repo_create_pr` | Create PR between any two branches |

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

#### Sync & PR Creation
```
"Create a sync PR for my-org/api"
"Preview sync PR for my-repo (dry run)"
"Create a PR from feature/auth to develop"
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
    }
  }
}
```

| Field | Description |
|:------|:------------|
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
make run-sse    # Run SSE server on :8080
make inspect    # MCP Inspector UI
make test       # Run all tests
```

### Docker

```bash
make docker-build
make docker-run
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
│   ├── entity/       # Repository, PR, Commit, Workflow, Branch
│   └── service/      # DriftDetector, RollbackService
├── application/      # Use cases
│   ├── port/         # GitHubClient interface
│   └── usecase/      # 8 use cases
└── infrastructure/   # External services
    ├── cache/        # In-memory TTL caching
    ├── github/       # go-github client with rate limiting & retry
    ├── logging/      # Structured JSON logging (slog)
    └── mcp/          # MCP server, handlers, presenters
```

| Aspect | Choice |
|:-------|:-------|
| **Architecture** | Clean Architecture with DI |
| **Language** | Go 1.23 |
| **GitHub Client** | google/go-github v60 |
| **MCP Framework** | mark3labs/mcp-go |
| **Auth** | OAuth2 via golang.org/x/oauth2 |
| **Caching** | In-memory TTL (5 min default) |

---

## Troubleshooting

### "GITHUB_TOKEN environment variable is required"

Ensure your token is set:
```bash
cat .env
# Or user config
cat ~/.mcp-repo-monitor/.env
```

### "invalid JSON in repos.json"

Validate your JSON syntax:
```bash
cat ~/.mcp-repo-monitor/repos.json | jq .
```

### Rate Limit Errors

The client automatically waits when rate limits are low. For heavy usage:
1. Use a token with higher limits (GitHub Pro/Enterprise)
2. The cache reduces API calls for repeated queries
3. Check logs: `--log-level=debug`

### MCP Connection Issues

1. Verify the binary runs standalone:
   ```bash
   echo '{"jsonrpc":"2.0","method":"initialize","id":1}' | ./bin/mcp-repo-monitor
   ```
2. Check Claude Code logs for errors
3. Ensure the path in settings is absolute

---

## Post-Mortems

| ID | Date | Severity | Title |
|:---|:-----|:---------|:------|
| [PM-001](docs/postmortem-001-drift-branch-comparison.md) | 2026-02-03 | CRITICAL | Inverted branch comparison in sync PR |

---

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

<div align="center">

**[Report Bug](https://github.com/jarero321/mcp-repo-monitor/issues)** · **[Request Feature](https://github.com/jarero321/mcp-repo-monitor/issues)**

</div>

<img width="100%" src="https://capsule-render.vercel.app/api?type=waving&color=gradient&customColorList=0,2,5,30&height=120&section=footer" />
