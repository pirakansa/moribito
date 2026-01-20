package queue

import (
	"context"
	"testing"
	"time"
)

func TestRepoLimiterWrap(t *testing.T) {
	limiter := NewRepoLimiter(map[string]int{"acme/widgets": 1})
	started := make(chan struct{}, 2)
	release := make(chan struct{})

	job := Job{
		Name:         "test",
		RepoFullName: "acme/widgets",
		Run: func(_ context.Context) error {
			started <- struct{}{}
			<-release
			return nil
		},
	}

	wrapped := limiter.Wrap(job)
	go func() {
		_ = wrapped.Run(context.Background())
	}()

	select {
	case <-started:
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("expected first job to start")
	}

	secondDone := make(chan struct{})
	go func() {
		_ = wrapped.Run(context.Background())
		close(secondDone)
	}()

	select {
	case <-started:
		t.Fatalf("expected second job to be blocked by limiter")
	case <-time.After(100 * time.Millisecond):
	}

	close(release)

	select {
	case <-secondDone:
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("expected second job to complete after release")
	}
}
