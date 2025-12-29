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
	if b.template.Name != DefaultPRReviewTemplate.Name {
		t.Errorf("expected default template, got %s", b.template.Name)
	}
	if b.maxDiffLen != 50000 {
		t.Errorf("expected default maxDiffLen 50000, got %d", b.maxDiffLen)
	}
}

func TestBuilderWithOptions(t *testing.T) {
	t.Run("WithTemplate", func(t *testing.T) {
		b := NewBuilder(WithTemplate(ConcisePRReviewTemplate))
		if b.template.Name != "pr-review-concise" {
			t.Errorf("expected concise template, got %s", b.template.Name)
		}
	})

	t.Run("WithTemplateName", func(t *testing.T) {
		b := NewBuilder(WithTemplateName("pr-review-ja"))
		if b.template.Name != "pr-review-ja" {
			t.Errorf("expected ja template, got %s", b.template.Name)
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
	b := NewBuilder()
	ctx := PRReviewContext{
		Title: "Add new feature",
		Body:  "This PR adds an amazing feature",
		Head:  "feature-branch",
		Base:  "main",
		URL:   "https://github.com/org/repo/pull/123",
		Diff:  "+func newFunc() {}",
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
	}

	for _, check := range checks {
		if !strings.Contains(prompt, check) {
			t.Errorf("prompt should contain %q", check)
		}
	}
}

func TestBuildPRReviewPromptTruncation(t *testing.T) {
	b := NewBuilder(WithMaxDiffLength(20))
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
			want:   "any text",
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
	if err != nil {
		t.Fatalf("QuickBuildPRReview failed: %v", err)
	}
	if !strings.Contains(prompt, "Title") {
		t.Error("prompt should contain title")
	}
}

func TestQuickBuildPRReviewWithTemplate(t *testing.T) {
	prompt, err := QuickBuildPRReviewWithTemplate("pr-review-ja", "タイトル", "本文", "head", "base", "url", "diff")
	if err != nil {
		t.Fatalf("QuickBuildPRReviewWithTemplate failed: %v", err)
	}
	if !strings.Contains(prompt, "タイトル") {
		t.Error("prompt should contain Japanese title")
	}
	if !strings.Contains(prompt, "レビュー") {
		t.Error("prompt should be in Japanese")
	}
}

func TestAllBuiltinTemplatesAreValid(t *testing.T) {
	ctx := PRReviewContext{
		Title: "Test PR",
		Body:  "Test description",
		Head:  "feature",
		Base:  "main",
		URL:   "https://example.com/pr/1",
		Diff:  "+test",
	}

	for _, tmpl := range BuiltinTemplates() {
		t.Run(tmpl.Name, func(t *testing.T) {
			b := NewBuilder(WithTemplate(tmpl))
			prompt, err := b.BuildPRReviewPrompt(ctx)
			if err != nil {
				t.Errorf("template %s failed to build: %v", tmpl.Name, err)
			}
			if prompt == "" {
				t.Errorf("template %s produced empty prompt", tmpl.Name)
			}
		})
	}
}
