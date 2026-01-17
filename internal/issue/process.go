package issue

import (
	"context"
	"fmt"

	"github.com/pirakansa/moribito/internal/githubapp"
	"github.com/pirakansa/moribito/internal/prompt"
)

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
