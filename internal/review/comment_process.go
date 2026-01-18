package review

import (
	"context"
	"fmt"

	"github.com/pirakansa/moribito/internal/githubapp"
	"github.com/pirakansa/moribito/internal/prompt"
)

func (s *PRCommentService) process(ctx context.Context, client githubapp.GitHubClient, event PRCommentEvent) (commentOutcome, error) {
	return s.processWithBuilder(ctx, client, event, s.promptBuilder, s.model)
}

func (s *PRCommentService) processWithBuilder(ctx context.Context, client githubapp.GitHubClient, event PRCommentEvent, builder *prompt.Builder, model string) (commentOutcome, error) {
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

	promptText, err := builder.BuildPRReviewPrompt(prompt.PRReviewContext{
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

	response, err := s.requestAIResponseWithModel(ctx, promptText, event, model)
	if err != nil {
		s.logger.Printf("pr-comment: AI response failed: %v", err)
		return commentOutcome{aiAttempted: true, aiSucceeded: false}, nil
	}

	formattedResponse := builder.FormatReviewComment(response)
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
