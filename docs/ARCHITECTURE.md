# Architecture

This repository is a GitHub App skeleton focused on webhook ingestion and
app authentication. It is intentionally small and avoids framework lock-in.

## High-Level Flow

1. GitHub sends a webhook event to the app server.
2. The server verifies the webhook signature (if configured).
3. The event is routed to a handler based on its type.
4. Handlers enqueue background jobs for long-running work.
5. Jobs run in worker goroutines.

## Key Components

- `cmd/app`: CLI entry point and server lifecycle.
- `internal/server`: HTTP server and webhook endpoint.
- `internal/webhook`: Event router and per-event handlers.
- `internal/githubapp`: JWT creation, token fetch, and API client.
- `internal/queue`: In-memory job queue with worker pool.
- `test/fixtures`: Webhook payload fixtures for tests.

## Notes

- The queue is in-memory and suitable for local development.
- Long-running tasks should be implemented as queue jobs.
- GitHub Enterprise Server uses a different API base URL.
