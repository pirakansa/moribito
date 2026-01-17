package prompt

import (
	"strings"
	"testing"
)

func TestNewBuilder(t *testing.T) {
	b := NewBuilder()
	if b == nil {
		t.Fatal("expected non-nil builder")
	}
	if b.template.Content != "" {
		t.Errorf("expected empty template content by default")
	}
	if b.maxDiffLen != 50000 {
		t.Errorf("expected default maxDiffLen 50000, got %d", b.maxDiffLen)
	}
}

func TestBuilderWithOptions(t *testing.T) {
	t.Run("WithTemplate", func(t *testing.T) {
		tmpl := Template{Name: "pr", Content: "Title: {{.Title}}"}
		b := NewBuilder(WithTemplate(tmpl))
		if b.template.Name != "pr" {
			t.Errorf("expected pr template, got %s", b.template.Name)
		}
	})

	t.Run("WithMaxDiffLength", func(t *testing.T) {
		b := NewBuilder(WithMaxDiffLength(1000))
		if b.maxDiffLen != 1000 {
			t.Errorf("expected maxDiffLen 1000, got %d", b.maxDiffLen)
		}
	})

	t.Run("WithCustomFooter", func(t *testing.T) {
		customFooter := "\n---\nCustom footer"
		b := NewBuilder(WithCustomFooter(customFooter))
		if b.customFooter != customFooter {
			t.Errorf("expected custom footer, got %s", b.customFooter)
		}
	})
}

func TestBuildPRReviewPrompt(t *testing.T) {
	tmpl := Template{Name: "pr", Content: "Title={{.Title}}\nBody={{.Body}}\nHead={{.Head}}\nBase={{.Base}}\nDiff={{.Diff}}\nOwner={{.Owner}}\nRepo={{.Repo}}\nRepoFull={{.RepoFullName}}\nNumber={{.Number}}"}
	b := NewBuilder(WithTemplate(tmpl))
	ctx := PRReviewContext{
		Title:        "Add new feature",
		Body:         "This PR adds an amazing feature",
		Head:         "feature-branch",
		Base:         "main",
		URL:          "https://github.com/org/repo/pull/123",
		Diff:         "+func newFunc() {}",
		Owner:        "org",
		Repo:         "repo",
		RepoFullName: "org/repo",
		Number:       123,
	}

	prompt, err := b.BuildPRReviewPrompt(ctx)
	if err != nil {
		t.Fatalf("BuildPRReviewPrompt failed: %v", err)
	}

	// Check that all context values are present
	checks := []string{
		"Add new feature",
		"This PR adds an amazing feature",
		"feature-branch",
		"main",
		"+func newFunc() {}",
		"org",
		"repo",
		"org/repo",
		"123",
	}

	for _, check := range checks {
		if !strings.Contains(prompt, check) {
			t.Errorf("prompt should contain %q", check)
		}
	}
}

func TestBuildIssueResponsePrompt(t *testing.T) {
	tmpl := Template{Name: "issue", Content: "Title={{.Title}}\nNumber=#{{.Number}}\nAuthor=@{{.Author}}\nBody={{.Body}}\nComment={{.Comment}}\nCommentAuthor=@{{.CommentAuthor}}\nOwner={{.Owner}}\nRepo={{.Repo}}\nRepoFull={{.RepoFullName}}"}
	b := NewBuilder(WithTemplate(tmpl))
	ctx := IssueContext{
		Title:         "Bug Report",
		Number:        42,
		Author:        "reporter",
		Body:          "Something is broken",
		URL:           "https://github.com/org/repo/issues/42",
		Comment:       "Can you help me fix this?",
		CommentAuthor: "helper",
		CommentID:     123,
		Owner:         "org",
		Repo:          "repo",
		RepoFullName:  "org/repo",
	}

	prompt, err := b.BuildIssueResponsePrompt(ctx)
	if err != nil {
		t.Fatalf("BuildIssueResponsePrompt failed: %v", err)
	}

	checks := []string{
		"Bug Report",
		"#42",
		"@reporter",
		"Something is broken",
		"Can you help me fix this?",
		"@helper",
		"org",
		"repo",
		"org/repo",
	}

	for _, check := range checks {
		if !strings.Contains(prompt, check) {
			t.Errorf("prompt should contain %q", check)
		}
	}
}

func TestBuildPRReviewPromptTruncation(t *testing.T) {
	tmpl := Template{Name: "pr", Content: "Diff={{.Diff}}"}
	b := NewBuilder(WithTemplate(tmpl), WithMaxDiffLength(20))
	ctx := PRReviewContext{
		Title: "Test",
		Body:  "Test body",
		Head:  "head",
		Base:  "base",
		Diff:  "This is a very long diff that should be truncated",
	}

	prompt, err := b.BuildPRReviewPrompt(ctx)
	if err != nil {
		t.Fatalf("BuildPRReviewPrompt failed: %v", err)
	}

	if !strings.Contains(prompt, "(truncated)") {
		t.Error("long diff should be truncated")
	}
}

func TestFormatReviewComment(t *testing.T) {
	b := NewBuilder()
	review := "This code looks good!"

	comment := b.FormatReviewComment(review)

	if !strings.Contains(comment, "AI Code Review") {
		t.Error("comment should contain header")
	}
	if !strings.Contains(comment, "This code looks good!") {
		t.Error("comment should contain review text")
	}
	if !strings.Contains(comment, "M.O.R.I.B.I.T.O.") {
		t.Error("comment should contain default footer")
	}
}

func TestFormatReviewCommentCustomFooter(t *testing.T) {
	customFooter := "\n---\nPowered by Custom AI"
	b := NewBuilder(WithCustomFooter(customFooter))
	review := "LGTM"

	comment := b.FormatReviewComment(review)

	if !strings.Contains(comment, "Custom AI") {
		t.Error("comment should contain custom footer")
	}
	if strings.Contains(comment, "M.O.R.I.B.I.T.O.") {
		t.Error("comment should not contain default footer when custom is set")
	}
}

func TestTruncateText(t *testing.T) {
	tests := []struct {
		name   string
		text   string
		maxLen int
		want   string
	}{
		{
			name:   "short text",
			text:   "short",
			maxLen: 100,
			want:   "short",
		},
		{
			name:   "exact length",
			text:   "exact",
			maxLen: 5,
			want:   "exact",
		},
		{
			name:   "truncated",
			text:   "this is long",
			maxLen: 7,
			want:   "this is\n... (truncated)",
		},
		{
			name:   "zero maxLen",
			text:   "any text",
			maxLen: 0,
			want:   "",
		},
		{
			name:   "negative maxLen",
			text:   "any text",
			maxLen: -1,
			want:   "any text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateText(tt.text, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncateText(%q, %d) = %q, want %q", tt.text, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestQuickBuildPRReview(t *testing.T) {
	prompt, err := QuickBuildPRReview("Title", "Body", "head", "base", "url", "diff")
	if err == nil {
		t.Fatalf("expected error for missing template, got prompt: %s", prompt)
	}
}
