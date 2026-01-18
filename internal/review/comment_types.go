package review

import (
	"context"
	"log"

	"github.com/pirakansa/moribito/internal/prompt"
)

const (
	commentReactionEyes     = "eyes"
	commentReactionThumbsUp = "+1"
	commentReactionConfused = "confused"
)

// PRCommentEvent represents a pull request comment event.
type PRCommentEvent struct {
	InstallationID int64
	Owner          string
	Repo           string
	Number         int
	CommentID      int64
	CommentBody    string
	CommentAuthor  string
}

// PRCommenter defines the interface for handling PR comment responses.
type PRCommenter interface {
	OnPullRequestComment(ctx context.Context, event PRCommentEvent) error
	OnPullRequestLabeled(ctx context.Context, event PRLabelEvent) error
}

// PRCommentService handles AI-powered PR comment responses.
type PRCommentService struct {
	logger         *log.Logger
	factory        ClientFactory
	opencodeClient OpenCodeClient
	promptBuilder  *prompt.Builder
	triggerPrefix  string
	model          string
	labelBuilder   *prompt.Builder
	labelModel     string
	labelTriggers  map[string]struct{}
}

// PRCommentOption configures the PRCommentService.
type PRCommentOption func(*PRCommentService)

// WithCommentOpenCodeClient sets the OpenCode client for PR comment responses.
func WithCommentOpenCodeClient(client OpenCodeClient) PRCommentOption {
	return func(s *PRCommentService) {
		s.opencodeClient = client
	}
}

// WithCommentPromptBuilder sets a custom prompt builder for PR comments.
func WithCommentPromptBuilder(builder *prompt.Builder) PRCommentOption {
	return func(s *PRCommentService) {
		s.promptBuilder = builder
	}
}

// WithCommentTriggerPrefix sets the comment prefix that triggers PR comment responses.
func WithCommentTriggerPrefix(prefix string) PRCommentOption {
	return func(s *PRCommentService) {
		s.triggerPrefix = prefix
	}
}

// WithCommentModel sets a specific OpenCode model for PR comment responses.
func WithCommentModel(model string) PRCommentOption {
	return func(s *PRCommentService) {
		s.model = model
	}
}

// WithCommentLabelPromptBuilder sets a custom prompt builder for label responses.
func WithCommentLabelPromptBuilder(builder *prompt.Builder) PRCommentOption {
	return func(s *PRCommentService) {
		s.labelBuilder = builder
	}
}

// WithCommentLabelTriggers sets labels that trigger PR responses.
func WithCommentLabelTriggers(labels []string) PRCommentOption {
	return func(s *PRCommentService) {
		s.labelTriggers = normalizeLabelSet(labels)
	}
}

// WithCommentLabelModel sets a specific OpenCode model for PR label responses.
func WithCommentLabelModel(model string) PRCommentOption {
	return func(s *PRCommentService) {
		s.labelModel = model
	}
}

// PRLabelEvent represents a pull request label event.
type PRLabelEvent struct {
	InstallationID int64
	Owner          string
	Repo           string
	Number         int
	LabelName      string
	Labeler        string
}

type commentOutcome struct {
	aiAttempted bool
	aiSucceeded bool
}
