package queue

import (
	"context"
	"sync"
)

// RepoLimiter enforces per-repository concurrency limits.
type RepoLimiter struct {
	limits map[string]int
	mu     sync.Mutex
	sems   map[string]chan struct{}
}

// NewRepoLimiter creates a limiter with the provided per-repo limits.
func NewRepoLimiter(limits map[string]int) *RepoLimiter {
	copied := make(map[string]int, len(limits))
	for repo, limit := range limits {
		if limit > 0 {
			copied[repo] = limit
		}
	}
	return &RepoLimiter{
		limits: copied,
		sems:   make(map[string]chan struct{}, len(copied)),
	}
}

// Wrap returns a job that respects per-repo limits when configured.
func (r *RepoLimiter) Wrap(job Job) Job {
	if job.Run == nil || job.RepoFullName == "" {
		return job
	}
	limit := r.limitFor(job.RepoFullName)
	if limit <= 0 {
		return job
	}
	sem := r.semaphore(job.RepoFullName, limit)
	original := job.Run
	job.Run = func(ctx context.Context) error {
		sem <- struct{}{}
		defer func() { <-sem }()
		return original(ctx)
	}
	return job
}

func (r *RepoLimiter) limitFor(repo string) int {
	if r == nil {
		return 0
	}
	return r.limits[repo]
}

func (r *RepoLimiter) semaphore(repo string, limit int) chan struct{} {
	r.mu.Lock()
	defer r.mu.Unlock()
	sem, ok := r.sems[repo]
	if ok {
		return sem
	}
	sem = make(chan struct{}, limit)
	r.sems[repo] = sem
	return sem
}
