package review

import (
	"context"
	"log"

	"github.com/pirakansa/moribito/internal/githubapp"
	"github.com/pirakansa/moribito/internal/opencode"
	"github.com/pirakansa/moribito/internal/prompt"
)

const (
	reactionEyes     = "eyes"
	reactionThumbsUp = "+1"
	reactionConfused = "confused"
)

// PullRequest holds information about a pull request.
type PullRequest struct {
	Number         int
	RepoName       string
	Action         string
	InstallationID int64
}

// ClientFactory creates GitHub API clients for installations.
type ClientFactory interface {
	NewClient(ctx context.Context, installationID int64) (githubapp.GitHubClient, error)
}

// OpenCodeClient defines the interface for OpenCode server operations.
// This allows for easy mocking in tests.
type OpenCodeClient interface {
	IsHealthy(ctx context.Context) bool
	CreateSession(ctx context.Context, req *opencode.CreateSessionRequest) (*opencode.Session, error)
	SendMessage(ctx context.Context, sessionID string, req *opencode.SendMessageRequest) (*opencode.MessageWithParts, error)
	DeleteSession(ctx context.Context, sessionID string) error
}

// Reviewer defines the interface for handling pull request reviews.
type Reviewer interface {
	// OnPullRequestOpened is called when a new pull request is created.
	OnPullRequestOpened(ctx context.Context, pr PullRequest) error
}

// Service implements the Reviewer interface and coordinates
// automated code review operations.
type Service struct {
	logger         *log.Logger
	clientFactory  ClientFactory
	opencodeClient OpenCodeClient
	promptBuilder  *prompt.Builder
	model          string
}

// ServiceOption configures the Service.
type ServiceOption func(*Service)

// WithOpenCodeClient sets the OpenCode client for AI-powered reviews.
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

// WithReviewModel sets a specific OpenCode model for PR reviews.
func WithReviewModel(model string) ServiceOption {
	return func(s *Service) {
		s.model = model
	}
}

type reviewOutcome struct {
	aiAttempted bool
	aiSucceeded bool
}
