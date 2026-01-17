// Package config provides configuration loading and validation for the GitHub App.
// Configuration is read from a JSON file.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

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

const (
	defaultAddr                = ":8080"
	defaultGitHubAPIBaseURL    = "https://api.github.com"
	defaultWebhookPath         = "/webhook"
	defaultOpenCodeHost        = "127.0.0.1"
	defaultOpenCodePort        = 4096
	defaultOpenCodeModel       = "opencode/big-pickle"
	defaultOpenCodeLongTimeout = 10 * time.Minute
	defaultPROpenMaxDiffLen    = 50000
	defaultPRCommentMaxDiffLen = 50000
	defaultIssueTriggerPrefix  = "@moribito"
	defaultQueueWorkers        = 2
	defaultQueueBuffer         = 100
)

// Load reads configuration from a JSON file specified by MORIBITO_CONFIG_PATH.
func Load() (Config, error) {
	path := strings.TrimSpace(os.Getenv("MORIBITO_CONFIG_PATH"))
	if path == "" {
		return Config{}, fmt.Errorf("MORIBITO_CONFIG_PATH is required")
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config file: %w", err)
	}

	var fileCfg fileConfig
	if err := json.Unmarshal(raw, &fileCfg); err != nil {
		return Config{}, fmt.Errorf("parse config file: %w", err)
	}

	cfg := Config{
		Addr:                   defaultAddr,
		GitHubAPIBaseURL:       defaultGitHubAPIBaseURL,
		GitHubWebhookPath:      defaultWebhookPath,
		OpenCodeHost:           defaultOpenCodeHost,
		OpenCodePort:           defaultOpenCodePort,
		OpenCodeLongTimeout:    defaultOpenCodeLongTimeout,
		QueueWorkers:           defaultQueueWorkers,
		QueueBuffer:            defaultQueueBuffer,
		PROpenModel:            defaultOpenCodeModel,
		PROpenMaxDiffLength:    defaultPROpenMaxDiffLen,
		PRCommentModel:         defaultOpenCodeModel,
		PRCommentMaxDiffLength: defaultPRCommentMaxDiffLen,
		PRCommentTriggerPrefix: defaultIssueTriggerPrefix,
		IssueResponseModel:     defaultOpenCodeModel,
		IssueTriggerPrefix:     defaultIssueTriggerPrefix,
	}

	if fileCfg.Server.Addr != "" {
		cfg.Addr = fileCfg.Server.Addr
	}
	if fileCfg.Server.WebhookPath != "" {
		cfg.GitHubWebhookPath = fileCfg.Server.WebhookPath
	}
	if fileCfg.GitHub.AppID != 0 {
		cfg.AppID = fileCfg.GitHub.AppID
	}
	if fileCfg.GitHub.InstallationID != 0 {
		cfg.InstallationID = fileCfg.GitHub.InstallationID
	}
	if fileCfg.GitHub.PrivateKeyPath != "" {
		cfg.PrivateKeyPath = fileCfg.GitHub.PrivateKeyPath
	}
	if fileCfg.GitHub.WebhookSecret != "" {
		cfg.WebhookSecret = fileCfg.GitHub.WebhookSecret
	}
	if fileCfg.GitHub.APIBaseURL != "" {
		cfg.GitHubAPIBaseURL = fileCfg.GitHub.APIBaseURL
	}
	if fileCfg.OpenCode.Host != "" {
		cfg.OpenCodeHost = fileCfg.OpenCode.Host
	}
	if fileCfg.OpenCode.Port != nil {
		cfg.OpenCodePort = *fileCfg.OpenCode.Port
	}
	if fileCfg.OpenCode.LongTimeoutSeconds != nil {
		cfg.OpenCodeLongTimeout = time.Duration(*fileCfg.OpenCode.LongTimeoutSeconds) * time.Second
	}
	if fileCfg.Queue.Workers != nil {
		cfg.QueueWorkers = *fileCfg.Queue.Workers
	}
	if fileCfg.Queue.Buffer != nil {
		cfg.QueueBuffer = *fileCfg.Queue.Buffer
	}
	if fileCfg.PROpen.TemplatePath != "" {
		cfg.PROpenTemplatePath = fileCfg.PROpen.TemplatePath
	}
	if fileCfg.PROpen.Model != "" {
		cfg.PROpenModel = fileCfg.PROpen.Model
	}
	if fileCfg.PROpen.MaxDiffLength != nil {
		cfg.PROpenMaxDiffLength = *fileCfg.PROpen.MaxDiffLength
	}
	if fileCfg.PRComment.TemplatePath != "" {
		cfg.PRCommentTemplatePath = fileCfg.PRComment.TemplatePath
	}
	if fileCfg.PRComment.Model != "" {
		cfg.PRCommentModel = fileCfg.PRComment.Model
	}
	if fileCfg.PRComment.MaxDiffLength != nil {
		cfg.PRCommentMaxDiffLength = *fileCfg.PRComment.MaxDiffLength
	}
	if fileCfg.PRComment.TriggerPrefix != "" {
		cfg.PRCommentTriggerPrefix = fileCfg.PRComment.TriggerPrefix
	}
	if fileCfg.IssueComment.TemplatePath != "" {
		cfg.IssueResponseTemplatePath = fileCfg.IssueComment.TemplatePath
	}
	if fileCfg.IssueComment.ResponseModel != "" {
		cfg.IssueResponseModel = fileCfg.IssueComment.ResponseModel
	}
	if fileCfg.IssueComment.TriggerPrefix != "" {
		cfg.IssueTriggerPrefix = fileCfg.IssueComment.TriggerPrefix
	}

	return cfg, nil
}

// ValidateForToken ensures the minimum config for fetching installation tokens.
func (c Config) ValidateForToken() error {
	if c.AppID == 0 {
		return fmt.Errorf("github.appID is required")
	}
	if c.InstallationID == 0 {
		return fmt.Errorf("github.installationID is required")
	}
	if c.PrivateKeyPath == "" {
		return fmt.Errorf("github.privateKeyPath is required")
	}
	return nil
}

// ValidateForWebhook ensures the minimum config for running the webhook server.
func (c Config) ValidateForWebhook() error {
	if strings.TrimSpace(c.Addr) == "" {
		return fmt.Errorf("server.addr is required")
	}
	if strings.TrimSpace(c.GitHubWebhookPath) == "" {
		return fmt.Errorf("server.webhookPath is required")
	}
	if strings.TrimSpace(c.PROpenTemplatePath) == "" {
		return fmt.Errorf("prOpen.templatePath is required")
	}
	if strings.TrimSpace(c.IssueResponseTemplatePath) == "" {
		return fmt.Errorf("issueComment.templatePath is required")
	}
	return nil
}

type fileConfig struct {
	Server struct {
		Addr        string `json:"addr"`
		WebhookPath string `json:"webhookPath"`
	} `json:"server"`
	GitHub struct {
		AppID          int64  `json:"appID"`
		InstallationID int64  `json:"installationID"`
		PrivateKeyPath string `json:"privateKeyPath"`
		WebhookSecret  string `json:"webhookSecret"`
		APIBaseURL     string `json:"apiBaseURL"`
	} `json:"github"`
	OpenCode struct {
		Host               string `json:"host"`
		Port               *int   `json:"port"`
		LongTimeoutSeconds *int   `json:"longTimeoutSeconds"`
	} `json:"opencode"`
	Queue struct {
		Workers *int `json:"workers"`
		Buffer  *int `json:"buffer"`
	} `json:"queue"`
	PROpen struct {
		TemplatePath  string `json:"templatePath"`
		Model         string `json:"model"`
		MaxDiffLength *int   `json:"maxDiffLength"`
	} `json:"prOpen"`
	PRComment struct {
		TemplatePath  string `json:"templatePath"`
		Model         string `json:"model"`
		MaxDiffLength *int   `json:"maxDiffLength"`
		TriggerPrefix string `json:"triggerPrefix"`
	} `json:"prComment"`
	IssueComment struct {
		TemplatePath  string `json:"templatePath"`
		ResponseModel string `json:"responseModel"`
		TriggerPrefix string `json:"triggerPrefix"`
	} `json:"issueComment"`
}
