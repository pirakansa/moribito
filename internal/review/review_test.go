package review

import (
	"bytes"
	"context"
	"log"
	"strings"
	"testing"

	"github.com/pirakansa/moribito/internal/githubapp"
)

// mockGitHubClient is a test double for githubapp.GitHubClient.
type mockGitHubClient struct {
	reactions []reactionCall
	comments  []commentCall
}

type reactionCall struct {
	owner, repo string
	number      int
	reaction    string
}

type commentCall struct {
	owner, repo string
	number      int
	body        string
}

func (m *mockGitHubClient) AddIssueReaction(_ context.Context, owner, repo string, number int, reaction string) error {
	m.reactions = append(m.reactions, reactionCall{owner, repo, number, reaction})
	return nil
}

func (m *mockGitHubClient) AddIssueComment(_ context.Context, owner, repo string, number int, body string) error {
	m.comments = append(m.comments, commentCall{owner, repo, number, body})
	return nil
}

// mockClientFactory is a test double for ClientFactory.
type mockClientFactory struct {
	client *mockGitHubClient
}

func (f *mockClientFactory) NewClient(_ context.Context, _ int64) (githubapp.GitHubClient, error) {
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

	// Verify acknowledge step: eyes reaction was added
	if len(mockClient.reactions) != 1 {
		t.Fatalf("expected 1 reaction, got %d", len(mockClient.reactions))
	}
	r := mockClient.reactions[0]
	if r.owner != "example" {
		t.Errorf("expected owner 'example', got '%s'", r.owner)
	}
	if r.repo != "repo" {
		t.Errorf("expected repo 'repo', got '%s'", r.repo)
	}
	if r.number != 42 {
		t.Errorf("expected number 42, got %d", r.number)
	}
	if r.reaction != "eyes" {
		t.Errorf("expected reaction 'eyes', got '%s'", r.reaction)
	}

	// Verify logs show the complete flow
	logOutput := buf.String()
	if !strings.Contains(logOutput, "acknowledging PR") {
		t.Errorf("expected log to contain 'acknowledging PR', got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "processing PR") {
		t.Errorf("expected log to contain 'processing PR', got: %s", logOutput)
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
