package issue

import (
	"context"

	"github.com/pirakansa/moribito/internal/githubapp"
	"github.com/pirakansa/moribito/internal/opencode"
)

const (
	issueReactionEyes     = "eyes"
	issueReactionThumbsUp = "+1"
	issueReactionConfused = "confused"
)

// IssueClient defines the GitHub API operations needed for issue handling.
type IssueClient interface {
	AddIssueReaction(ctx context.Context, owner, repo string, number int, reaction string) error
	AddIssueComment(ctx context.Context, owner, repo string, number int, body string) error
	AddCommentReaction(ctx context.Context, owner, repo string, commentID int64, reaction string) error
	GetIssue(ctx context.Context, owner, repo string, number int) (*githubapp.IssueInfo, error)
}

// ClientFactory creates authenticated GitHub clients per installation.
type ClientFactory interface {
	NewClient(ctx context.Context, installationID int64) (githubapp.GitHubClient, error)
}

// OpenCodeClient defines the OpenCode operations needed for issue responses.
type OpenCodeClient interface {
	IsHealthy(ctx context.Context) bool
	CreateSession(ctx context.Context, req *opencode.CreateSessionRequest) (*opencode.Session, error)
	SendMessage(ctx context.Context, sessionID string, req *opencode.SendMessageRequest) (*opencode.MessageWithParts, error)
	DeleteSession(ctx context.Context, sessionID string) error
}

// CommentEvent represents an issue comment event.
type CommentEvent struct {
	InstallationID int64
	Owner          string
	Repo           string
	IssueNumber    int
	CommentID      int64
	CommentBody    string
	CommentAuthor  string
	IssueTitle     string
	IssueBody      string
	IssueAuthor    string
	IssueURL       string
}

// LabelEvent represents an issue label event.
type LabelEvent struct {
	InstallationID int64
	Owner          string
	Repo           string
	IssueNumber    int
	IssueTitle     string
	IssueBody      string
	IssueAuthor    string
	IssueURL       string
	LabelName      string
	Labeler        string
}
