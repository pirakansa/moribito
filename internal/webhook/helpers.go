package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/pirakansa/moribito/internal/queue"
)

// Submitter enqueues background jobs for asynchronous processing.
type Submitter interface {
	Enqueue(ctx context.Context, job queue.Job) error
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

// parseRepoFullName splits "owner/repo" into owner and repo parts.
func parseRepoFullName(fullName string) (owner, repo string, err error) {
	for i := 0; i < len(fullName); i++ {
		if fullName[i] == '/' {
			return fullName[:i], fullName[i+1:], nil
		}
	}
	return "", "", fmt.Errorf("invalid repo full name: %s", fullName)
}
