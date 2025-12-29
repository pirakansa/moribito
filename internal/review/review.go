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
	NewClient(ctx context.Context, installationID int64) (githubapp.IssueReactor, error)
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
// This is the entry point for triggering automated code reviews.
func (s *Service) OnPullRequestOpened(ctx context.Context, pr PullRequest) error {
	s.logger.Printf("review: pull request opened repo=%s number=%d", pr.RepoName, pr.Number)

	if s.clientFactory == nil {
		s.logger.Printf("review: no client factory configured, skipping comment")
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

	// Add 👀 (eyes) reaction to indicate the hook was triggered
	if err := client.AddIssueReaction(ctx, owner, repo, pr.Number, "eyes"); err != nil {
		s.logger.Printf("review: failed to add reaction: %v", err)
		return err
	}

	s.logger.Printf("review: added eyes reaction to PR repo=%s number=%d", pr.RepoName, pr.Number)
	return nil
}
