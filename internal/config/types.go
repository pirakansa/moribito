package config

import "time"

// Config holds all configuration values for the GitHub App server.
type Config struct {
	Addr              string // HTTP server listen address
	AppID             int64  // GitHub App ID
	InstallationID    int64  // GitHub App Installation ID
	PrivateKeyPath    string // Path to RSA private key (PEM format)
	WebhookSecret     string // Webhook signature secret (optional)
	GitHubAPIBaseURL  string // GitHub API base URL (for Enterprise Server)
	GitHubWebhookPath string // Webhook endpoint path

	// OpenCode server configuration
	OpenCodeHost        string        // OpenCode server hostname (default: 127.0.0.1)
	OpenCodePort        int           // OpenCode server port (default: 4096)
	OpenCodeLongTimeout time.Duration // Timeout for long-running OpenCode requests

	// Queue configuration
	QueueWorkers int // Number of worker goroutines for background jobs
	QueueBuffer  int // Queue buffer size

	// PR open response configuration
	PROpenTemplatePath  string // PR open prompt template file path
	PROpenModel         string // OpenCode model for PR open response (optional)
	PROpenMaxDiffLength int    // Max diff length for PR open prompt (0 disables diff)

	// PR comment response configuration
	PRCommentTemplatePath  string // PR comment prompt template file path
	PRCommentModel         string // OpenCode model for PR comment response (optional)
	PRCommentMaxDiffLength int    // Max diff length for PR comment prompt (0 disables diff)
	PRCommentTriggerPrefix string // Comment prefix to trigger PR comment response (default: @moribito)

	// Issue response configuration
	IssueResponseTemplatePath string // Issue response prompt template file path
	IssueResponseModel        string // OpenCode model for issue response (optional)
	IssueTriggerPrefix        string // Comment prefix to trigger AI response (default: @moribito)
}
