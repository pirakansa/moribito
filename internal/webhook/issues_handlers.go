package webhook

import (
	"context"
	"fmt"
	"log"

	"github.com/pirakansa/moribito/internal/githubapp"
	"github.com/pirakansa/moribito/internal/issue"
	"github.com/pirakansa/moribito/internal/queue"
)

// HandleIssues returns a handler for issue events.
// Actions include: opened, edited, labeled, unlabeled, etc.
// When action is "labeled", it triggers the issue service for AI response.
func HandleIssues(logger *log.Logger, submitter Submitter, resolver RepoServiceResolver) Handler {
	return func(ctx context.Context, event, delivery string, body []byte) (HandleResult, error) {
		payload, err := decodeJSON[issuesPayload](body)
		if err != nil {
			return HandleResult{}, fmt.Errorf("decode issues payload: %w", err)
		}

		_, _, issueService, ok := resolveRepoServices(resolver, payload.Repository.FullName)
		if !ok {
			return HandleResult{Skipped: true}, nil
		}

		logger.Printf("event=%s delivery=%s action=%s repo=%s issue=%d label=%s",
			event, delivery, payload.Action, payload.Repository.FullName,
			payload.Issue.Number, payload.Label.Name)

		if payload.Action == "labeled" && issueService != nil {
			owner, repo, perr := githubapp.ParseRepoFullName(payload.Repository.FullName)
			if perr != nil {
				return HandleResult{}, fmt.Errorf("parse repo name: %w", perr)
			}

			evt := issue.LabelEvent{
				InstallationID: payload.Installation.ID,
				Owner:          owner,
				Repo:           repo,
				IssueNumber:    payload.Issue.Number,
				IssueTitle:     payload.Issue.Title,
				IssueBody:      payload.Issue.Body,
				IssueAuthor:    payload.Issue.User.Login,
				IssueURL:       payload.Issue.HTMLURL,
				LabelName:      payload.Label.Name,
				Labeler:        payload.Sender.Login,
			}

			return HandleResult{}, enqueueJob(ctx, logger, submitter, queue.Job{
				Name: "issue_labeled_response",
				Run: func(jobCtx context.Context) error {
					logger.Printf("job=issue_labeled_response repo=%s issue=%d label=%s",
						payload.Repository.FullName, payload.Issue.Number, payload.Label.Name)
					return issueService.OnIssueLabeled(jobCtx, evt)
				},
			})
		}

		return HandleResult{}, enqueueJob(ctx, logger, submitter, queue.Job{
			Name: "issues",
			Run: func(_ context.Context) error {
				logger.Printf("job=issues action=%s repo=%s issue=%d label=%s",
					payload.Action, payload.Repository.FullName,
					payload.Issue.Number, payload.Label.Name)
				return nil
			},
		})
	}
}
