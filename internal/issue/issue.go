// Package issue provides AI-powered issue comment response functionality.
package issue

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/pirakansa/moribito/internal/githubapp"
	"github.com/pirakansa/moribito/internal/opencode"
	"github.com/pirakansa/moribito/internal/prompt"
)

const (
	issueReactionEyes     = "eyes"
	issueReactionThumbsUp = "+1"
	issueReactionConfused = "confused"
)

// IssueClient defines the GitHub API operations needed for issue handling.
type IssueClient interface {
	AddIssueReaction(ctx context.Context, owner, repo string, number int, reaction string) error
	AddIssueComment(ctx context.Context, owner, repo string, number int, body string) error
	AddCommentReaction(ctx context.Context, owner, repo string, commentID int64, reaction string) error
	GetIssue(ctx context.Context, owner, repo string, number int) (*githubapp.IssueInfo, error)
}

// ClientFactory creates authenticated GitHub clients per installation.
type ClientFactory interface {
	NewClient(ctx context.Context, installationID int64) (githubapp.GitHubClient, error)
}

// OpenCodeClient defines the OpenCode operations needed for issue responses.
type OpenCodeClient interface {
	IsHealthy(ctx context.Context) bool
	CreateSession(ctx context.Context, req *opencode.CreateSessionRequest) (*opencode.Session, error)
	SendMessage(ctx context.Context, sessionID string, req *opencode.SendMessageRequest) (*opencode.MessageWithParts, error)
	DeleteSession(ctx context.Context, sessionID string) error
}

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

// CommentEvent represents an issue comment event.
type CommentEvent struct {
	InstallationID int64
	Owner          string
	Repo           string
	IssueNumber    int
	CommentID      int64
	CommentBody    string
	CommentAuthor  string
	IssueTitle     string
	IssueBody      string
	IssueAuthor    string
	IssueURL       string
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

func (s *Service) process(ctx context.Context, client githubapp.GitHubClient, event CommentEvent) (issueOutcome, error) {
	// Check if OpenCode is available
	if s.opencodeClient == nil || !s.opencodeClient.IsHealthy(ctx) {
		s.logger.Printf("issue: opencode not available, skipping AI response")
		return issueOutcome{aiAttempted: false, aiSucceeded: true}, nil
	}

	// Build issue context
	issueCtx := prompt.IssueContext{
		Title:         event.IssueTitle,
		Number:        event.IssueNumber,
		Author:        event.IssueAuthor,
		Body:          event.IssueBody,
		URL:           event.IssueURL,
		Comment:       s.ExtractQuestion(event.CommentBody),
		CommentAuthor: event.CommentAuthor,
		CommentID:     event.CommentID,
		Owner:         event.Owner,
		Repo:          event.Repo,
		RepoFullName:  event.Owner + "/" + event.Repo,
	}

	// Build prompt
	promptText, err := s.promptBuilder.BuildIssueResponsePrompt(issueCtx)
	if err != nil {
		return issueOutcome{aiAttempted: false, aiSucceeded: false}, fmt.Errorf("build prompt: %w", err)
	}

	// Request AI response
	response, err := s.requestAIResponse(ctx, promptText)
	if err != nil {
		s.logger.Printf("issue: AI response failed: %v", err)
		return issueOutcome{aiAttempted: true, aiSucceeded: false}, nil
	}

	// Post response as comment
	formattedResponse := s.promptBuilder.FormatIssueResponse(response)
	commentClient, err := s.createClient(ctx, event.InstallationID)
	if err != nil {
		return issueOutcome{aiAttempted: true, aiSucceeded: true}, fmt.Errorf("refresh client: %w", err)
	}
	if err := commentClient.AddIssueComment(ctx, event.Owner, event.Repo, event.IssueNumber, formattedResponse); err != nil {
		return issueOutcome{aiAttempted: true, aiSucceeded: true}, fmt.Errorf("post comment: %w", err)
	}

	s.logger.Printf("issue: posted AI response to %s/%s#%d", event.Owner, event.Repo, event.IssueNumber)
	return issueOutcome{aiAttempted: true, aiSucceeded: true}, nil
}

func (s *Service) requestAIResponse(ctx context.Context, promptText string) (string, error) {
	// Create session for this issue response
	session, err := s.opencodeClient.CreateSession(ctx, &opencode.CreateSessionRequest{
		Title: "Issue Response",
	})
	if err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}
	defer func() {
		// Clean up session after use
		_ = s.opencodeClient.DeleteSession(context.Background(), session.ID)
	}()

	// Send prompt and get response
	req := opencode.NewTextMessageRequest(promptText)
	if s.model != "" {
		req = opencode.NewTextMessageRequestWithModel(promptText, s.model)
	}
	msg, err := s.opencodeClient.SendMessage(ctx, session.ID, req)
	if err != nil {
		return "", fmt.Errorf("send message: %w", err)
	}

	// Extract text from response
	response := opencode.ExtractTextFromResponse(msg)
	if response == "" {
		return "", fmt.Errorf("empty response from AI")
	}

	return response, nil
}
