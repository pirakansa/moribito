package queue

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestQueueRunsJob(t *testing.T) {
	q := New(1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var ran atomic.Bool
	q.Start(ctx, 1)
	err := q.Enqueue(ctx, Job{
		Name: "test",
		Run: func(_ context.Context) error {
			ran.Store(true)
			return nil
		},
	})
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	waitUntil(t, time.Second, func() bool { return ran.Load() })
	q.Close()
	q.Wait()
}

func TestQueueEnqueueAfterClose(t *testing.T) {
	q := New(1)
	q.Close()
	if err := q.Enqueue(context.Background(), Job{Name: "x"}); err == nil {
		t.Fatalf("expected error after close")
	}
}

func waitUntil(t *testing.T, timeout time.Duration, fn func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if fn() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for condition")
}
