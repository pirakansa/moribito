# Configuration

## Environment Variables

### Server Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `APP_ADDR` | `:8080` | Address to listen on |
| `GITHUB_WEBHOOK_PATH` | `/webhook` | Webhook endpoint path |

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
| `PROMPT_TEMPLATE` | `pr-review` | Prompt template for AI review |

#### Available Templates

| Template Name | Description |
|---------------|-------------|
| `pr-review` | Standard code review in English |
| `pr-review-concise` | Brief review focusing only on critical issues |
| `pr-review-ja` | Review output in Japanese |
| `pr-review-security` | Security-focused analysis |

## Usage Examples

### Basic Setup

```bash
export GITHUB_APP_ID=123
export GITHUB_PRIVATE_KEY_PATH=/path/to/private-key.pem
./bin/host/moribito
```

### With AI Review (OpenCode)

```bash
export GITHUB_APP_ID=123
export GITHUB_PRIVATE_KEY_PATH=/path/to/private-key.pem
export OPENCODE_HOST=127.0.0.1
export OPENCODE_PORT=4096
export PROMPT_TEMPLATE=pr-review-ja  # Japanese review
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
export PROMPT_TEMPLATE=pr-review
./bin/host/moribito
```

### GitHub Enterprise Server

```bash
export GITHUB_API_BASE_URL=https://github.example.com/api/v3
export GITHUB_APP_ID=123
export GITHUB_PRIVATE_KEY_PATH=/path/to/private-key.pem
./bin/host/moribito
```
