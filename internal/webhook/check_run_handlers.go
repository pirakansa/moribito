package webhook

import (
	"context"
	"fmt"
	"log"

	"github.com/pirakansa/moribito/internal/queue"
)

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
