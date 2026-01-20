package webhook

import (
	"context"

	"github.com/pirakansa/moribito/internal/queue"
)

type repoLimitedSubmitter struct {
	base    Submitter
	limiter *queue.RepoLimiter
}

// NewRepoLimitedSubmitter wraps a submitter with per-repo concurrency limits.
func NewRepoLimitedSubmitter(base Submitter, limits map[string]int) Submitter {
	if base == nil || len(limits) == 0 {
		return base
	}
	return &repoLimitedSubmitter{
		base:    base,
		limiter: queue.NewRepoLimiter(limits),
	}
}

func (s *repoLimitedSubmitter) Enqueue(ctx context.Context, job queue.Job) error {
	return s.base.Enqueue(ctx, s.limiter.Wrap(job))
}
