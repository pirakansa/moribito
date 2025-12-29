package review

import (
	"bytes"
	"context"
	"log"
	"strings"
	"testing"

	"github.com/pirakansa/moribito/internal/githubapp"
)

// mockGitHubClient is a test double for githubapp.IssueReactor.
type mockGitHubClient struct {
	addReactionCalled bool
	lastOwner         string
	lastRepo          string
	lastNumber        int
	lastReaction      string
}

func (m *mockGitHubClient) AddIssueReaction(_ context.Context, owner, repo string, number int, reaction string) error {
	m.addReactionCalled = true
	m.lastOwner = owner
	m.lastRepo = repo
	m.lastNumber = number
	m.lastReaction = reaction
	return nil
}

// mockClientFactory is a test double for ClientFactory.
type mockClientFactory struct {
	client *mockGitHubClient
}

func (f *mockClientFactory) NewClient(_ context.Context, _ int64) (githubapp.IssueReactor, error) {
	return f.client, nil
}

func TestOnPullRequestOpened(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "test: ", 0)
	svc := NewService(logger, nil)

	pr := PullRequest{
		Number:   42,
		RepoName: "example/repo",
		Action:   "opened",
	}

	if err := svc.OnPullRequestOpened(context.Background(), pr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	logOutput := buf.String()
	if !strings.Contains(logOutput, "pull request opened") {
		t.Errorf("expected log to contain 'pull request opened', got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "example/repo") {
		t.Errorf("expected log to contain repo name, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "42") {
		t.Errorf("expected log to contain PR number, got: %s", logOutput)
	}
}

func TestOnPullRequestOpenedWithClient(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "test: ", 0)
	mockClient := &mockGitHubClient{}
	factory := &mockClientFactory{client: mockClient}
	svc := NewService(logger, factory)

	pr := PullRequest{
		Number:         42,
		RepoName:       "example/repo",
		Action:         "opened",
		InstallationID: 12345,
	}

	if err := svc.OnPullRequestOpened(context.Background(), pr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !mockClient.addReactionCalled {
		t.Error("expected AddIssueReaction to be called")
	}
	if mockClient.lastOwner != "example" {
		t.Errorf("expected owner 'example', got '%s'", mockClient.lastOwner)
	}
	if mockClient.lastRepo != "repo" {
		t.Errorf("expected repo 'repo', got '%s'", mockClient.lastRepo)
	}
	if mockClient.lastNumber != 42 {
		t.Errorf("expected number 42, got %d", mockClient.lastNumber)
	}
	if mockClient.lastReaction != "eyes" {
		t.Errorf("expected reaction 'eyes', got '%s'", mockClient.lastReaction)
	}
}

func TestNewService(t *testing.T) {
	logger := log.New(&bytes.Buffer{}, "", 0)
	svc := NewService(logger, nil)

	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.logger == nil {
		t.Fatal("expected logger to be set")
	}
}
