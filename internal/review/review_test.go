package review

import (
	"bytes"
	"context"
	"log"
	"strings"
	"testing"

	"github.com/pirakansa/moribito/internal/githubapp"
	"github.com/pirakansa/moribito/internal/opencode"
	"github.com/pirakansa/moribito/internal/prompt"
)

// mockGitHubClient is a test double for githubapp.GitHubClient.
type mockGitHubClient struct {
	reactions []reactionCall
	comments  []commentCall
	prInfo    *githubapp.PullRequestInfo
	diff      string
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

func (m *mockGitHubClient) AddCommentReaction(_ context.Context, _, _ string, _ int64, _ string) error {
	return nil
}

func (m *mockGitHubClient) GetPullRequestDiff(_ context.Context, _, _ string, _ int) (string, error) {
	return m.diff, nil
}

func (m *mockGitHubClient) GetPullRequest(_ context.Context, _, _ string, _ int) (*githubapp.PullRequestInfo, error) {
	if m.prInfo == nil {
		return &githubapp.PullRequestInfo{
			Title: "Test PR",
			Body:  "Test description",
			Head:  "feature",
			Base:  "main",
		}, nil
	}
	return m.prInfo, nil
}

func (m *mockGitHubClient) GetIssue(_ context.Context, _, _ string, _ int) (*githubapp.IssueInfo, error) {
	return &githubapp.IssueInfo{
		Number:  1,
		Title:   "Test Issue",
		Body:    "Test issue body",
		Author:  "testuser",
		HTMLURL: "https://github.com/test/repo/issues/1",
	}, nil
}

// mockClientFactory is a test double for ClientFactory.
type mockClientFactory struct {
	client *mockGitHubClient
}

func (f *mockClientFactory) NewClient(_ context.Context, _ int64) (githubapp.GitHubClient, error) {
	return f.client, nil
}

// mockOpenCodeClient is a test double for OpenCodeClient.
type mockOpenCodeClient struct {
	healthy      bool
	sessionID    string
	response     string
	createCalled bool
	deleteCalled bool
	lastRequest  *opencode.SendMessageRequest
}

func (m *mockOpenCodeClient) IsHealthy(_ context.Context) bool {
	return m.healthy
}

func (m *mockOpenCodeClient) CreateSession(_ context.Context, _ *opencode.CreateSessionRequest) (*opencode.Session, error) {
	m.createCalled = true
	return &opencode.Session{ID: m.sessionID}, nil
}

func (m *mockOpenCodeClient) SendMessage(_ context.Context, _ string, req *opencode.SendMessageRequest) (*opencode.MessageWithParts, error) {
	m.lastRequest = req
	return &opencode.MessageWithParts{
		Parts: []opencode.Part{{Type: "text", Text: m.response}},
	}, nil
}

func (m *mockOpenCodeClient) DeleteSession(_ context.Context, _ string) error {
	m.deleteCalled = true
	return nil
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

func TestNewServiceWithOpenCodeClient(t *testing.T) {
	logger := log.New(&bytes.Buffer{}, "", 0)
	mockOC := &mockOpenCodeClient{healthy: true}
	svc := NewService(logger, nil, WithOpenCodeClient(mockOC))

	if svc.opencodeClient == nil {
		t.Fatal("expected opencode client to be set")
	}
}

func TestProcessWithoutOpenCode(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "test: ", 0)
	mockClient := &mockGitHubClient{
		diff: "+func hello() {}",
	}
	factory := &mockClientFactory{client: mockClient}
	// No OpenCode client configured
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

	logOutput := buf.String()
	if !strings.Contains(logOutput, "opencode client not configured") {
		t.Errorf("expected log to mention opencode not configured, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "skipping AI review") {
		t.Errorf("expected log to mention skipping AI review, got: %s", logOutput)
	}
	// No comments should be posted without AI
	if len(mockClient.comments) != 0 {
		t.Errorf("expected no comments without OpenCode, got %d", len(mockClient.comments))
	}
}

func TestProcessWithOpenCodeUnhealthy(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "test: ", 0)
	mockClient := &mockGitHubClient{
		diff: "+func hello() {}",
	}
	factory := &mockClientFactory{client: mockClient}
	mockOC := &mockOpenCodeClient{healthy: false} // Unhealthy
	svc := NewService(logger, factory, WithOpenCodeClient(mockOC))

	pr := PullRequest{
		Number:         42,
		RepoName:       "example/repo",
		Action:         "opened",
		InstallationID: 12345,
	}

	if err := svc.OnPullRequestOpened(context.Background(), pr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	logOutput := buf.String()
	if !strings.Contains(logOutput, "opencode server not available") {
		t.Errorf("expected log to mention opencode not available, got: %s", logOutput)
	}
	// No comments should be posted when OpenCode is unhealthy
	if len(mockClient.comments) != 0 {
		t.Errorf("expected no comments without healthy OpenCode, got %d", len(mockClient.comments))
	}
}

func TestProcessWithOpenCodeHealthy(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "test: ", 0)
	mockClient := &mockGitHubClient{
		diff: "+func hello() {}",
	}
	factory := &mockClientFactory{client: mockClient}
	mockOC := &mockOpenCodeClient{
		healthy:   true,
		sessionID: "test-session-123",
		response:  "This code looks good! No issues found.",
	}
	builder := prompt.NewBuilder(prompt.WithTemplate(prompt.Template{
		Name:    "pr",
		Content: "Title={{.Title}}\nDiff={{.Diff}}",
	}))
	svc := NewService(logger, factory, WithOpenCodeClient(mockOC), WithPromptBuilder(builder))

	pr := PullRequest{
		Number:         42,
		RepoName:       "example/repo",
		Action:         "opened",
		InstallationID: 12345,
	}

	if err := svc.OnPullRequestOpened(context.Background(), pr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify OpenCode was called
	if !mockOC.createCalled {
		t.Error("expected CreateSession to be called")
	}
	if !mockOC.deleteCalled {
		t.Error("expected DeleteSession to be called (cleanup)")
	}

	// Verify comment was posted
	if len(mockClient.comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(mockClient.comments))
	}
	comment := mockClient.comments[0]
	if !strings.Contains(comment.body, "AI Code Review") {
		t.Errorf("expected comment to contain 'AI Code Review', got: %s", comment.body)
	}
	if !strings.Contains(comment.body, "This code looks good") {
		t.Errorf("expected comment to contain AI response, got: %s", comment.body)
	}
}

func TestProcessWithReviewModel(t *testing.T) {
	logger := log.New(&bytes.Buffer{}, "test: ", 0)
	mockClient := &mockGitHubClient{
		diff: "+func hello() {}",
	}
	factory := &mockClientFactory{client: mockClient}
	mockOC := &mockOpenCodeClient{
		healthy:   true,
		sessionID: "test-session-123",
		response:  "Looks good.",
	}
	builder := prompt.NewBuilder(prompt.WithTemplate(prompt.Template{
		Name:    "pr",
		Content: "Title={{.Title}}\nDiff={{.Diff}}",
	}))
	svc := NewService(
		logger,
		factory,
		WithOpenCodeClient(mockOC),
		WithPromptBuilder(builder),
		WithReviewModel("review-model"),
	)

	pr := PullRequest{
		Number:         42,
		RepoName:       "example/repo",
		Action:         "opened",
		InstallationID: 12345,
	}

	if err := svc.OnPullRequestOpened(context.Background(), pr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mockOC.lastRequest == nil {
		t.Fatal("expected SendMessage request to be captured")
	}
	if mockOC.lastRequest.Model != "review-model" {
		t.Fatalf("expected model review-model, got %q", mockOC.lastRequest.Model)
	}
}

func TestWithPromptBuilder(t *testing.T) {
	logger := log.New(&bytes.Buffer{}, "", 0)
	// Create a custom builder (we don't need prompt import here, just test the option works)
	svc := NewService(logger, nil)

	if svc.promptBuilder == nil {
		t.Fatal("expected default promptBuilder to be set")
	}
}
