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

	// Queue configuration
	QueueWorkers int // Number of worker goroutines for background jobs
	QueueBuffer  int // Queue buffer size

	// Repository-specific settings
	Repositories map[string]RepositoryConfig
}

// RepositoryConfig holds per-repository settings.
type RepositoryConfig struct {
	// OpenCode server configuration
	OpenCodeHost        string        // OpenCode server hostname (default: 127.0.0.1)
	OpenCodePort        int           // OpenCode server port (default: 4096)
	OpenCodeLongTimeout time.Duration // Timeout for long-running OpenCode requests

	// Queue configuration
	QueueWorkersLimit int // Max concurrent jobs for this repository (0 disables limit)

	// PR open response configuration
	PROpenTemplatePath  string // PR open prompt template file path
	PROpenModel         string // OpenCode model for PR open response (optional)
	PROpenMaxDiffLength int    // Max diff length for PR open prompt (0 disables diff)
	PROpenConfigured    bool   // True if prOpen section is configured
	PROpenTemplateSet   bool   // True if prOpen.templatePath is set explicitly
	PROpenModelSet      bool   // True if prOpen.model is set explicitly

	// PR comment response configuration
	PRCommentTemplatePath  string // PR comment prompt template file path
	PRCommentModel         string // OpenCode model for PR comment response (optional)
	PRCommentMaxDiffLength int    // Max diff length for PR comment prompt (0 disables diff)
	PRCommentTriggerPrefix string // Comment prefix to trigger PR comment response (default: @moribito)
	PRCommentConfigured    bool   // True if prComment section is configured
	PRCommentTemplateSet   bool   // True if prComment.templatePath is set explicitly
	PRCommentModelSet      bool   // True if prComment.model is set explicitly

	// Issue response configuration
	IssueResponseTemplatePath string // Issue response prompt template file path
	IssueResponseModel        string // OpenCode model for issue response (optional)
	IssueTriggerPrefix        string // Comment prefix to trigger AI response (default: @moribito)
	IssueCommentConfigured    bool   // True if issueComment section is configured
	IssueCommentTemplateSet   bool   // True if issueComment.templatePath is set explicitly
	IssueCommentModelSet      bool   // True if issueComment.model is set explicitly

	// Issue label response configuration
	IssueLabelTemplatePath string   // Issue label prompt template file path (optional)
	IssueLabelModel        string   // OpenCode model for issue label response (optional)
	IssueLabelTriggers     []string // Labels that trigger issue responses
	IssueLabelConfigured   bool     // True if issueLabel section is configured
	IssueLabelTemplateSet  bool     // True if issueLabel.templatePath is set explicitly
	IssueLabelModelSet     bool     // True if issueLabel.model is set explicitly

	// PR label response configuration
	PRLabelTemplatePath  string   // PR label prompt template file path (optional)
	PRLabelModel         string   // OpenCode model for PR label response (optional)
	PRLabelTriggers      []string // Labels that trigger PR responses
	PRLabelMaxDiffLength int      // Max diff length for PR label prompt (0 disables diff)
	PRLabelConfigured    bool     // True if prLabel section is configured
	PRLabelTemplateSet   bool     // True if prLabel.templatePath is set explicitly
	PRLabelModelSet      bool     // True if prLabel.model is set explicitly
}
