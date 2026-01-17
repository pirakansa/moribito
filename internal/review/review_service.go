package review

import (
	"context"
	"log"

	"github.com/pirakansa/moribito/internal/githubapp"
	"github.com/pirakansa/moribito/internal/prompt"
)

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
//  1. Acknowledge receipt with an eyes reaction
//  2. Process the pull request (review, analysis, etc.)
//  3. Post results (comments, status, etc.)
//  4. Acknowledge completion with a thumbs up or confused reaction based on AI outcome
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

	// Step 1: Acknowledge - Add eyes reaction to show "request received"
	if err := s.acknowledge(ctx, client, owner, repo, pr.Number); err != nil {
		return err
	}

	// Step 2: Process - Execute the actual review logic
	outcome, err := s.process(ctx, client, owner, repo, pr.Number, pr.InstallationID)
	if err != nil {
		return err
	}

	// Step 3: Complete - Add thumbs up or confused reaction based on AI outcome
	s.complete(ctx, client, owner, repo, pr.Number, outcome)

	return nil
}

// acknowledge adds eyes reaction to indicate the request was received.
func (s *Service) acknowledge(ctx context.Context, client githubapp.GitHubClient, owner, repo string, number int) error {
	s.logger.Printf("review: acknowledging PR repo=%s/%s number=%d", owner, repo, number)

	if err := client.AddIssueReaction(ctx, owner, repo, number, reactionEyes); err != nil {
		s.logger.Printf("review: failed to add eyes reaction: %v", err)
		return err
	}

	s.logger.Printf("review: acknowledged PR with eyes reaction")
	return nil
}

func (s *Service) complete(ctx context.Context, client githubapp.GitHubClient, owner, repo string, number int, outcome reviewOutcome) {
	reaction := reactionThumbsUp
	if outcome.aiAttempted && !outcome.aiSucceeded {
		reaction = reactionConfused
	}

	if err := client.AddIssueReaction(ctx, owner, repo, number, reaction); err != nil {
		s.logger.Printf("review: failed to add completion reaction %q: %v", reaction, err)
		return
	}

	s.logger.Printf("review: added completion reaction %q", reaction)
}
