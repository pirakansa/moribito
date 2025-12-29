// Package review provides functionality for automated code review
// triggered by pull request events.
package review

import (
	"context"
	"log"

	"github.com/pirakansa/moribito/internal/githubapp"
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

// Reviewer defines the interface for handling pull request reviews.
type Reviewer interface {
	// OnPullRequestOpened is called when a new pull request is created.
	OnPullRequestOpened(ctx context.Context, pr PullRequest) error
}

// Service implements the Reviewer interface and coordinates
// automated code review operations.
type Service struct {
	logger        *log.Logger
	clientFactory ClientFactory
}

// NewService creates a new review Service.
func NewService(logger *log.Logger, clientFactory ClientFactory) *Service {
	return &Service{
		logger:        logger,
		clientFactory: clientFactory,
	}
}

// OnPullRequestOpened handles the event when a new pull request is created.
// Flow:
//  1. Acknowledge receipt with 👀 reaction
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

// acknowledge adds 👀 (eyes) reaction to indicate the request was received.
func (s *Service) acknowledge(ctx context.Context, client githubapp.GitHubClient, owner, repo string, number int) error {
	s.logger.Printf("review: acknowledging PR repo=%s/%s number=%d", owner, repo, number)

	if err := client.AddIssueReaction(ctx, owner, repo, number, "eyes"); err != nil {
		s.logger.Printf("review: failed to add eyes reaction: %v", err)
		return err
	}

	s.logger.Printf("review: acknowledged PR with eyes reaction")
	return nil
}

// process executes the main review logic for the pull request.
// This is where the actual analysis, review, or automation happens.
func (s *Service) process(ctx context.Context, client githubapp.GitHubClient, owner, repo string, number int) error {
	s.logger.Printf("review: processing PR repo=%s/%s number=%d", owner, repo, number)

	// TODO: Implement actual review logic here
	// Examples:
	// - Fetch PR diff and analyze code changes
	// - Run automated code review with AI
	// - Check for coding standards violations
	// - Validate PR description and title
	// - Post review comments with suggestions

	s.logger.Printf("review: processing complete for PR repo=%s/%s number=%d", owner, repo, number)
	return nil
}
