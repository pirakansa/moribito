package review

import (
	"context"
	"fmt"

	"github.com/pirakansa/moribito/internal/githubapp"
	"github.com/pirakansa/moribito/internal/opencode"
	"github.com/pirakansa/moribito/internal/prompt"
)

// process executes the main review logic for the pull request.
// This is where the actual analysis, review, or automation happens.
func (s *Service) process(ctx context.Context, client githubapp.GitHubClient, owner, repo string, number int, installationID int64) (reviewOutcome, error) {
	outcome := reviewOutcome{}
	s.logger.Printf("review: processing PR repo=%s/%s number=%d", owner, repo, number)

	// Step 1: Fetch PR details and diff
	prInfo, err := client.GetPullRequest(ctx, owner, repo, number)
	if err != nil {
		s.logger.Printf("review: failed to get PR details: %v", err)
		return outcome, err
	}
	s.logger.Printf("review: PR title=%q base=%s head=%s", prInfo.Title, prInfo.Base, prInfo.Head)

	diff, err := client.GetPullRequestDiff(ctx, owner, repo, number)
	if err != nil {
		s.logger.Printf("review: failed to get PR diff: %v", err)
		return outcome, err
	}
	s.logger.Printf("review: fetched diff size=%d bytes", len(diff))

	// Step 2: Check if OpenCode is available for AI review
	if s.opencodeClient == nil {
		s.logger.Printf("review: opencode client not configured, skipping AI review")
		s.logger.Printf("review: processing complete (no AI) for PR repo=%s/%s number=%d", owner, repo, number)
		return reviewOutcome{aiAttempted: false, aiSucceeded: true}, nil
	}

	if !s.opencodeClient.IsHealthy(ctx) {
		s.logger.Printf("review: opencode server not available, skipping AI review")
		s.logger.Printf("review: processing complete (no AI) for PR repo=%s/%s number=%d", owner, repo, number)
		return reviewOutcome{aiAttempted: false, aiSucceeded: true}, nil
	}

	// Step 3: Create session and request AI review
	reviewComment, err := s.requestAIReview(ctx, prInfo, diff, owner, repo, number)
	if err != nil {
		s.logger.Printf("review: AI review failed: %v", err)
		// Don't fail the whole process if AI review fails
		s.logger.Printf("review: continuing without AI review")
		return reviewOutcome{aiAttempted: true, aiSucceeded: false}, nil
	}

	outcome.aiAttempted = true
	outcome.aiSucceeded = true

	// Step 4: Post review comment
	if reviewComment != "" {
		commentClient, err := s.clientFactory.NewClient(ctx, installationID)
		if err != nil {
			s.logger.Printf("review: failed to refresh client: %v", err)
			return outcome, err
		}
		if err := commentClient.AddIssueComment(ctx, owner, repo, number, reviewComment); err != nil {
			s.logger.Printf("review: failed to post review comment: %v", err)
			return outcome, err
		}
		s.logger.Printf("review: posted AI review comment")
	}

	s.logger.Printf("review: processing complete for PR repo=%s/%s number=%d", owner, repo, number)
	return outcome, nil
}

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
		Title:        prInfo.Title,
		Body:         prInfo.Body,
		Head:         prInfo.Head,
		Base:         prInfo.Base,
		URL:          prInfo.HTMLURL,
		Diff:         diff,
		Owner:        owner,
		Repo:         repo,
		RepoFullName: owner + "/" + repo,
		Number:       number,
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
