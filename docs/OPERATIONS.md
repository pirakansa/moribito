# Operations

## Webhook Delivery

- Verify incoming events using `GITHUB_WEBHOOK_SECRET`.
- Log identifiers (`X-GitHub-Delivery`, `X-GitHub-Request-Id`) for traceability.

## Job Queue

- The queue is in-memory and will be lost on restart.
- Increase workers or buffer size in `cmd/moribito/main.go` if needed.

## GitHub Enterprise Server

- Set `GITHUB_API_BASE_URL` to `https://<host>/api/v3`.
