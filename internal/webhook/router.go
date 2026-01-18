package webhook

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/pirakansa/moribito/internal/issue"
	"github.com/pirakansa/moribito/internal/review"
)

// Handler processes a GitHub webhook event.
// Parameters:
//   - ctx: context for cancellation and deadlines
//   - event: the event type (e.g., "pull_request", "issue_comment")
//   - delivery: unique delivery ID from GitHub
//   - body: raw JSON payload
type Handler func(ctx context.Context, event, delivery string, body []byte) error

// Router dispatches GitHub webhook events to their registered handlers.
// It provides a simple way to register handlers for specific event types
// and falls back gracefully for unhandled events.
type Router struct {
	logger   *log.Logger
	handlers map[string]Handler
}

// NewRouter creates a Router with default handlers for common GitHub App events.
// Default handlers:
//   - installation: app installed/uninstalled events
//   - installation_repositories: repository selection changes
//   - pull_request: PR opened, closed, synchronized, etc.
//   - issues: issues opened, labeled, etc.
//   - issue_comment: comments on issues and PRs
//   - check_run: CI check status updates
func NewRouter(logger *log.Logger, submitter Submitter, reviewer review.Reviewer, issueService *issue.Service, prCommenter review.PRCommenter) *Router {
	r := &Router{
		logger:   logger,
		handlers: make(map[string]Handler),
	}

	// Register default handlers for common GitHub App events
	r.Register("installation", HandleInstallation(logger, submitter))
	r.Register("installation_repositories", HandleInstallationRepositories(logger, submitter))
	r.Register("pull_request", HandlePullRequest(logger, submitter, reviewer, prCommenter))
	r.Register("issues", HandleIssues(logger, submitter, issueService))
	r.Register("issue_comment", HandleIssueComment(logger, submitter, issueService, prCommenter))
	r.Register("check_run", HandleCheckRun(logger, submitter))

	return r
}

// Register adds or replaces a handler for the specified event type.
// Event names should match GitHub webhook event names exactly.
func (r *Router) Register(event string, handler Handler) {
	r.handlers[event] = handler
}

// Handle dispatches an event to its registered handler.
// Returns nil for unknown events (logged but not treated as errors).
// Returns an error if the handler fails.
func (r *Router) Handle(ctx context.Context, event, delivery string, body []byte) error {
	event = strings.TrimSpace(event)
	if event == "" {
		return fmt.Errorf("missing event name")
	}

	handler, ok := r.handlers[event]
	if !ok {
		r.logger.Printf("webhook unhandled event=%s delivery=%s", event, delivery)
		return nil
	}

	if err := handler(ctx, event, delivery, body); err != nil {
		return fmt.Errorf("handle event %s: %w", event, err)
	}
	return nil
}
