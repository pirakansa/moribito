package issue

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/pirakansa/moribito/internal/githubapp"
	"github.com/pirakansa/moribito/internal/prompt"
)

// Service handles AI-powered issue comment responses.
type Service struct {
	logger         *log.Logger
	factory        ClientFactory
	opencodeClient OpenCodeClient
	promptBuilder  *prompt.Builder
	triggerPrefix  string // Comment prefix to trigger AI response (e.g., "@moribito")
	model          string
}

// ServiceOption configures the Service.
type ServiceOption func(*Service)

// WithOpenCodeClient sets the OpenCode client for AI responses.
func WithOpenCodeClient(client OpenCodeClient) ServiceOption {
	return func(s *Service) {
		s.opencodeClient = client
	}
}

// WithPromptBuilder sets a custom prompt builder.
func WithPromptBuilder(builder *prompt.Builder) ServiceOption {
	return func(s *Service) {
		s.promptBuilder = builder
	}
}

// WithTriggerPrefix sets the comment prefix that triggers AI response.
func WithTriggerPrefix(prefix string) ServiceOption {
	return func(s *Service) {
		s.triggerPrefix = prefix
	}
}

// WithResponseModel sets a specific OpenCode model for issue responses.
func WithResponseModel(model string) ServiceOption {
	return func(s *Service) {
		s.model = model
	}
}

// NewService creates a new issue response service.
func NewService(logger *log.Logger, factory ClientFactory, opts ...ServiceOption) *Service {
	s := &Service{
		logger:        logger,
		factory:       factory,
		promptBuilder: prompt.NewBuilder(), // Template is required via options.
		triggerPrefix: "@moribito",
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// ShouldRespond checks if the comment should trigger an AI response.
func (s *Service) ShouldRespond(comment string) bool {
	return strings.HasPrefix(strings.TrimSpace(comment), s.triggerPrefix)
}

// ExtractQuestion removes the trigger prefix from the comment.
func (s *Service) ExtractQuestion(comment string) string {
	trimmed := strings.TrimSpace(comment)
	if strings.HasPrefix(trimmed, s.triggerPrefix) {
		return strings.TrimSpace(trimmed[len(s.triggerPrefix):])
	}
	return trimmed
}

// OnIssueComment handles an issue comment event.
// It acknowledges the comment and responds with AI if configured.
func (s *Service) OnIssueComment(ctx context.Context, event CommentEvent) error {
	s.logger.Printf("issue: received comment on %s/%s#%d from @%s",
		event.Owner, event.Repo, event.IssueNumber, event.CommentAuthor)

	// Check if this comment should trigger a response
	if !s.ShouldRespond(event.CommentBody) {
		s.logger.Printf("issue: comment does not start with %q, skipping", s.triggerPrefix)
		return nil
	}

	// Create GitHub client
	client, err := s.createClient(ctx, event.InstallationID)
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}

	// Acknowledge the comment with eyes reaction
	if err := s.acknowledge(ctx, client, event); err != nil {
		s.logger.Printf("issue: failed to acknowledge comment: %v", err)
		// Continue even if reaction fails
	}

	// Process AI response
	outcome, err := s.process(ctx, client, event)
	if err != nil {
		return err
	}

	s.complete(ctx, client, event, outcome)
	return nil
}

func (s *Service) createClient(ctx context.Context, installationID int64) (githubapp.GitHubClient, error) {
	if s.factory == nil {
		return nil, fmt.Errorf("client factory not configured")
	}
	return s.factory.NewClient(ctx, installationID)
}

func (s *Service) acknowledge(ctx context.Context, client githubapp.GitHubClient, event CommentEvent) error {
	return client.AddCommentReaction(ctx, event.Owner, event.Repo, event.CommentID, issueReactionEyes)
}

type issueOutcome struct {
	aiAttempted bool
	aiSucceeded bool
}

func (s *Service) complete(ctx context.Context, client githubapp.GitHubClient, event CommentEvent, outcome issueOutcome) {
	reaction := issueReactionThumbsUp
	if outcome.aiAttempted && !outcome.aiSucceeded {
		reaction = issueReactionConfused
	}

	if err := client.AddCommentReaction(ctx, event.Owner, event.Repo, event.CommentID, reaction); err != nil {
		s.logger.Printf("issue: failed to add completion reaction %q: %v", reaction, err)
		return
	}

	s.logger.Printf("issue: added completion reaction %q", reaction)
}
