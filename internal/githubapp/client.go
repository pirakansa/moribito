package githubapp

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-github/v80/github"
)

// GitHubClient defines the interface for GitHub API operations.
// This interface provides methods needed for PR review automation.
type GitHubClient interface {
	// AddIssueReaction adds a reaction to an issue or pull request.
	AddIssueReaction(ctx context.Context, owner, repo string, number int, reaction string) error
	// AddIssueComment posts a comment on an issue or pull request.
	AddIssueComment(ctx context.Context, owner, repo string, number int, body string) error
}

// Client provides GitHub API operations for the application.
type Client struct {
	gh *github.Client
}

// NewClient creates a GitHub API client authenticated with an installation token.
// Parameters:
//   - baseURL: GitHub API base URL (empty for github.com)
//   - httpClient: HTTP client to use (nil uses default)
//   - installationToken: installation access token for authentication
func NewClient(baseURL string, httpClient *http.Client, installationToken string) (*Client, error) {
	gh, err := newGitHubAppClient(baseURL, httpClient, installationToken)
	if err != nil {
		return nil, fmt.Errorf("create github client: %w", err)
	}
	return &Client{gh: gh}, nil
}

// AddIssueReaction adds a reaction to an issue or pull request.
// Supported reactions: +1, -1, laugh, confused, heart, hooray, rocket, eyes
// Parameters:
//   - ctx: context for cancellation
//   - owner: repository owner
//   - repo: repository name
//   - number: issue or PR number
//   - reaction: reaction content (e.g., "eyes")
func (c *Client) AddIssueReaction(ctx context.Context, owner, repo string, number int, reaction string) error {
	_, _, err := c.gh.Reactions.CreateIssueReaction(ctx, owner, repo, number, reaction)
	if err != nil {
		return fmt.Errorf("create reaction: %w", err)
	}
	return nil
}

// AddIssueComment posts a comment on an issue or pull request.
// Parameters:
//   - ctx: context for cancellation
//   - owner: repository owner
//   - repo: repository name
//   - number: issue or PR number
//   - body: comment body text
func (c *Client) AddIssueComment(ctx context.Context, owner, repo string, number int, body string) error {
	comment := &github.IssueComment{Body: github.Ptr(body)}
	_, _, err := c.gh.Issues.CreateComment(ctx, owner, repo, number, comment)
	if err != nil {
		return fmt.Errorf("create comment: %w", err)
	}
	return nil
}

// ParseRepoFullName splits "owner/repo" into owner and repo parts.
func ParseRepoFullName(fullName string) (owner, repo string, err error) {
	parts := strings.SplitN(fullName, "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repo full name: %s", fullName)
	}
	return parts[0], parts[1], nil
}
