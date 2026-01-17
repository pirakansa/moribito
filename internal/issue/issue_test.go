package issue

import (
	"bytes"
	"context"
	"log"
	"testing"

	"github.com/pirakansa/moribito/internal/githubapp"
	"github.com/pirakansa/moribito/internal/opencode"
	"github.com/pirakansa/moribito/internal/prompt"
)

type mockGitHubClient struct {
	reactions []string
	comments  []string
}

func (m *mockGitHubClient) AddIssueReaction(_ context.Context, _, _ string, _ int, reaction string) error {
	m.reactions = append(m.reactions, reaction)
	return nil
}

func (m *mockGitHubClient) AddIssueComment(_ context.Context, _, _ string, _ int, body string) error {
	m.comments = append(m.comments, body)
	return nil
}

func (m *mockGitHubClient) AddCommentReaction(_ context.Context, _, _ string, _ int64, reaction string) error {
	m.reactions = append(m.reactions, reaction)
	return nil
}

func (m *mockGitHubClient) GetPullRequestDiff(_ context.Context, _, _ string, _ int) (string, error) {
	return "", nil
}

func (m *mockGitHubClient) GetPullRequest(_ context.Context, _, _ string, _ int) (*githubapp.PullRequestInfo, error) {
	return &githubapp.PullRequestInfo{}, nil
}

func (m *mockGitHubClient) GetIssue(_ context.Context, _, _ string, _ int) (*githubapp.IssueInfo, error) {
	return &githubapp.IssueInfo{}, nil
}

type mockClientFactory struct {
	client *mockGitHubClient
}

func (f *mockClientFactory) NewClient(_ context.Context, _ int64) (githubapp.GitHubClient, error) {
	return f.client, nil
}

type mockOpenCodeClient struct {
	healthy     bool
	sessionID   string
	response    string
	lastRequest *opencode.SendMessageRequest
}

func (m *mockOpenCodeClient) IsHealthy(_ context.Context) bool {
	return m.healthy
}

func (m *mockOpenCodeClient) CreateSession(_ context.Context, _ *opencode.CreateSessionRequest) (*opencode.Session, error) {
	return &opencode.Session{ID: m.sessionID}, nil
}

func (m *mockOpenCodeClient) SendMessage(_ context.Context, _ string, req *opencode.SendMessageRequest) (*opencode.MessageWithParts, error) {
	m.lastRequest = req
	return &opencode.MessageWithParts{
		Parts: []opencode.Part{{Type: "text", Text: m.response}},
	}, nil
}

func (m *mockOpenCodeClient) DeleteSession(_ context.Context, _ string) error {
	return nil
}

func TestOnIssueCommentWithResponseModel(t *testing.T) {
	logger := log.New(&bytes.Buffer{}, "test: ", 0)
	mockClient := &mockGitHubClient{}
	factory := &mockClientFactory{client: mockClient}
	mockOC := &mockOpenCodeClient{
		healthy:   true,
		sessionID: "issue-session-1",
		response:  "AI response",
	}
	builder := prompt.NewBuilder(prompt.WithTemplate(prompt.Template{
		Name:    "issue",
		Content: "Title={{.Title}}\nComment={{.Comment}}",
	}))

	svc := NewService(
		logger,
		factory,
		WithOpenCodeClient(mockOC),
		WithPromptBuilder(builder),
		WithResponseModel("issue-model"),
	)

	event := CommentEvent{
		InstallationID: 42,
		Owner:          "example",
		Repo:           "repo",
		IssueNumber:    1,
		CommentID:      10,
		CommentBody:    "@moribito please help",
		CommentAuthor:  "alice",
		IssueTitle:     "Issue title",
		IssueBody:      "Issue body",
		IssueAuthor:    "bob",
		IssueURL:       "https://github.com/example/repo/issues/1",
	}

	if err := svc.OnIssueComment(context.Background(), event); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mockOC.lastRequest == nil {
		t.Fatal("expected SendMessage request to be captured")
	}
	if mockOC.lastRequest.Model != "issue-model" {
		t.Fatalf("expected model issue-model, got %q", mockOC.lastRequest.Model)
	}
	if len(mockClient.comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(mockClient.comments))
	}
}
