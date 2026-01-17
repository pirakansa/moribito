// Package review provides functionality for automated PR review
// and PR comment responses.
package review

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
}

// PRCommentService handles AI-powered PR comment responses.
type PRCommentService struct {
	logger         *log.Logger
	factory        ClientFactory
	opencodeClient OpenCodeClient
	promptBuilder  *prompt.Builder
	triggerPrefix  string
	model          string
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

// NewPRCommentService creates a new PR comment response service.
func NewPRCommentService(logger *log.Logger, factory ClientFactory, opts ...PRCommentOption) *PRCommentService {
	s := &PRCommentService{
		logger:        logger,
		factory:       factory,
		promptBuilder: prompt.NewBuilder(),
		triggerPrefix: "@moribito",
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// ShouldRespond checks if the comment should trigger an AI response.
func (s *PRCommentService) ShouldRespond(comment string) bool {
	return strings.HasPrefix(strings.TrimSpace(comment), s.triggerPrefix)
}

// OnPullRequestComment handles PR comment events.
func (s *PRCommentService) OnPullRequestComment(ctx context.Context, event PRCommentEvent) error {
	s.logger.Printf("pr-comment: received comment on %s/%s#%d from @%s",
		event.Owner, event.Repo, event.Number, event.CommentAuthor)

	if !s.ShouldRespond(event.CommentBody) {
		s.logger.Printf("pr-comment: comment does not start with %q, skipping", s.triggerPrefix)
		return nil
	}

	client, err := s.createClient(ctx, event.InstallationID)
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}

	if err := client.AddCommentReaction(ctx, event.Owner, event.Repo, event.CommentID, commentReactionEyes); err != nil {
		s.logger.Printf("pr-comment: failed to acknowledge comment: %v", err)
	}

	outcome, err := s.process(ctx, client, event)
	if err != nil {
		return err
	}

	s.complete(ctx, client, event, outcome)
	return nil
}

func (s *PRCommentService) createClient(ctx context.Context, installationID int64) (githubapp.GitHubClient, error) {
	if s.factory == nil {
		return nil, fmt.Errorf("client factory not configured")
	}
	return s.factory.NewClient(ctx, installationID)
}

type commentOutcome struct {
	aiAttempted bool
	aiSucceeded bool
}

func (s *PRCommentService) complete(ctx context.Context, client githubapp.GitHubClient, event PRCommentEvent, outcome commentOutcome) {
	reaction := commentReactionThumbsUp
	if outcome.aiAttempted && !outcome.aiSucceeded {
		reaction = commentReactionConfused
	}

	if err := client.AddCommentReaction(ctx, event.Owner, event.Repo, event.CommentID, reaction); err != nil {
		s.logger.Printf("pr-comment: failed to add completion reaction %q: %v", reaction, err)
		return
	}

	s.logger.Printf("pr-comment: added completion reaction %q", reaction)
}

func (s *PRCommentService) process(ctx context.Context, client githubapp.GitHubClient, event PRCommentEvent) (commentOutcome, error) {
	if s.opencodeClient == nil || !s.opencodeClient.IsHealthy(ctx) {
		s.logger.Printf("pr-comment: opencode not available, skipping AI response")
		return commentOutcome{aiAttempted: false, aiSucceeded: true}, nil
	}

	prInfo, err := client.GetPullRequest(ctx, event.Owner, event.Repo, event.Number)
	if err != nil {
		return commentOutcome{aiAttempted: false, aiSucceeded: false}, fmt.Errorf("get pull request: %w", err)
	}

	diff, err := client.GetPullRequestDiff(ctx, event.Owner, event.Repo, event.Number)
	if err != nil {
		return commentOutcome{aiAttempted: false, aiSucceeded: false}, fmt.Errorf("get pull request diff: %w", err)
	}

	promptText, err := s.promptBuilder.BuildPRReviewPrompt(prompt.PRReviewContext{
		Title:        prInfo.Title,
		Body:         prInfo.Body,
		Head:         prInfo.Head,
		Base:         prInfo.Base,
		URL:          prInfo.HTMLURL,
		Diff:         diff,
		Owner:        event.Owner,
		Repo:         event.Repo,
		RepoFullName: event.Owner + "/" + event.Repo,
		Number:       event.Number,
	})
	if err != nil {
		return commentOutcome{aiAttempted: false, aiSucceeded: false}, fmt.Errorf("build prompt: %w", err)
	}

	response, err := s.requestAIResponse(ctx, promptText, event)
	if err != nil {
		s.logger.Printf("pr-comment: AI response failed: %v", err)
		return commentOutcome{aiAttempted: true, aiSucceeded: false}, nil
	}

	formattedResponse := s.promptBuilder.FormatReviewComment(response)
	commentClient, err := s.createClient(ctx, event.InstallationID)
	if err != nil {
		return commentOutcome{aiAttempted: true, aiSucceeded: true}, fmt.Errorf("refresh client: %w", err)
	}
	if err := commentClient.AddIssueComment(ctx, event.Owner, event.Repo, event.Number, formattedResponse); err != nil {
		return commentOutcome{aiAttempted: true, aiSucceeded: true}, fmt.Errorf("post comment: %w", err)
	}

	s.logger.Printf("pr-comment: posted AI response to %s/%s#%d", event.Owner, event.Repo, event.Number)
	return commentOutcome{aiAttempted: true, aiSucceeded: true}, nil
}

func (s *PRCommentService) requestAIResponse(ctx context.Context, promptText string, event PRCommentEvent) (string, error) {
	session, err := s.opencodeClient.CreateSession(ctx, &opencode.CreateSessionRequest{
		Title: fmt.Sprintf("PR Comment: %s/%s#%d", event.Owner, event.Repo, event.Number),
	})
	if err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}
	defer func() {
		_ = s.opencodeClient.DeleteSession(context.Background(), session.ID)
	}()

	req := opencode.NewTextMessageRequest(promptText)
	if s.model != "" {
		req = opencode.NewTextMessageRequestWithModel(promptText, s.model)
	}
	resp, err := s.opencodeClient.SendMessage(ctx, session.ID, req)
	if err != nil {
		return "", fmt.Errorf("send message: %w", err)
	}

	text := opencode.ExtractTextFromResponse(resp)
	if text == "" {
		s.logger.Printf("pr-comment: AI returned empty response")
		return "", nil
	}

	return text, nil
}
