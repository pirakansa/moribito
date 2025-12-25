package queue

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
)

// Job represents a background task.
type Job struct {
	Name string
	Run  func(ctx context.Context) error
}

// Queue is an in-memory job queue with worker goroutines.
type Queue struct {
	jobs   chan Job
	wg     sync.WaitGroup
	closed atomic.Bool
}

// New creates a queue with the given buffer size.
func New(buffer int) *Queue {
	if buffer <= 0 {
		buffer = 100
	}
	return &Queue{
		jobs: make(chan Job, buffer),
	}
}

// Start launches worker goroutines.
func (q *Queue) Start(ctx context.Context, workers int) {
	if workers <= 0 {
		workers = 1
	}
	for i := 0; i < workers; i++ {
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

// Close stops accepting new jobs and closes the channel.
func (q *Queue) Close() {
	if q.closed.CompareAndSwap(false, true) {
		close(q.jobs)
	}
}

// Wait blocks until workers finish.
func (q *Queue) Wait() {
	q.wg.Wait()
}
