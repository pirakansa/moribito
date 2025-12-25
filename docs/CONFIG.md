# Configuration

## Environment Variables

Required for webhook server:

- `APP_ADDR`: Address to listen on (default `:8080`)
- `GITHUB_WEBHOOK_PATH`: Webhook path (default `/webhook`)

Required for token generation:

- `GITHUB_APP_ID`: GitHub App ID
- `GITHUB_INSTALLATION_ID`: Installation ID
- `GITHUB_PRIVATE_KEY_PATH`: Private key path

Optional:

- `GITHUB_WEBHOOK_SECRET`: Webhook signature secret
- `GITHUB_API_BASE_URL`: GitHub API base URL (Enterprise Server uses `https://<host>/api/v3`)

## Usage Examples

```bash
export APP_ADDR=:8080
export GITHUB_WEBHOOK_PATH=/webhook
export GITHUB_APP_ID=123
export GITHUB_INSTALLATION_ID=456
export GITHUB_PRIVATE_KEY_PATH=/path/to/private-key.pem
export GITHUB_WEBHOOK_SECRET=secret
export GITHUB_API_BASE_URL=https://api.github.com
```
