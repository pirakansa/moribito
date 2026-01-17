package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
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

	cfg := defaultConfig()
	applyFileConfig(&cfg, fileCfg)
	return cfg, nil
}

func applyFileConfig(cfg *Config, fileCfg fileConfig) {
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
}
