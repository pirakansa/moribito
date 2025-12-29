<p align="center">
  <img src="./docs/moribito-logo.png" width="320" alt="MORIBITO logo"/>
</p>

<h1 align="center">M.O.R.I.B.I.T.O.</h1>
<p align="center">
  <strong>Monitoring Orchestrator for Repository Issues and Build Integrity; Trustee Officer.</strong><br>
  A GitHub App appointed to oversee issues, reviews, and entrusted operations within your repository.
</p>

---

## About

**M.O.R.I.B.I.T.O.** is a GitHub App that acts as a repository's appointed "Trustee Officer."  
It monitors activity, supports discussions, assists reviews, and can execute entrusted tasks on your behalf.

Built with Go using `net/http` for webhook handling and `go-github` for GitHub API interactions.

## Features

- **Webhook Processing**: Receives and validates GitHub webhook events
- **PR Review Automation**: Automatically acknowledges new PRs with 👀 reaction
- **AI-Powered Code Review**: Integrates with [OpenCode](https://opencode.ai) server for intelligent code reviews
- **Issue AI Response**: Responds to issue comments starting with `@moribito`
- **Customizable Prompts**: Multiple built-in templates for PR reviews and issue responses
- **Async Job Queue**: Background processing for long-running tasks
- **Extensible Architecture**: Easy to add new event handlers and review logic
- **Graceful Degradation**: Works without AI when OpenCode is unavailable

## Quick Start

### Prerequisites

- Go 1.21+
- GitHub App credentials (App ID, Private Key)

### Installation

```bash
git clone <repository-url>
cd moribito
go mod tidy
make build
```

### Configuration

Set the required environment variables:

```bash
export GITHUB_APP_ID=123
export GITHUB_PRIVATE_KEY_PATH=/path/to/private-key.pem
```

### Run

```bash
./bin/host/moribito
```

The server starts on `:8080` with endpoints:
- `GET /healthz` - Health check
- `POST /webhook` - GitHub webhook receiver

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `APP_ADDR` | No | Listen address (default: `:8080`) |
| `GITHUB_APP_ID` | **Yes** | GitHub App ID for JWT creation |
| `GITHUB_PRIVATE_KEY_PATH` | **Yes** | Path to App private key (PEM) |
| `GITHUB_INSTALLATION_ID` | No | Installation ID (for CLI token command) |
| `GITHUB_WEBHOOK_SECRET` | No | Webhook signature secret |
| `GITHUB_WEBHOOK_PATH` | No | Webhook endpoint path (default: `/webhook`) |
| `GITHUB_API_BASE_URL` | No | API URL (default: `https://api.github.com`) |
| `OPENCODE_HOST` | No | OpenCode server hostname (default: `127.0.0.1`) |
| `OPENCODE_PORT` | No | OpenCode server port (default: `4096`) |
| `PROMPT_TEMPLATE` | No | Review prompt template (default: `pr-review`) |

## PR Review Flow

When a Pull Request is opened, the app follows this flow:

```
PR Created → Webhook Received → Acknowledge (👀) → AI Review → Post Comment
```

1. **Acknowledge**: Adds 👀 (eyes) reaction to show the request was received
2. **Fetch Diff**: Retrieves PR diff from GitHub API
3. **AI Review**: Sends diff to OpenCode server for analysis (if available)
4. **Post Comment**: Posts AI review as a PR comment

### Prompt Templates

| Template | Description |
|----------|-------------|
| `pr-review` | Standard code review (default) |
| `pr-review-concise` | Brief review focusing on critical issues |
| `pr-review-ja` | Review output in Japanese |
| `pr-review-security` | Security-focused analysis |
| `issue-response` | Standard issue comment response |
| `issue-response-ja` | Issue response in Japanese |
| `issue-technical` | Technical troubleshooting focus |

## Issue AI Response

The app responds to issue comments that start with `@moribito`:

```
User Comment → @moribito detected? → Acknowledge (👀) → AI Response → Post Comment
```

### How to Use

In any issue, add a comment starting with `@moribito`:

```
@moribito How do I fix this error?
```

The bot will:
1. Add 👀 reaction to acknowledge
2. Analyze the issue context with AI
3. Post a helpful response

## Architecture

```
cmd/moribito/          # CLI entry point
internal/
├── config/            # Environment configuration
├── githubapp/         # GitHub API client & authentication
├── opencode/          # OpenCode server API client
├── prompt/            # Prompt templates & builder
├── queue/             # Background job queue
├── review/            # PR review service
├── server/            # HTTP server
└── webhook/           # Event routing & handlers
test/fixtures/         # Test data
```

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for details.

## Development

### Build & Test

```bash
make build    # Build binary
make test     # Run tests
make lint     # Run linter
```

### Local Webhook Testing with smee

1. Create a smee channel at [smee.io](https://smee.io)
2. Set the smee URL as your GitHub App's Webhook URL
3. Run the smee client:

```bash
npx smee-client -u https://smee.io/<your-id> -p 8080 -P /webhook
```

4. Start the app:

```bash
export GITHUB_APP_ID=<your-app-id>
export GITHUB_PRIVATE_KEY_PATH=<path-to-key.pem>
./bin/host/moribito
```

### Verify Installation Token

```bash
export GITHUB_APP_ID=123
export GITHUB_INSTALLATION_ID=456
export GITHUB_PRIVATE_KEY_PATH=/path/to/private-key.pem

./bin/host/moribito --print-installation-token
```

## Authentication Flow

```mermaid
sequenceDiagram
    participant GitHub
    participant App as M.O.R.I.B.I.T.O.
    participant API as GitHub API

    GitHub->>App: Webhook (pull_request opened)
    App->>App: Parse payload, extract Installation ID
    App->>App: Generate JWT (App ID + Private Key)
    App->>API: Request Installation Token
    API-->>App: Installation Token
    App->>API: Add reaction to PR (👀)
    App->>App: Execute review logic
```

## Documentation

- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) - System architecture
- [docs/CONFIG.md](docs/CONFIG.md) - Configuration details
- [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) - Development guide
- [docs/OPERATIONS.md](docs/OPERATIONS.md) - Operations guide

## License

See [LICENSE](LICENSE) file.
