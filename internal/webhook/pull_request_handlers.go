package webhook

import (
	"context"
	"fmt"
	"log"

	"github.com/pirakansa/moribito/internal/githubapp"
	"github.com/pirakansa/moribito/internal/queue"
	"github.com/pirakansa/moribito/internal/review"
)

// HandlePullRequest returns a handler for pull request events.
// Actions include: opened, closed, reopened, synchronize, etc.
// When action is "opened", it triggers the review service for automated code review.
func HandlePullRequest(logger *log.Logger, submitter Submitter, reviewer review.Reviewer, prCommenter review.PRCommenter) Handler {
	return func(ctx context.Context, event, delivery string, body []byte) error {
		payload, err := decodeJSON[pullRequestPayload](body)
		if err != nil {
			return fmt.Errorf("decode pull_request payload: %w", err)
		}

		logger.Printf("event=%s delivery=%s action=%s repo=%s number=%d",
			event, delivery, payload.Action, payload.Repository.FullName, payload.PullRequest.Number)

		// Handle PR opened event for automated review
		if payload.Action == "opened" && reviewer != nil {
			pr := review.PullRequest{
				Number:         payload.PullRequest.Number,
				RepoName:       payload.Repository.FullName,
				Action:         payload.Action,
				InstallationID: payload.Installation.ID,
			}
			return enqueueJob(ctx, logger, submitter, queue.Job{
				Name: "pull_request_opened",
				Run: func(jobCtx context.Context) error {
					logger.Printf("job=pull_request_opened repo=%s number=%d",
						payload.Repository.FullName, payload.PullRequest.Number)
					return reviewer.OnPullRequestOpened(jobCtx, pr)
				},
			})
		}

		if payload.Action == "labeled" && prCommenter != nil {
			owner, repo, perr := githubapp.ParseRepoFullName(payload.Repository.FullName)
			if perr != nil {
				return fmt.Errorf("parse repo name: %w", perr)
			}

			evt := review.PRLabelEvent{
				InstallationID: payload.Installation.ID,
				Owner:          owner,
				Repo:           repo,
				Number:         payload.PullRequest.Number,
				LabelName:      payload.Label.Name,
				Labeler:        payload.Sender.Login,
			}

			return enqueueJob(ctx, logger, submitter, queue.Job{
				Name: "pull_request_labeled_response",
				Run: func(jobCtx context.Context) error {
					logger.Printf("job=pull_request_labeled_response repo=%s number=%d label=%s",
						payload.Repository.FullName, payload.PullRequest.Number, payload.Label.Name)
					return prCommenter.OnPullRequestLabeled(jobCtx, evt)
				},
			})
		}

		return enqueueJob(ctx, logger, submitter, queue.Job{
			Name: "pull_request",
			Run: func(_ context.Context) error {
				logger.Printf("job=pull_request action=%s repo=%s number=%d",
					payload.Action, payload.Repository.FullName, payload.PullRequest.Number)
				return nil
			},
		})
	}
}
