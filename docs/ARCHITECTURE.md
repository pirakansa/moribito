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

## Package Structure

```
cmd/
└── moribito/           # CLI entry point, server lifecycle

internal/
├── config/             # Environment variable loading & validation
├── githubapp/          # GitHub API client, JWT, token management
│   ├── client.go       # GitHubClient interface implementation
│   ├── client_factory.go # Creates authenticated clients per installation
│   ├── jwt.go          # App JWT generation
│   ├── token.go        # Installation token fetching
│   └── token_cache.go  # Token caching with auto-refresh
├── issue/              # Issue comment response service
│   └── issue.go        # Trigger detection → AI Response → Comment
├── opencode/           # OpenCode server API client
│   ├── client.go       # HTTP client, health check, providers
│   ├── session.go      # Session lifecycle management
│   └── message.go      # Message sending/receiving
├── prompt/             # AI prompt templates
│   ├── templates.go    # Built-in prompt templates (PR + Issue)
│   └── builder.go      # Prompt construction with options
├── queue/              # In-memory job queue with worker pool
├── review/             # PR review service
│   └── review.go       # Acknowledge → AI Review → Comment flow
├── server/             # HTTP server and middleware
└── webhook/            # Event router and handlers
    ├── router.go       # Event type dispatch
    └── handlers.go     # Per-event handler functions

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

```go
func (s *Service) OnPullRequestOpened(ctx, pr) error {
    s.acknowledge(ctx, client, owner, repo, pr.Number)  // 👀
    s.process(ctx, client, owner, repo, pr.Number)      // AI Review → Comment
}
```

### Issue Service

Handles issue comment events triggered by `@moribito` prefix:

1. **Check Trigger**: Verify comment starts with `@moribito`
2. **Acknowledge**: Add 👀 reaction to the comment
3. **AI Response**: Send issue context to OpenCode for analysis
4. **Reply**: Post AI response as a new comment

```go
func (s *Service) OnIssueComment(ctx, event) error {
    if !s.ShouldRespond(event.CommentBody) { return nil }
    s.acknowledge(ctx, client, event)     // 👀
    s.process(ctx, client, event)         // AI Response → Comment
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
    prompt.WithTemplateName("pr-review-ja"),
    prompt.WithMaxDiffLength(50000),
)
prompt, err := builder.BuildPRReviewPrompt(ctx)
```

Built-in templates:

**PR Review:**
- `pr-review` - Standard code review
- `pr-review-concise` - Critical issues only
- `pr-review-ja` - Japanese output
- `pr-review-security` - Security focus

**Issue Response:**
- `issue-response` - Standard issue response
- `issue-response-ja` - Japanese output
- `issue-technical` - Technical troubleshooting

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

1. Add handler in `internal/webhook/handlers.go`
2. Register in `internal/webhook/router.go`
3. Add fixture in `test/fixtures/webhook/`

### Adding Review Logic

Extend `internal/review/review.go`:

```go
func (s *Service) process(ctx, client, owner, repo string, number int) error {
    // Fetch PR diff
    // Analyze code
    // Post review comments
    return nil
}
```
