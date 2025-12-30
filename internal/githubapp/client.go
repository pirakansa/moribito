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
	// AddCommentReaction adds a reaction to a comment.
	AddCommentReaction(ctx context.Context, owner, repo string, commentID int64, reaction string) error
	// GetPullRequestDiff returns the diff of a pull request in unified diff format.
	GetPullRequestDiff(ctx context.Context, owner, repo string, number int) (string, error)
	// GetPullRequest returns pull request details.
	GetPullRequest(ctx context.Context, owner, repo string, number int) (*PullRequestInfo, error)
	// GetIssue returns issue details.
	GetIssue(ctx context.Context, owner, repo string, number int) (*IssueInfo, error)
}

// PullRequestInfo holds essential pull request information.
type PullRequestInfo struct {
	Title   string
	Body    string
	Head    string // Head branch name
	Base    string // Base branch name
	HTMLURL string
}

// IssueInfo holds essential issue information.
type IssueInfo struct {
	Number  int
	Title   string
	Body    string
	Author  string
	HTMLURL string
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

// AddCommentReaction adds a reaction to a comment.
// Supported reactions: +1, -1, laugh, confused, heart, hooray, rocket, eyes
// Parameters:
//   - ctx: context for cancellation
//   - owner: repository owner
//   - repo: repository name
//   - commentID: comment ID
//   - reaction: reaction content (e.g., "eyes")
func (c *Client) AddCommentReaction(ctx context.Context, owner, repo string, commentID int64, reaction string) error {
	_, _, err := c.gh.Reactions.CreateIssueCommentReaction(ctx, owner, repo, commentID, reaction)
	if err != nil {
		return fmt.Errorf("create comment reaction: %w", err)
	}
	return nil
}

// GetPullRequestDiff returns the diff of a pull request in unified diff format.
// Parameters:
//   - ctx: context for cancellation
//   - owner: repository owner
//   - repo: repository name
//   - number: PR number
func (c *Client) GetPullRequestDiff(ctx context.Context, owner, repo string, number int) (string, error) {
	diff, _, err := c.gh.PullRequests.GetRaw(ctx, owner, repo, number, github.RawOptions{Type: github.Diff})
	if err != nil {
		return "", fmt.Errorf("get pull request diff: %w", err)
	}
	return diff, nil
}

// GetPullRequest returns pull request details.
// Parameters:
//   - ctx: context for cancellation
//   - owner: repository owner
//   - repo: repository name
//   - number: PR number
func (c *Client) GetPullRequest(ctx context.Context, owner, repo string, number int) (*PullRequestInfo, error) {
	pr, _, err := c.gh.PullRequests.Get(ctx, owner, repo, number)
	if err != nil {
		return nil, fmt.Errorf("get pull request: %w", err)
	}
	return &PullRequestInfo{
		Title:   pr.GetTitle(),
		Body:    pr.GetBody(),
		Head:    pr.GetHead().GetRef(),
		Base:    pr.GetBase().GetRef(),
		HTMLURL: pr.GetHTMLURL(),
	}, nil
}

// GetIssue returns issue details.
// Parameters:
//   - ctx: context for cancellation
//   - owner: repository owner
//   - repo: repository name
//   - number: issue number
func (c *Client) GetIssue(ctx context.Context, owner, repo string, number int) (*IssueInfo, error) {
	issue, _, err := c.gh.Issues.Get(ctx, owner, repo, number)
	if err != nil {
		return nil, fmt.Errorf("get issue: %w", err)
	}
	return &IssueInfo{
		Number:  issue.GetNumber(),
		Title:   issue.GetTitle(),
		Body:    issue.GetBody(),
		Author:  issue.GetUser().GetLogin(),
		HTMLURL: issue.GetHTMLURL(),
	}, nil
}

// ParseRepoFullName splits "owner/repo" into owner and repo parts.
func ParseRepoFullName(fullName string) (owner, repo string, err error) {
	parts := strings.SplitN(fullName, "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repo full name: %s", fullName)
	}
	return parts[0], parts[1], nil
}
