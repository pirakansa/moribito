# Operations

## Webhook Delivery

- Verify incoming events using `github.webhookSecret` in the JSON config.
- Log identifiers (`X-GitHub-Delivery`, `X-GitHub-Request-Id`) for traceability.

## Job Queue

- The queue is in-memory and will be lost on restart.
- Increase workers or buffer size via `queue.workers` / `queue.buffer` in the JSON config.

## GitHub Enterprise Server

- Set `github.apiBaseURL` to `https://<host>/api/v3`.
