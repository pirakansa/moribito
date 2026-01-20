package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	path := writeConfigFile(t, `{
	"repositories": {
		"acme/widgets": {
			"prOpen": { "templatePath": "/tmp/pr-open.tmpl" },
			"prComment": { "templatePath": "/tmp/pr-comment.tmpl", "triggerPrefix": "@review" },
			"issueComment": { "templatePath": "/tmp/issue.tmpl" }
		}
	}
}`)
	t.Setenv("MORIBITO_CONFIG_PATH", path)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Addr != defaultAddr {
		t.Fatalf("expected default addr %s, got %s", defaultAddr, cfg.Addr)
	}
	if cfg.GitHubAPIBaseURL != defaultGitHubAPIBaseURL {
		t.Fatalf("expected default api url %s, got %s", defaultGitHubAPIBaseURL, cfg.GitHubAPIBaseURL)
	}
	if cfg.GitHubWebhookPath != defaultWebhookPath {
		t.Fatalf("expected default webhook path %s, got %s", defaultWebhookPath, cfg.GitHubWebhookPath)
	}
	if cfg.QueueWorkers != defaultQueueWorkers {
		t.Fatalf("expected default queue workers %d, got %d", defaultQueueWorkers, cfg.QueueWorkers)
	}
	if cfg.QueueBuffer != defaultQueueBuffer {
		t.Fatalf("expected default queue buffer %d, got %d", defaultQueueBuffer, cfg.QueueBuffer)
	}

	repoCfg, ok := cfg.Repositories["acme/widgets"]
	if !ok {
		t.Fatalf("expected repository config for acme/widgets")
	}
	if repoCfg.OpenCodeHost != defaultOpenCodeHost {
		t.Fatalf("expected default opencode host %s, got %s", defaultOpenCodeHost, repoCfg.OpenCodeHost)
	}
	if repoCfg.OpenCodePort != defaultOpenCodePort {
		t.Fatalf("expected default opencode port %d, got %d", defaultOpenCodePort, repoCfg.OpenCodePort)
	}
	if repoCfg.OpenCodeLongTimeout != defaultOpenCodeLongTimeout {
		t.Fatalf("expected default opencode long timeout %s, got %s", defaultOpenCodeLongTimeout, repoCfg.OpenCodeLongTimeout)
	}
	if repoCfg.PROpenTemplatePath != "/tmp/pr-open.tmpl" {
		t.Fatalf("expected pr open template path %s, got %s", "/tmp/pr-open.tmpl", repoCfg.PROpenTemplatePath)
	}
	if repoCfg.PROpenModel != defaultOpenCodeModel {
		t.Fatalf("expected pr open model %s, got %s", defaultOpenCodeModel, repoCfg.PROpenModel)
	}
	if repoCfg.PROpenMaxDiffLength != defaultPROpenMaxDiffLen {
		t.Fatalf("expected pr open max diff length %d, got %d", defaultPROpenMaxDiffLen, repoCfg.PROpenMaxDiffLength)
	}
	if !repoCfg.PROpenConfigured || !repoCfg.PROpenTemplateSet {
		t.Fatalf("expected pr open to be configured with template set")
	}
	if repoCfg.PROpenModelSet {
		t.Fatalf("expected pr open model to be unset")
	}
	if repoCfg.PRCommentTemplatePath != "/tmp/pr-comment.tmpl" {
		t.Fatalf("expected pr comment template path %s, got %s", "/tmp/pr-comment.tmpl", repoCfg.PRCommentTemplatePath)
	}
	if repoCfg.PRCommentModel != defaultOpenCodeModel {
		t.Fatalf("expected pr comment model %s, got %s", defaultOpenCodeModel, repoCfg.PRCommentModel)
	}
	if repoCfg.PRCommentMaxDiffLength != defaultPRCommentMaxDiffLen {
		t.Fatalf("expected pr comment max diff length %d, got %d", defaultPRCommentMaxDiffLen, repoCfg.PRCommentMaxDiffLength)
	}
	if repoCfg.PRCommentTriggerPrefix != "@review" {
		t.Fatalf("expected pr comment trigger prefix %s, got %s", "@review", repoCfg.PRCommentTriggerPrefix)
	}
	if !repoCfg.PRCommentConfigured || !repoCfg.PRCommentTemplateSet {
		t.Fatalf("expected pr comment to be configured with template set")
	}
	if repoCfg.PRCommentModelSet {
		t.Fatalf("expected pr comment model to be unset")
	}
	if repoCfg.IssueResponseTemplatePath != "/tmp/issue.tmpl" {
		t.Fatalf("expected issue response template path %s, got %s", "/tmp/issue.tmpl", repoCfg.IssueResponseTemplatePath)
	}
	if repoCfg.IssueResponseModel != defaultOpenCodeModel {
		t.Fatalf("expected issue response model %s, got %s", defaultOpenCodeModel, repoCfg.IssueResponseModel)
	}
	if repoCfg.IssueTriggerPrefix != defaultIssueTriggerPrefix {
		t.Fatalf("expected default issue trigger prefix %s, got %s", defaultIssueTriggerPrefix, repoCfg.IssueTriggerPrefix)
	}
	if !repoCfg.IssueCommentConfigured || !repoCfg.IssueCommentTemplateSet {
		t.Fatalf("expected issue comment to be configured with template set")
	}
	if repoCfg.IssueCommentModelSet {
		t.Fatalf("expected issue comment model to be unset")
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	path := writeConfigFile(t, `{"repositories": {"acme/widgets": {"opencode": {"port": "nope"}}}}`)
	t.Setenv("MORIBITO_CONFIG_PATH", path)
	if _, err := Load(); err == nil {
		t.Fatalf("expected error for invalid json")
	}
}

func TestLoadOverrides(t *testing.T) {
	path := writeConfigFile(t, `{
	"server": { "addr": ":9999", "webhookPath": "/hook" },
	"github": {
		"appID": 100,
		"installationID": 200,
		"privateKeyPath": "/tmp/key.pem",
		"webhookSecret": "secret",
		"apiBaseURL": "https://example.com"
	},
	"queue": { "workers": 5, "buffer": 200 },
	"repositories": {
		"acme/widgets": {
			"opencode": { "host": "opencode.local", "port": 1234, "longTimeoutSeconds": 120 },
			"prOpen": { "templatePath": "/tmp/pr-open.tmpl", "model": "custom/open", "maxDiffLength": 123 },
			"prComment": {
				"templatePath": "/tmp/pr-comment.tmpl",
				"model": "custom/comment",
				"maxDiffLength": 321,
				"triggerPrefix": "@bot"
			},
			"issueComment": {
				"templatePath": "/tmp/issue.tmpl",
				"model": "custom/issue",
				"triggerPrefix": "@issue"
			},
			"issueLabel": {
				"templatePath": "/tmp/issue-label.tmpl",
				"model": "custom/issue-label",
				"labels": ["triage", "needs-info"]
			},
			"prLabel": {
				"templatePath": "/tmp/pr-label.tmpl",
				"model": "custom/pr-label",
				"maxDiffLength": 456,
				"labels": ["needs-review"]
			}
		}
	}
}`)
	t.Setenv("MORIBITO_CONFIG_PATH", path)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Addr != ":9999" {
		t.Fatalf("expected addr :9999, got %s", cfg.Addr)
	}
	if cfg.GitHubWebhookPath != "/hook" {
		t.Fatalf("expected webhook path /hook, got %s", cfg.GitHubWebhookPath)
	}
	if cfg.AppID != 100 {
		t.Fatalf("expected appID 100, got %d", cfg.AppID)
	}
	if cfg.InstallationID != 200 {
		t.Fatalf("expected installationID 200, got %d", cfg.InstallationID)
	}
	if cfg.PrivateKeyPath != "/tmp/key.pem" {
		t.Fatalf("expected privateKeyPath /tmp/key.pem, got %s", cfg.PrivateKeyPath)
	}
	if cfg.WebhookSecret != "secret" {
		t.Fatalf("expected webhookSecret secret, got %s", cfg.WebhookSecret)
	}
	if cfg.GitHubAPIBaseURL != "https://example.com" {
		t.Fatalf("expected apiBaseURL https://example.com, got %s", cfg.GitHubAPIBaseURL)
	}
	if cfg.QueueWorkers != 5 {
		t.Fatalf("expected queue workers 5, got %d", cfg.QueueWorkers)
	}
	if cfg.QueueBuffer != 200 {
		t.Fatalf("expected queue buffer 200, got %d", cfg.QueueBuffer)
	}

	repoCfg, ok := cfg.Repositories["acme/widgets"]
	if !ok {
		t.Fatalf("expected repository config for acme/widgets")
	}
	if repoCfg.OpenCodeHost != "opencode.local" {
		t.Fatalf("expected opencode host opencode.local, got %s", repoCfg.OpenCodeHost)
	}
	if repoCfg.OpenCodePort != 1234 {
		t.Fatalf("expected opencode port 1234, got %d", repoCfg.OpenCodePort)
	}
	if repoCfg.OpenCodeLongTimeout != 120*time.Second {
		t.Fatalf("expected opencode long timeout 120s, got %s", repoCfg.OpenCodeLongTimeout)
	}
	if repoCfg.PROpenModel != "custom/open" {
		t.Fatalf("expected pr open model custom/open, got %s", repoCfg.PROpenModel)
	}
	if repoCfg.PROpenMaxDiffLength != 123 {
		t.Fatalf("expected pr open max diff length 123, got %d", repoCfg.PROpenMaxDiffLength)
	}
	if !repoCfg.PROpenConfigured || !repoCfg.PROpenTemplateSet || !repoCfg.PROpenModelSet {
		t.Fatalf("expected pr open to be configured with template/model set")
	}
	if repoCfg.PRCommentModel != "custom/comment" {
		t.Fatalf("expected pr comment model custom/comment, got %s", repoCfg.PRCommentModel)
	}
	if repoCfg.PRCommentMaxDiffLength != 321 {
		t.Fatalf("expected pr comment max diff length 321, got %d", repoCfg.PRCommentMaxDiffLength)
	}
	if repoCfg.PRCommentTriggerPrefix != "@bot" {
		t.Fatalf("expected pr comment trigger prefix @bot, got %s", repoCfg.PRCommentTriggerPrefix)
	}
	if !repoCfg.PRCommentConfigured || !repoCfg.PRCommentTemplateSet || !repoCfg.PRCommentModelSet {
		t.Fatalf("expected pr comment to be configured with template/model set")
	}
	if repoCfg.IssueResponseModel != "custom/issue" {
		t.Fatalf("expected issue response model custom/issue, got %s", repoCfg.IssueResponseModel)
	}
	if repoCfg.IssueTriggerPrefix != "@issue" {
		t.Fatalf("expected issue trigger prefix @issue, got %s", repoCfg.IssueTriggerPrefix)
	}
	if !repoCfg.IssueCommentConfigured || !repoCfg.IssueCommentTemplateSet || !repoCfg.IssueCommentModelSet {
		t.Fatalf("expected issue comment to be configured with template/model set")
	}
	if repoCfg.IssueLabelTemplatePath != "/tmp/issue-label.tmpl" {
		t.Fatalf("expected issue label template path /tmp/issue-label.tmpl, got %s", repoCfg.IssueLabelTemplatePath)
	}
	if repoCfg.IssueLabelModel != "custom/issue-label" {
		t.Fatalf("expected issue label model custom/issue-label, got %s", repoCfg.IssueLabelModel)
	}
	if !repoCfg.IssueLabelConfigured || !repoCfg.IssueLabelTemplateSet || !repoCfg.IssueLabelModelSet {
		t.Fatalf("expected issue label to be configured with template/model set")
	}
	if len(repoCfg.IssueLabelTriggers) != 2 || repoCfg.IssueLabelTriggers[0] != "triage" || repoCfg.IssueLabelTriggers[1] != "needs-info" {
		t.Fatalf("expected issue label triggers [triage needs-info], got %v", repoCfg.IssueLabelTriggers)
	}
	if repoCfg.PRLabelTemplatePath != "/tmp/pr-label.tmpl" {
		t.Fatalf("expected pr label template path /tmp/pr-label.tmpl, got %s", repoCfg.PRLabelTemplatePath)
	}
	if repoCfg.PRLabelModel != "custom/pr-label" {
		t.Fatalf("expected pr label model custom/pr-label, got %s", repoCfg.PRLabelModel)
	}
	if repoCfg.PRLabelMaxDiffLength != 456 {
		t.Fatalf("expected pr label max diff length 456, got %d", repoCfg.PRLabelMaxDiffLength)
	}
	if !repoCfg.PRLabelConfigured || !repoCfg.PRLabelTemplateSet || !repoCfg.PRLabelModelSet {
		t.Fatalf("expected pr label to be configured with template/model set")
	}
	if len(repoCfg.PRLabelTriggers) != 1 || repoCfg.PRLabelTriggers[0] != "needs-review" {
		t.Fatalf("expected pr label triggers [needs-review], got %v", repoCfg.PRLabelTriggers)
	}
}

func TestValidateForToken(t *testing.T) {
	cfg := Config{}
	if err := cfg.ValidateForToken(); err == nil {
		t.Fatalf("expected error for missing fields")
	}

	cfg = Config{
		AppID:          1,
		InstallationID: 2,
		PrivateKeyPath: "/tmp/key.pem",
	}
	if err := cfg.ValidateForToken(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateForWebhook(t *testing.T) {
	cfg := Config{}
	if err := cfg.ValidateForWebhook(); err == nil {
		t.Fatalf("expected error for missing addr")
	}

	cfg = Config{
		Addr:              ":8080",
		GitHubWebhookPath: "/webhook",
	}
	if err := cfg.ValidateForWebhook(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg = Config{
		Addr:              ":8080",
		GitHubWebhookPath: "/webhook",
		Repositories: map[string]RepositoryConfig{
			"acme/widgets": {
				PROpenConfigured:  true,
				PROpenTemplateSet: true,
			},
		},
	}
	if err := cfg.ValidateForWebhook(); err == nil {
		t.Fatalf("expected error for missing prOpen.model when configured")
	}

	cfg = Config{
		Addr:              ":8080",
		GitHubWebhookPath: "/webhook",
		Repositories: map[string]RepositoryConfig{
			"acme/widgets": {
				PROpenConfigured: true,
				PROpenModelSet:   true,
			},
		},
	}
	if err := cfg.ValidateForWebhook(); err == nil {
		t.Fatalf("expected error for missing prOpen.templatePath when configured")
	}

	cfg = Config{
		Addr:              ":8080",
		GitHubWebhookPath: "/webhook",
		Repositories: map[string]RepositoryConfig{
			"acme/widgets": {
				IssueCommentConfigured:  true,
				IssueCommentTemplateSet: true,
			},
		},
	}
	if err := cfg.ValidateForWebhook(); err == nil {
		t.Fatalf("expected error for missing issueComment.model when configured")
	}

	cfg = Config{
		Addr:              ":8080",
		GitHubWebhookPath: "/webhook",
		Repositories: map[string]RepositoryConfig{
			"acme/widgets": {
				IssueCommentConfigured: true,
				IssueCommentModelSet:   true,
			},
		},
	}
	if err := cfg.ValidateForWebhook(); err == nil {
		t.Fatalf("expected error for missing issueComment.templatePath when configured")
	}

	cfg = Config{
		Addr:              ":8080",
		GitHubWebhookPath: "/webhook",
		Repositories: map[string]RepositoryConfig{
			"acme/widgets": {
				PRCommentConfigured:  true,
				PRCommentTemplateSet: true,
			},
		},
	}
	if err := cfg.ValidateForWebhook(); err == nil {
		t.Fatalf("expected error for missing prComment.model when configured")
	}

	cfg = Config{
		Addr:              ":8080",
		GitHubWebhookPath: "/webhook",
		Repositories: map[string]RepositoryConfig{
			"acme/widgets": {
				PRCommentConfigured: true,
				PRCommentModelSet:   true,
			},
		},
	}
	if err := cfg.ValidateForWebhook(); err == nil {
		t.Fatalf("expected error for missing prComment.templatePath when configured")
	}

	cfg = Config{
		Addr:              ":8080",
		GitHubWebhookPath: "/webhook",
		Repositories: map[string]RepositoryConfig{
			"acme/widgets": {
				IssueLabelConfigured:  true,
				IssueLabelTemplateSet: true,
			},
		},
	}
	if err := cfg.ValidateForWebhook(); err == nil {
		t.Fatalf("expected error for missing issueLabel.model when configured")
	}

	cfg = Config{
		Addr:              ":8080",
		GitHubWebhookPath: "/webhook",
		Repositories: map[string]RepositoryConfig{
			"acme/widgets": {
				IssueLabelConfigured: true,
				IssueLabelModelSet:   true,
			},
		},
	}
	if err := cfg.ValidateForWebhook(); err == nil {
		t.Fatalf("expected error for missing issueLabel.templatePath when configured")
	}

	cfg = Config{
		Addr:              ":8080",
		GitHubWebhookPath: "/webhook",
		Repositories: map[string]RepositoryConfig{
			"acme/widgets": {
				PRLabelConfigured:  true,
				PRLabelTemplateSet: true,
			},
		},
	}
	if err := cfg.ValidateForWebhook(); err == nil {
		t.Fatalf("expected error for missing prLabel.model when configured")
	}

	cfg = Config{
		Addr:              ":8080",
		GitHubWebhookPath: "/webhook",
		Repositories: map[string]RepositoryConfig{
			"acme/widgets": {
				PRLabelConfigured: true,
				PRLabelModelSet:   true,
			},
		},
	}
	if err := cfg.ValidateForWebhook(); err == nil {
		t.Fatalf("expected error for missing prLabel.templatePath when configured")
	}
}

func writeConfigFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}
