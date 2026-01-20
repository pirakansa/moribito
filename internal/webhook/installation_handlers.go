package webhook

import (
	"context"
	"fmt"
	"log"

	"github.com/pirakansa/moribito/internal/queue"
)

// HandleInstallation returns a handler for GitHub App installation events.
// Actions include: created, deleted, suspend, unsuspend.
func HandleInstallation(logger *log.Logger, submitter Submitter) Handler {
	return func(ctx context.Context, event, delivery string, body []byte) (HandleResult, error) {
		payload, err := decodeJSON[installationPayload](body)
		if err != nil {
			return HandleResult{}, fmt.Errorf("decode installation payload: %w", err)
		}

		logger.Printf("event=%s delivery=%s action=%s installation_id=%d",
			event, delivery, payload.Action, payload.Installation.ID)

		return HandleResult{}, enqueueJob(ctx, logger, submitter, queue.Job{
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
	return func(ctx context.Context, event, delivery string, body []byte) (HandleResult, error) {
		payload, err := decodeJSON[installationPayload](body)
		if err != nil {
			return HandleResult{}, fmt.Errorf("decode installation_repositories payload: %w", err)
		}

		logger.Printf("event=%s delivery=%s action=%s installation_id=%d",
			event, delivery, payload.Action, payload.Installation.ID)

		return HandleResult{}, enqueueJob(ctx, logger, submitter, queue.Job{
			Name: "installation_repositories",
			Run: func(_ context.Context) error {
				logger.Printf("job=installation_repositories action=%s installation_id=%d",
					payload.Action, payload.Installation.ID)
				return nil
			},
		})
	}
}
