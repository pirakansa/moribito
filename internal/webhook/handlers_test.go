package webhook

import (
	"bytes"
	"context"
	"log"
	"testing"

	"github.com/pirakansa/moribito/internal/review"
)

// mockReviewer is a test double for the review.Reviewer interface.
type mockReviewer struct {
	called bool
	pr     review.PullRequest
}

func (m *mockReviewer) OnPullRequestOpened(_ context.Context, pr review.PullRequest) error {
	m.called = true
	m.pr = pr
	return nil
}

func TestHandleInstallationInvalidJSON(t *testing.T) {
	logger := log.New(&bytes.Buffer{}, "test: ", 0)
	handler := HandleInstallation(logger, nil)
	if err := handler(context.Background(), "installation", "d1", []byte("{")); err == nil {
		t.Fatalf("expected error for invalid json")
	}
}

func TestHandleInstallationRepositories(t *testing.T) {
	logger := log.New(&bytes.Buffer{}, "test: ", 0)
	handler := HandleInstallationRepositories(logger, nil)
	body, err := readFixture("installation_repositories.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	if err := handler(context.Background(), "installation_repositories", "d1", body); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHandleInstallationFixture(t *testing.T) {
	logger := log.New(&bytes.Buffer{}, "test: ", 0)
	handler := HandleInstallation(logger, nil)
	body, err := readFixture("installation.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	if err := handler(context.Background(), "installation", "d1", body); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHandlePullRequest(t *testing.T) {
	logger := log.New(&bytes.Buffer{}, "test: ", 0)
	handler := HandlePullRequest(logger, nil, nil, nil)
	body, err := readFixture("pull_request.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	if err := handler(context.Background(), "pull_request", "d1", body); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHandlePullRequestOpened(t *testing.T) {
	logger := log.New(&bytes.Buffer{}, "test: ", 0)
	reviewer := &mockReviewer{}
	handler := HandlePullRequest(logger, nil, reviewer, nil)
	body, err := readFixture("pull_request.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	if err := handler(context.Background(), "pull_request", "d1", body); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Note: With nil submitter, the job is not enqueued, so reviewer is not called.
	// This test verifies the handler doesn't error with a reviewer provided.
}

func TestHandleIssueComment(t *testing.T) {
	logger := log.New(&bytes.Buffer{}, "test: ", 0)
	handler := HandleIssueComment(logger, nil, nil, nil)
	body, err := readFixture("issue_comment.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	if err := handler(context.Background(), "issue_comment", "d1", body); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHandleIssues(t *testing.T) {
	logger := log.New(&bytes.Buffer{}, "test: ", 0)
	handler := HandleIssues(logger, nil, nil)
	body, err := readFixture("issues_labeled.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	if err := handler(context.Background(), "issues", "d1", body); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHandleCheckRun(t *testing.T) {
	logger := log.New(&bytes.Buffer{}, "test: ", 0)
	handler := HandleCheckRun(logger, nil)
	body, err := readFixture("check_run.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	if err := handler(context.Background(), "check_run", "d1", body); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
