# Configuration

## Environment Variables

### Server Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `APP_ADDR` | `:8080` | Address to listen on |
| `GITHUB_WEBHOOK_PATH` | `/webhook` | Webhook endpoint path |
| `QUEUE_WORKERS` | `2` | Number of queue workers |
| `QUEUE_BUFFER` | `100` | Queue buffer size |

### GitHub App Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `GITHUB_APP_ID` | (required) | GitHub App ID |
| `GITHUB_INSTALLATION_ID` | - | Installation ID (for CLI token command) |
| `GITHUB_PRIVATE_KEY_PATH` | (required) | Path to private key (PEM format) |
| `GITHUB_WEBHOOK_SECRET` | - | Webhook signature secret |
| `GITHUB_API_BASE_URL` | `https://api.github.com` | GitHub API base URL |

### OpenCode Integration

| Variable | Default | Description |
|----------|---------|-------------|
| `OPENCODE_HOST` | `127.0.0.1` | OpenCode server hostname |
| `OPENCODE_PORT` | `4096` | OpenCode server port |

The app automatically detects OpenCode availability at startup. If unavailable, AI review is disabled but the app continues to function normally.

### Prompt Templates

| Variable | Default | Description |
|----------|---------|-------------|
| `PR_REVIEW_TEMPLATE_PATH` | (required) | PR review prompt template file path |
| `ISSUE_RESPONSE_TEMPLATE_PATH` | (required) | Issue response prompt template file path |

#### Sample Templates

##### PR Review Templates

| File | Description |
|------|-------------|
| `docs/templates/pr-review.md.tmpl` | Standard code review in English |
| `docs/templates/pr-review-concise.md.tmpl` | Brief review focusing only on critical issues |
| `docs/templates/pr-review-ja.md.tmpl` | Review output in Japanese |
| `docs/templates/pr-review-security.md.tmpl` | Security-focused analysis |

##### Issue Response Templates

| File | Description |
|------|-------------|
| `docs/templates/issue-response.md.tmpl` | Standard issue comment response in English |
| `docs/templates/issue-response-ja.md.tmpl` | Issue response in Japanese |
| `docs/templates/issue-technical.md.tmpl` | Technical troubleshooting focus |

### Issue AI Response

The app responds to issue comments that start with `@moribito`. When triggered:
1. Adds 👀 reaction to acknowledge the comment
2. Sends the issue context to OpenCode for AI analysis
3. Posts the AI response as a new comment

Example trigger:
```
@moribito このエラーの原因を教えてください
```

## Usage Examples

### Basic Setup

```bash
export GITHUB_APP_ID=123
export GITHUB_PRIVATE_KEY_PATH=/path/to/private-key.pem
export PR_REVIEW_TEMPLATE_PATH=docs/templates/pr-review.md.tmpl
export ISSUE_RESPONSE_TEMPLATE_PATH=docs/templates/issue-response.md.tmpl
./bin/host/moribito
```

### With AI Review (OpenCode)

```bash
export GITHUB_APP_ID=123
export GITHUB_PRIVATE_KEY_PATH=/path/to/private-key.pem
export OPENCODE_HOST=127.0.0.1
export OPENCODE_PORT=4096
export QUEUE_WORKERS=2
export QUEUE_BUFFER=100
export PR_REVIEW_TEMPLATE_PATH=docs/templates/pr-review-ja.md.tmpl
export ISSUE_RESPONSE_TEMPLATE_PATH=docs/templates/issue-response-ja.md.tmpl
./bin/host/moribito
```

### Full Configuration

```bash
export APP_ADDR=:8080
export GITHUB_WEBHOOK_PATH=/webhook
export GITHUB_APP_ID=123
export GITHUB_INSTALLATION_ID=456
export GITHUB_PRIVATE_KEY_PATH=/path/to/private-key.pem
export GITHUB_WEBHOOK_SECRET=secret
export GITHUB_API_BASE_URL=https://api.github.com
export OPENCODE_HOST=127.0.0.1
export OPENCODE_PORT=4096
export QUEUE_WORKERS=2
export QUEUE_BUFFER=100
export PR_REVIEW_TEMPLATE_PATH=docs/templates/pr-review.md.tmpl
export ISSUE_RESPONSE_TEMPLATE_PATH=docs/templates/issue-response.md.tmpl
./bin/host/moribito
```

### GitHub Enterprise Server

```bash
export GITHUB_API_BASE_URL=https://github.example.com/api/v3
export GITHUB_APP_ID=123
export GITHUB_PRIVATE_KEY_PATH=/path/to/private-key.pem
export PR_REVIEW_TEMPLATE_PATH=docs/templates/pr-review.md.tmpl
export ISSUE_RESPONSE_TEMPLATE_PATH=docs/templates/issue-response.md.tmpl
./bin/host/moribito
```
