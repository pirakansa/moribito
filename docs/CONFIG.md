# Configuration

Set the JSON config path via:

| Variable | Required | Description |
|----------|----------|-------------|
| `MORIBITO_CONFIG_PATH` | **Yes** | Path to JSON config file |

## JSON Format

```json
{
  "server": {
    "addr": ":8080",
    "webhookPath": "/webhook"
  },
  "github": {
    "appID": 123,
    "installationID": 456,
    "privateKeyPath": "/path/to/private-key.pem",
    "webhookSecret": "secret",
    "apiBaseURL": "https://api.github.com"
  },
  "opencode": {
    "host": "127.0.0.1",
    "port": 4096,
    "longTimeoutSeconds": 600
  },
  "queue": {
    "workers": 2,
    "buffer": 100
  },
  "review": {
    "templatePath": "docs/templates/pr-review.md.tmpl",
    "model": "opencode/big-pickle",
    "maxDiffLength": 50000
  },
  "issue": {
    "responseTemplatePath": "docs/templates/issue-response.md.tmpl",
    "responseModel": "opencode/big-pickle",
    "triggerPrefix": "@moribito"
  }
}
```

Defaults apply when fields are omitted. Set `review.maxDiffLength` to `0` to omit diffs entirely.

## Template Variables

### PR Review Templates

Available fields:
`Title`, `Body`, `Head`, `Base`, `URL`, `Diff`, `Owner`, `Repo`, `RepoFullName`, `Number`

### Issue Response Templates

Available fields:
`Title`, `Number`, `Author`, `Body`, `URL`, `Comment`, `CommentAuthor`, `CommentID`, `Owner`, `Repo`, `RepoFullName`

## Sample Templates

### PR Review Templates

| File | Description |
|------|-------------|
| `docs/templates/pr-review.md.tmpl` | Standard code review in English |
| `docs/templates/pr-review-concise.md.tmpl` | Brief review focusing only on critical issues |
| `docs/templates/pr-review-ja.md.tmpl` | Review output in Japanese |
| `docs/templates/pr-review-security.md.tmpl` | Security-focused analysis |

### Issue Response Templates

| File | Description |
|------|-------------|
| `docs/templates/issue-response.md.tmpl` | Standard issue comment response in English |
| `docs/templates/issue-response-ja.md.tmpl` | Issue response in Japanese |
| `docs/templates/issue-technical.md.tmpl` | Technical troubleshooting focus |

## Issue AI Response

The app responds to issue comments that start with `@moribito`. When triggered:
1. Adds 👍 reaction to acknowledge the comment
2. Sends the issue context to OpenCode for AI analysis
3. Posts the AI response as a new comment

Example trigger:
```
@moribito このエラーの原因を教えてください
```

## Usage Example

```bash
export MORIBITO_CONFIG_PATH=/path/to/moribito.json
./bin/host/moribito
```
