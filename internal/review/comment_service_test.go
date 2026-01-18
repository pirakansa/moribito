package review

import (
	"bytes"
	"context"
	"log"
	"testing"

	"github.com/pirakansa/moribito/internal/prompt"
)

func TestOnPullRequestLabeled(t *testing.T) {
	logger := log.New(&bytes.Buffer{}, "test: ", 0)
	mockClient := &mockGitHubClient{
		diff: "+func hello() {}",
	}
	factory := &mockClientFactory{client: mockClient}
	mockOC := &mockOpenCodeClient{
		healthy:   true,
		sessionID: "pr-label-session-1",
		response:  "AI response",
	}
	builder := prompt.NewBuilder(prompt.WithTemplate(prompt.Template{
		Name:    "pr-comment",
		Content: "Title={{.Title}}\nDiff={{.Diff}}",
	}))

	svc := NewPRCommentService(
		logger,
		factory,
		WithCommentOpenCodeClient(mockOC),
		WithCommentPromptBuilder(builder),
		WithCommentLabelModel("pr-label-model"),
		WithCommentLabelTriggers([]string{"needs-review"}),
	)

	event := PRLabelEvent{
		InstallationID: 42,
		Owner:          "example",
		Repo:           "repo",
		Number:         1,
		LabelName:      "needs-review",
		Labeler:        "alice",
	}

	if err := svc.OnPullRequestLabeled(context.Background(), event); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mockOC.lastRequest == nil {
		t.Fatal("expected SendMessage request to be captured")
	}
	if mockOC.lastRequest.Model == nil {
		t.Fatal("expected label model to be set")
	}
	if mockOC.lastRequest.Model.ModelID != "pr-label-model" {
		t.Fatalf("expected model pr-label-model, got %q", mockOC.lastRequest.Model.ModelID)
	}
	if len(mockClient.reactions) != 2 {
		t.Fatalf("expected 2 reactions, got %d", len(mockClient.reactions))
	}
	if len(mockClient.comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(mockClient.comments))
	}
}
