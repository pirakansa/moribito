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

	// AI review configuration
	PRReviewTemplatePath  string // PR review prompt template file path
	PRReviewModel         string // OpenCode model for PR review (optional)
	PRReviewMaxDiffLength int    // Max diff length for PR review prompt (0 disables diff)

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
	defaultPRReviewMaxDiffLen  = 50000
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
		Addr:                  defaultAddr,
		GitHubAPIBaseURL:      defaultGitHubAPIBaseURL,
		GitHubWebhookPath:     defaultWebhookPath,
		OpenCodeHost:          defaultOpenCodeHost,
		OpenCodePort:          defaultOpenCodePort,
		OpenCodeLongTimeout:   defaultOpenCodeLongTimeout,
		QueueWorkers:          defaultQueueWorkers,
		QueueBuffer:           defaultQueueBuffer,
		PRReviewModel:         defaultOpenCodeModel,
		PRReviewMaxDiffLength: defaultPRReviewMaxDiffLen,
		IssueResponseModel:    defaultOpenCodeModel,
		IssueTriggerPrefix:    defaultIssueTriggerPrefix,
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
	if fileCfg.Review.TemplatePath != "" {
		cfg.PRReviewTemplatePath = fileCfg.Review.TemplatePath
	}
	if fileCfg.Review.Model != "" {
		cfg.PRReviewModel = fileCfg.Review.Model
	}
	if fileCfg.Review.MaxDiffLength != nil {
		cfg.PRReviewMaxDiffLength = *fileCfg.Review.MaxDiffLength
	}
	if fileCfg.Issue.ResponseTemplatePath != "" {
		cfg.IssueResponseTemplatePath = fileCfg.Issue.ResponseTemplatePath
	}
	if fileCfg.Issue.ResponseModel != "" {
		cfg.IssueResponseModel = fileCfg.Issue.ResponseModel
	}
	if fileCfg.Issue.TriggerPrefix != "" {
		cfg.IssueTriggerPrefix = fileCfg.Issue.TriggerPrefix
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
	if strings.TrimSpace(c.PRReviewTemplatePath) == "" {
		return fmt.Errorf("review.templatePath is required")
	}
	if strings.TrimSpace(c.IssueResponseTemplatePath) == "" {
		return fmt.Errorf("issue.responseTemplatePath is required")
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
	Review struct {
		TemplatePath  string `json:"templatePath"`
		Model         string `json:"model"`
		MaxDiffLength *int   `json:"maxDiffLength"`
	} `json:"review"`
	Issue struct {
		ResponseTemplatePath string `json:"responseTemplatePath"`
		ResponseModel        string `json:"responseModel"`
		TriggerPrefix        string `json:"triggerPrefix"`
	} `json:"issue"`
}
