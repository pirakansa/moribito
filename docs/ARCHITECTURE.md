# Architecture

M.O.R.I.B.I.T.O. is a GitHub App focused on webhook processing and automated PR review.
It is intentionally lightweight and avoids framework lock-in.

## High-Level Flow

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   GitHub    │────▶│   Server    │────▶│   Router    │────▶│   Handler   │
│  (Webhook)  │     │  (HTTP)     │     │  (Events)   │     │  (Logic)    │
└─────────────┘     └─────────────┘     └─────────────┘     └─────────────┘
                                                                   │
                                                                   ▼
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  OpenCode   │◀────│  AI Review  │◀────│   Review    │◀────│    Queue    │
│  (Server)   │     │  (Prompt)   │     │  Service    │     │   (Jobs)    │
└─────────────┘     └─────────────┘     └─────────────┘     └─────────────┘
                                               │
                                               ▼
                    ┌─────────────┐     ┌─────────────┐
                    │  GitHub API │◀────│   Client    │
                    │  (Actions)  │     │  (go-github)│
                    └─────────────┘     └─────────────┘
```

### Request Flow

1. GitHub sends a webhook event to the app server
2. Server verifies the webhook signature (if configured)
3. Router dispatches event to the appropriate handler
4. Handler enqueues a background job for processing
5. Worker executes the job:
   - Add reaction (👀) to acknowledge
   - Fetch PR diff from GitHub
   - Send diff to OpenCode for AI analysis (if available)
   - Post AI review as PR comment
    - Add reaction (👍/😕) on completion

## Package Structure

```
cmd/
└── moribito/           # CLI entry point, server lifecycle

internal/
├── config/             # Configuration loading & validation
├── githubapp/          # GitHub API client, JWT, token management
│   ├── client.go       # GitHubClient interface implementation
│   ├── client_factory.go # Creates authenticated clients per installation
│   ├── jwt.go          # App JWT generation
│   ├── token.go        # Installation token fetching
│   └── token_cache.go  # Token caching with auto-refresh
├── issue/              # Issue comment response service
│   ├── service.go      # Trigger detection → AI Response → Comment
│   ├── process.go      # Issue prompt construction + posting
│   ├── ai.go           # OpenCode request/response
│   └── types.go        # Event and config types
├── opencode/           # OpenCode server API client
│   ├── client.go       # HTTP client, health check, providers
│   ├── session_client.go # Session lifecycle management
│   ├── session_types.go  # Session types/requests
│   ├── message_client.go # Message sending/receiving
│   ├── message_helpers.go # Message helpers
│   └── message_types.go   # Message types/requests
├── prompt/             # AI prompt templates
│   ├── templates.go    # Template file loader
│   └── builder.go      # Prompt construction with options
├── queue/              # In-memory job queue with worker pool
├── review/             # PR review service
│   ├── review_service.go # Acknowledge → AI Review → Comment → Complete flow
│   ├── review_process.go # PR diff + OpenCode review
│   ├── review_types.go   # Review types/options
│   ├── comment_service.go # PR comment flow
│   ├── comment_process.go # PR comment prompt + posting
│   └── comment_types.go   # PR comment types/options
├── server/             # HTTP server and middleware
└── webhook/            # Event router and handlers
    ├── router.go       # Event type dispatch
    ├── helpers.go      # Common helpers
    ├── payloads.go     # Event payload types
    └── *_handlers.go   # Per-event handler functions

test/
└── fixtures/           # Webhook payload fixtures for tests
```

## Key Components

### GitHubClient

Provides the interface for GitHub API operations:

```go
type GitHubClient interface {
    AddIssueReaction(ctx, owner, repo string, number int, reaction string) error
    AddIssueComment(ctx, owner, repo string, number int, body string) error
    AddCommentReaction(ctx, owner, repo string, commentID int64, reaction string) error
    GetIssue(ctx, owner, repo string, number int) (*IssueInfo, error)
}
```

### ClientFactory

Creates authenticated GitHub clients for each installation:

```go
type ClientFactory interface {
    NewClient(ctx, installationID int64) (GitHubClient, error)
}
```

### Review Service

Handles PR events with a three-phase approach:

1. **Acknowledge**: Add 👀 reaction immediately
2. **AI Review**: Send diff to OpenCode for analysis (gracefully degrades if unavailable)
3. **Comment**: Post review as PR comment
4. **Complete**: Add 👍 on success or 😕 if AI failed

```go
func (s *Service) OnPullRequestOpened(ctx, pr) error {
    s.acknowledge(ctx, client, owner, repo, pr.Number)  // 👀
    outcome := s.process(ctx, client, owner, repo, pr.Number)  // AI Review → Comment
    s.complete(ctx, client, owner, repo, pr.Number, outcome)   // 👍/😕
}
```

### Issue Service

Handles issue comment events triggered by `@moribito` prefix:

1. **Check Trigger**: Verify comment starts with `@moribito`
2. **Acknowledge**: Add 👀 reaction to the comment
3. **AI Response**: Send issue context to OpenCode for analysis
4. **Reply**: Post AI response as a new comment
5. **Complete**: Add 👍 on success or 😕 if AI failed

```go
func (s *Service) OnIssueComment(ctx, event) error {
    if !s.ShouldRespond(event.CommentBody) { return nil }
    s.acknowledge(ctx, client, event)     // 👀
    outcome := s.process(ctx, client, event) // AI Response → Comment
    s.complete(ctx, client, event, outcome)  // 👍/😕
}
```

### OpenCode Client

Communicates with OpenCode server for AI-powered code analysis:

```go
type Client struct {
    baseURL string
    // ...
}

func (c *Client) IsHealthy(ctx context.Context) bool
func (c *Client) CreateSession(ctx, req) (*Session, error)
func (c *Client) SendMessage(ctx, sessionID, req) (*Message, error)
```

Features:
- Health check at startup
- Session-based API for conversation context
- Graceful degradation when unavailable

### Prompt Builder

Constructs prompts using customizable templates:

```go
builder := prompt.NewBuilder(
    prompt.WithTemplate(template),
    prompt.WithMaxDiffLength(50000),
)
prompt, err := builder.BuildPRReviewPrompt(ctx)
```
Where `template` is loaded via `prompt.LoadTemplateFromFile(path)`.
Template files are loaded from disk (see `docs/templates/` for samples).

### Job Queue

In-memory queue for background processing:

- Configurable worker pool size
- Graceful shutdown support
- Suitable for local development

## Authentication Flow

```
App ID + Private Key  ──▶  JWT (10 min TTL)
         │
         ▼
JWT + Installation ID  ──▶  Installation Token (1 hour TTL)
         │
         ▼
Installation Token  ──▶  GitHub API calls
```

## Design Decisions

- **No frameworks**: Uses stdlib `net/http` for simplicity
- **Interfaces**: Key components are interface-based for testability
- **In-memory queue**: Suitable for development; swap for Redis/SQS in production
- **Token caching**: Reduces API calls with automatic refresh before expiry

## Extending the App

### Adding New Event Handlers

1. Add handler in a new `internal/webhook/*_handlers.go` file
2. Register in `internal/webhook/router.go`
3. Add fixture in `test/fixtures/webhook/`

### Adding Review Logic

Extend `internal/review/review_process.go`:

```go
func (s *Service) process(ctx context.Context, client githubapp.GitHubClient, owner, repo string, number int, installationID int64) (reviewOutcome, error) {
    // Fetch PR diff
    // Analyze code
    // Post review comments
    return reviewOutcome{}, nil
}
```
