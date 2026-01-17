package prompt

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

// PRReviewContext holds the data for PR review prompt templates.
type PRReviewContext struct {
	Title        string // PR title
	Body         string // PR description
	Head         string // Head branch name
	Base         string // Base branch name
	URL          string // PR URL
	Diff         string // PR diff content
	Owner        string // Repository owner or organization
	Repo         string // Repository name
	RepoFullName string // Repository full name (owner/repo)
	Number       int    // PR number
}

// IssueContext holds the data for issue response prompt templates.
type IssueContext struct {
	Title         string // Issue title
	Number        int    // Issue number
	Author        string // Issue author
	Body          string // Issue description
	URL           string // Issue URL
	Comment       string // Comment content to respond to
	CommentAuthor string // Comment author
	CommentID     int64  // Comment ID
	Owner         string // Repository owner or organization
	Repo          string // Repository name
	RepoFullName  string // Repository full name (owner/repo)
}

// Builder constructs prompts from templates.
type Builder struct {
	template     Template
	maxDiffLen   int
	customFooter string
}

// BuilderOption configures the Builder.
type BuilderOption func(*Builder)

// WithTemplate sets a specific template.
func WithTemplate(t Template) BuilderOption {
	return func(b *Builder) {
		b.template = t
	}
}

// WithMaxDiffLength sets the maximum diff length.
func WithMaxDiffLength(maxLen int) BuilderOption {
	return func(b *Builder) {
		b.maxDiffLen = maxLen
	}
}

// WithCustomFooter sets a custom footer for comments.
func WithCustomFooter(footer string) BuilderOption {
	return func(b *Builder) {
		b.customFooter = footer
	}
}

// NewBuilder creates a new prompt Builder with the given options.
func NewBuilder(opts ...BuilderOption) *Builder {
	b := &Builder{
		maxDiffLen: 50000, // Default max diff length
	}
	for _, opt := range opts {
		opt(b)
	}
	return b
}

// BuildPRReviewPrompt constructs a PR review prompt from the context.
func (b *Builder) BuildPRReviewPrompt(ctx PRReviewContext) (string, error) {
	if strings.TrimSpace(b.template.Content) == "" {
		return "", fmt.Errorf("template content is empty")
	}

	// Truncate diff if needed
	ctx.Diff = truncateText(ctx.Diff, b.maxDiffLen)

	tmpl, err := template.New(b.template.Name).Parse(b.template.Content)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return buf.String(), nil
}

// FormatReviewComment wraps an AI review in a formatted comment.
func (b *Builder) FormatReviewComment(review string) string {
	footer := CommentFooterTemplate
	if b.customFooter != "" {
		footer = b.customFooter
	}

	return fmt.Sprintf("## 🤖 AI Code Review\n\n%s\n%s", review, footer)
}

// BuildIssueResponsePrompt constructs an issue response prompt from the context.
func (b *Builder) BuildIssueResponsePrompt(ctx IssueContext) (string, error) {
	if strings.TrimSpace(b.template.Content) == "" {
		return "", fmt.Errorf("template content is empty")
	}

	tmpl, err := template.New(b.template.Name).Parse(b.template.Content)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return buf.String(), nil
}

// FormatIssueResponse wraps an AI response in a formatted comment.
func (b *Builder) FormatIssueResponse(response string) string {
	footer := CommentFooterTemplate
	if b.customFooter != "" {
		footer = b.customFooter
	}

	return fmt.Sprintf("%s\n%s", response, footer)
}

// truncateText limits text length with a truncation marker.
func truncateText(text string, maxLen int) string {
	if maxLen == 0 {
		return ""
	}
	if maxLen < 0 || len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "\n... (truncated)"
}

// QuickBuildPRReview is a convenience function for building a PR review prompt
// with default settings.
func QuickBuildPRReview(title, body, head, base, url, diff string) (string, error) {
	return NewBuilder().BuildPRReviewPrompt(PRReviewContext{
		Title: title,
		Body:  body,
		Head:  head,
		Base:  base,
		URL:   url,
		Diff:  diff,
	})
}
