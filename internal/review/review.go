// Package review provides functionality for automated code review
// triggered by pull request events.
package review

import (
	"context"
	"fmt"
	"log"

	"github.com/pirakansa/moribito/internal/githubapp"
	"github.com/pirakansa/moribito/internal/opencode"
	"github.com/pirakansa/moribito/internal/prompt"
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

// NewService creates a new review Service.
func NewService(logger *log.Logger, clientFactory ClientFactory, opts ...ServiceOption) *Service {
	s := &Service{
		logger:        logger,
		clientFactory: clientFactory,
		promptBuilder: prompt.NewBuilder(), // Template is required via options.
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// OnPullRequestOpened handles the event when a new pull request is created.
// Flow:
//  1. Acknowledge receipt with 👍 reaction
//  2. Process the pull request (review, analysis, etc.)
//  3. Post results (comments, status, etc.)
func (s *Service) OnPullRequestOpened(ctx context.Context, pr PullRequest) error {
	s.logger.Printf("review: pull request opened repo=%s number=%d", pr.RepoName, pr.Number)

	if s.clientFactory == nil {
		s.logger.Printf("review: no client factory configured, skipping")
		return nil
	}

	client, err := s.clientFactory.NewClient(ctx, pr.InstallationID)
	if err != nil {
		s.logger.Printf("review: failed to create client: %v", err)
		return err
	}

	owner, repo, err := githubapp.ParseRepoFullName(pr.RepoName)
	if err != nil {
		s.logger.Printf("review: failed to parse repo name: %v", err)
		return err
	}

	// Step 1: Acknowledge - Add 👀 reaction to show "request received"
	if err := s.acknowledge(ctx, client, owner, repo, pr.Number); err != nil {
		return err
	}

	// Step 2: Process - Execute the actual review logic
	if err := s.process(ctx, client, owner, repo, pr.Number); err != nil {
		return err
	}

	return nil
}

// acknowledge adds 👍 (thumbs up) reaction to indicate the request was received.
func (s *Service) acknowledge(ctx context.Context, client githubapp.GitHubClient, owner, repo string, number int) error {
	s.logger.Printf("review: acknowledging PR repo=%s/%s number=%d", owner, repo, number)

	if err := client.AddIssueReaction(ctx, owner, repo, number, "+1"); err != nil {
		s.logger.Printf("review: failed to add thumbs up reaction: %v", err)
		return err
	}

	s.logger.Printf("review: acknowledged PR with thumbs up reaction")
	return nil
}

// process executes the main review logic for the pull request.
// This is where the actual analysis, review, or automation happens.
func (s *Service) process(ctx context.Context, client githubapp.GitHubClient, owner, repo string, number int) error {
	s.logger.Printf("review: processing PR repo=%s/%s number=%d", owner, repo, number)

	// Step 1: Fetch PR details and diff
	prInfo, err := client.GetPullRequest(ctx, owner, repo, number)
	if err != nil {
		s.logger.Printf("review: failed to get PR details: %v", err)
		return err
	}
	s.logger.Printf("review: PR title=%q base=%s head=%s", prInfo.Title, prInfo.Base, prInfo.Head)

	diff, err := client.GetPullRequestDiff(ctx, owner, repo, number)
	if err != nil {
		s.logger.Printf("review: failed to get PR diff: %v", err)
		return err
	}
	s.logger.Printf("review: fetched diff size=%d bytes", len(diff))

	// Step 2: Check if OpenCode is available for AI review
	if s.opencodeClient == nil {
		s.logger.Printf("review: opencode client not configured, skipping AI review")
		s.logger.Printf("review: processing complete (no AI) for PR repo=%s/%s number=%d", owner, repo, number)
		return nil
	}

	if !s.opencodeClient.IsHealthy(ctx) {
		s.logger.Printf("review: opencode server not available, skipping AI review")
		s.logger.Printf("review: processing complete (no AI) for PR repo=%s/%s number=%d", owner, repo, number)
		return nil
	}

	// Step 3: Create session and request AI review
	reviewComment, err := s.requestAIReview(ctx, prInfo, diff, owner, repo, number)
	if err != nil {
		s.logger.Printf("review: AI review failed: %v", err)
		// Don't fail the whole process if AI review fails
		s.logger.Printf("review: continuing without AI review")
		return nil
	}

	// Step 4: Post review comment
	if reviewComment != "" {
		if err := client.AddIssueComment(ctx, owner, repo, number, reviewComment); err != nil {
			s.logger.Printf("review: failed to post review comment: %v", err)
			return err
		}
		s.logger.Printf("review: posted AI review comment")
	}

	s.logger.Printf("review: processing complete for PR repo=%s/%s number=%d", owner, repo, number)
	return nil
}

// requestAIReview sends the PR to OpenCode for AI-powered code review.
func (s *Service) requestAIReview(ctx context.Context, prInfo *githubapp.PullRequestInfo, diff, owner, repo string, number int) (string, error) {
	// Create a session for this review
	session, err := s.opencodeClient.CreateSession(ctx, &opencode.CreateSessionRequest{
		Title: fmt.Sprintf("PR Review: %s/%s#%d", owner, repo, number),
	})
	if err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}
	s.logger.Printf("review: created opencode session id=%s", session.ID)

	// Clean up session when done
	defer func() {
		if err := s.opencodeClient.DeleteSession(ctx, session.ID); err != nil {
			s.logger.Printf("review: failed to delete session: %v", err)
		}
	}()

	// Build the review prompt using the prompt builder
	reviewPrompt, err := s.promptBuilder.BuildPRReviewPrompt(prompt.PRReviewContext{
		Title: prInfo.Title,
		Body:  prInfo.Body,
		Head:  prInfo.Head,
		Base:  prInfo.Base,
		URL:   prInfo.HTMLURL,
		Diff:  diff,
	})
	if err != nil {
		return "", fmt.Errorf("build prompt: %w", err)
	}

	// Send message and wait for response
	req := opencode.NewTextMessageRequest(reviewPrompt)
	if s.model != "" {
		req = opencode.NewTextMessageRequestWithModel(reviewPrompt, s.model)
	}
	resp, err := s.opencodeClient.SendMessage(ctx, session.ID, req)
	if err != nil {
		return "", fmt.Errorf("send message: %w", err)
	}

	// Extract the review text from the response
	reviewText := opencode.ExtractTextFromResponse(resp)
	if reviewText == "" {
		s.logger.Printf("review: AI returned empty response")
		return "", nil
	}

	s.logger.Printf("review: received AI review length=%d", len(reviewText))
	return s.promptBuilder.FormatReviewComment(reviewText), nil
}
