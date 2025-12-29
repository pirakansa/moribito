// Package webhook provides event handlers for GitHub webhook events.
// Each handler parses the event payload and enqueues a background job
// for processing.
package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/pirakansa/moribito/internal/queue"
	"github.com/pirakansa/moribito/internal/review"
)

// Submitter enqueues background jobs for asynchronous processing.
type Submitter interface {
	Enqueue(ctx context.Context, job queue.Job) error
}

// Payload types for GitHub webhook events.
// These structs capture only the fields needed for routing and logging.
// Extend them as needed when implementing actual business logic.

type (
	// installationPayload represents GitHub App installation events.
	installationPayload struct {
		Action       string `json:"action"`
		Installation struct {
			ID int64 `json:"id"`
		} `json:"installation"`
	}

	// pullRequestPayload represents pull request events.
	pullRequestPayload struct {
		Action       string `json:"action"`
		Installation struct {
			ID int64 `json:"id"`
		} `json:"installation"`
		PullRequest struct {
			Number int `json:"number"`
		} `json:"pull_request"`
		Repository repositoryInfo `json:"repository"`
	}

	// issueCommentPayload represents issue comment events.
	issueCommentPayload struct {
		Action     string         `json:"action"`
		Issue      issueInfo      `json:"issue"`
		Comment    commentInfo    `json:"comment"`
		Repository repositoryInfo `json:"repository"`
	}

	// checkRunPayload represents check run events.
	checkRunPayload struct {
		Action     string         `json:"action"`
		CheckRun   checkRunInfo   `json:"check_run"`
		Repository repositoryInfo `json:"repository"`
	}

	// repositoryInfo is a common struct for repository metadata.
	repositoryInfo struct {
		FullName string `json:"full_name"`
	}

	// issueInfo holds issue metadata.
	issueInfo struct {
		Number int `json:"number"`
	}

	// commentInfo holds comment metadata.
	commentInfo struct {
		ID int64 `json:"id"`
	}

	// checkRunInfo holds check run metadata.
	checkRunInfo struct {
		ID      int64  `json:"id"`
		Name    string `json:"name"`
		HeadSHA string `json:"head_sha"`
	}
)

// HandleInstallation returns a handler for GitHub App installation events.
// Actions include: created, deleted, suspend, unsuspend.
func HandleInstallation(logger *log.Logger, submitter Submitter) Handler {
	return func(ctx context.Context, event, delivery string, body []byte) error {
		payload, err := decodeJSON[installationPayload](body)
		if err != nil {
			return fmt.Errorf("decode installation payload: %w", err)
		}

		logger.Printf("event=%s delivery=%s action=%s installation_id=%d",
			event, delivery, payload.Action, payload.Installation.ID)

		return enqueueJob(ctx, logger, submitter, queue.Job{
			Name: "installation",
			Run: func(_ context.Context) error {
				logger.Printf("job=installation action=%s installation_id=%d",
					payload.Action, payload.Installation.ID)
				return nil
			},
		})
	}
}

// HandleInstallationRepositories returns a handler for repository selection changes.
// Triggered when repositories are added/removed from an installation.
func HandleInstallationRepositories(logger *log.Logger, submitter Submitter) Handler {
	return func(ctx context.Context, event, delivery string, body []byte) error {
		payload, err := decodeJSON[installationPayload](body)
		if err != nil {
			return fmt.Errorf("decode installation_repositories payload: %w", err)
		}

		logger.Printf("event=%s delivery=%s action=%s installation_id=%d",
			event, delivery, payload.Action, payload.Installation.ID)

		return enqueueJob(ctx, logger, submitter, queue.Job{
			Name: "installation_repositories",
			Run: func(_ context.Context) error {
				logger.Printf("job=installation_repositories action=%s installation_id=%d",
					payload.Action, payload.Installation.ID)
				return nil
			},
		})
	}
}

// HandlePullRequest returns a handler for pull request events.
// Actions include: opened, closed, reopened, synchronize, etc.
// When action is "opened", it triggers the review service for automated code review.
func HandlePullRequest(logger *log.Logger, submitter Submitter, reviewer review.Reviewer) Handler {
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

// HandleIssueComment returns a handler for issue comment events.
// Actions include: created, edited, deleted.
func HandleIssueComment(logger *log.Logger, submitter Submitter) Handler {
	return func(ctx context.Context, event, delivery string, body []byte) error {
		payload, err := decodeJSON[issueCommentPayload](body)
		if err != nil {
			return fmt.Errorf("decode issue_comment payload: %w", err)
		}

		logger.Printf("event=%s delivery=%s action=%s repo=%s issue=%d comment_id=%d",
			event, delivery, payload.Action, payload.Repository.FullName,
			payload.Issue.Number, payload.Comment.ID)

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

// HandleCheckRun returns a handler for check run events.
// Actions include: created, completed, rerequested, requested_action.
func HandleCheckRun(logger *log.Logger, submitter Submitter) Handler {
	return func(ctx context.Context, event, delivery string, body []byte) error {
		payload, err := decodeJSON[checkRunPayload](body)
		if err != nil {
			return fmt.Errorf("decode check_run payload: %w", err)
		}

		logger.Printf("event=%s delivery=%s action=%s repo=%s check_run=%d name=%s",
			event, delivery, payload.Action, payload.Repository.FullName,
			payload.CheckRun.ID, payload.CheckRun.Name)

		return enqueueJob(ctx, logger, submitter, queue.Job{
			Name: "check_run",
			Run: func(_ context.Context) error {
				logger.Printf("job=check_run action=%s repo=%s check_run=%d name=%s",
					payload.Action, payload.Repository.FullName,
					payload.CheckRun.ID, payload.CheckRun.Name)
				return nil
			},
		})
	}
}

// decodeJSON is a generic helper to unmarshal JSON payloads.
func decodeJSON[T any](body []byte) (T, error) {
	var payload T
	if err := json.Unmarshal(body, &payload); err != nil {
		return payload, err
	}
	return payload, nil
}

// enqueueJob submits a job to the queue if a submitter is available.
// Returns nil if submitter is nil (useful for testing without a queue).
func enqueueJob(ctx context.Context, logger *log.Logger, submitter Submitter, job queue.Job) error {
	if submitter == nil {
		return nil
	}
	if err := submitter.Enqueue(ctx, job); err != nil {
		logger.Printf("job enqueue failed name=%s err=%v", job.Name, err)
		return nil // Log but don't fail the webhook response
	}
	return nil
}
