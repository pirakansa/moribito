package webhook

import (
	"context"
	"fmt"
	"log"

	"github.com/pirakansa/moribito/internal/issue"
	"github.com/pirakansa/moribito/internal/queue"
	"github.com/pirakansa/moribito/internal/review"
)

// HandleIssueComment returns a handler for issue comment events.
// Actions include: created, edited, deleted.
// When issueService is provided and action is "created", the handler
// processes the comment for AI response.
func HandleIssueComment(logger *log.Logger, submitter Submitter, issueService *issue.Service, prCommenter review.PRCommenter) Handler {
	return func(ctx context.Context, event, delivery string, body []byte) error {
		payload, err := decodeJSON[issueCommentPayload](body)
		if err != nil {
			return fmt.Errorf("decode issue_comment payload: %w", err)
		}

		logger.Printf("event=%s delivery=%s action=%s repo=%s issue=%d comment_id=%d",
			event, delivery, payload.Action, payload.Repository.FullName,
			payload.Issue.Number, payload.Comment.ID)

		// Handle comment created event for AI response
		if payload.Action == "created" {
			owner, repo, perr := parseRepoFullName(payload.Repository.FullName)
			if perr != nil {
				return fmt.Errorf("parse repo name: %w", perr)
			}

			if payload.Issue.PullRequest != nil && prCommenter != nil {
				evt := review.PRCommentEvent{
					InstallationID: payload.Installation.ID,
					Owner:          owner,
					Repo:           repo,
					Number:         payload.Issue.Number,
					CommentID:      payload.Comment.ID,
					CommentBody:    payload.Comment.Body,
					CommentAuthor:  payload.Comment.User.Login,
				}

				return enqueueJob(ctx, logger, submitter, queue.Job{
					Name: "pull_request_comment_response",
					Run: func(jobCtx context.Context) error {
						logger.Printf("job=pull_request_comment_response repo=%s number=%d",
							payload.Repository.FullName, payload.Issue.Number)
						return prCommenter.OnPullRequestComment(jobCtx, evt)
					},
				})
			}

			if payload.Issue.PullRequest == nil && issueService != nil {
				evt := issue.CommentEvent{
					InstallationID: payload.Installation.ID,
					IssueNumber:    payload.Issue.Number,
					IssueTitle:     payload.Issue.Title,
					IssueBody:      payload.Issue.Body,
					IssueAuthor:    payload.Issue.User.Login,
					IssueURL:       payload.Issue.HTMLURL,
					CommentID:      payload.Comment.ID,
					CommentBody:    payload.Comment.Body,
					CommentAuthor:  payload.Comment.User.Login,
					Owner:          owner,
					Repo:           repo,
				}

				return enqueueJob(ctx, logger, submitter, queue.Job{
					Name: "issue_comment_response",
					Run: func(jobCtx context.Context) error {
						logger.Printf("job=issue_comment_response repo=%s issue=%d",
							payload.Repository.FullName, payload.Issue.Number)
						return issueService.OnIssueComment(jobCtx, evt)
					},
				})
			}
		}

		return enqueueJob(ctx, logger, submitter, queue.Job{
			Name: "issue_comment",
			Run: func(_ context.Context) error {
				logger.Printf("job=issue_comment action=%s repo=%s issue=%d comment_id=%d",
					payload.Action, payload.Repository.FullName,
					payload.Issue.Number, payload.Comment.ID)
				return nil
			},
		})
	}
}
