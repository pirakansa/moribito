// Package queue provides an in-memory job queue with worker pool.
// It is designed for local development and processing webhook events
// asynchronously. For production, consider persistent queue backends.
package queue

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
)

// Job represents a background task to be executed by a worker.
type Job struct {
	Name         string                          // Name identifies the job type for logging
	RepoFullName string                          // RepoFullName scopes per-repo limits when set
	Run          func(ctx context.Context) error // Run executes the job logic
}

// Queue is an in-memory job queue with a pool of worker goroutines.
// It provides a simple producer-consumer pattern for background processing.
type Queue struct {
	jobs   chan Job
	wg     sync.WaitGroup
	closed atomic.Bool
}

// New creates a Queue with the specified buffer size.
// If buffer is <= 0, defaults to 100.
func New(buffer int) *Queue {
	if buffer <= 0 {
		buffer = 100
	}
	return &Queue{
		jobs: make(chan Job, buffer),
	}
}

// Start launches the specified number of worker goroutines.
// Workers process jobs until the context is cancelled or the queue is closed.
// If workers is <= 0, defaults to 1.
func (q *Queue) Start(ctx context.Context, workers int) {
	if workers <= 0 {
		workers = 1
	}
	for range workers {
		q.wg.Add(1)
		go func() {
			defer q.wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case job, ok := <-q.jobs:
					if !ok {
						return
					}
					if job.Run != nil {
						_ = job.Run(ctx)
					}
				}
			}
		}()
	}
}

// Enqueue adds a job to the queue.
// Returns an error if the queue is closed or the context is cancelled.
func (q *Queue) Enqueue(ctx context.Context, job Job) error {
	if q.closed.Load() {
		return fmt.Errorf("queue closed")
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case q.jobs <- job:
		return nil
	}
}

// Close stops accepting new jobs and closes the internal channel.
// Safe to call multiple times.
func (q *Queue) Close() {
	if q.closed.CompareAndSwap(false, true) {
		close(q.jobs)
	}
}

// Wait blocks until all workers have finished processing.
func (q *Queue) Wait() {
	q.wg.Wait()
}
