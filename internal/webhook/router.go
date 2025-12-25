package webhook

import (
	"context"
	"fmt"
	"log"
	"strings"
)

// Handler handles a GitHub webhook event.
type Handler func(ctx context.Context, event, delivery string, body []byte) error

// Router routes events to their handlers.
type Router struct {
	logger   *log.Logger
	handlers map[string]Handler
}

// NewRouter builds a Router with default handlers.
func NewRouter(logger *log.Logger, submitter Submitter) *Router {
	r := &Router{
		logger:   logger,
		handlers: make(map[string]Handler),
	}
	r.Register("installation", HandleInstallation(logger, submitter))
	r.Register("installation_repositories", HandleInstallationRepositories(logger, submitter))
	r.Register("pull_request", HandlePullRequest(logger, submitter))
	r.Register("issue_comment", HandleIssueComment(logger, submitter))
	r.Register("check_run", HandleCheckRun(logger, submitter))
	return r
}

// Register adds a handler for the given event.
func (r *Router) Register(event string, handler Handler) {
	r.handlers[event] = handler
}

// Handle dispatches the event to the appropriate handler.
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
