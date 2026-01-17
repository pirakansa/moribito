package review

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/pirakansa/moribito/internal/githubapp"
	"github.com/pirakansa/moribito/internal/prompt"
)

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
