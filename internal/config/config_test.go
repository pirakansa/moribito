package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	path := writeConfigFile(t, `{
	"prOpen": { "templatePath": "/tmp/pr-open.tmpl" },
	"prComment": { "templatePath": "/tmp/pr-comment.tmpl" },
	"issue": { "responseTemplatePath": "/tmp/issue.tmpl" }
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
	if cfg.OpenCodeHost != defaultOpenCodeHost {
		t.Fatalf("expected default opencode host %s, got %s", defaultOpenCodeHost, cfg.OpenCodeHost)
	}
	if cfg.OpenCodePort != defaultOpenCodePort {
		t.Fatalf("expected default opencode port %d, got %d", defaultOpenCodePort, cfg.OpenCodePort)
	}
	if cfg.OpenCodeLongTimeout != defaultOpenCodeLongTimeout {
		t.Fatalf("expected default opencode long timeout %s, got %s", defaultOpenCodeLongTimeout, cfg.OpenCodeLongTimeout)
	}
	if cfg.QueueWorkers != defaultQueueWorkers {
		t.Fatalf("expected default queue workers %d, got %d", defaultQueueWorkers, cfg.QueueWorkers)
	}
	if cfg.QueueBuffer != defaultQueueBuffer {
		t.Fatalf("expected default queue buffer %d, got %d", defaultQueueBuffer, cfg.QueueBuffer)
	}
	if cfg.PROpenTemplatePath != "/tmp/pr-open.tmpl" {
		t.Fatalf("expected pr open template path %s, got %s", "/tmp/pr-open.tmpl", cfg.PROpenTemplatePath)
	}
	if cfg.PROpenModel != defaultOpenCodeModel {
		t.Fatalf("expected pr open model %s, got %s", defaultOpenCodeModel, cfg.PROpenModel)
	}
	if cfg.PROpenMaxDiffLength != defaultPROpenMaxDiffLen {
		t.Fatalf("expected pr open max diff length %d, got %d", defaultPROpenMaxDiffLen, cfg.PROpenMaxDiffLength)
	}
	if cfg.PRCommentTemplatePath != "/tmp/pr-comment.tmpl" {
		t.Fatalf("expected pr comment template path %s, got %s", "/tmp/pr-comment.tmpl", cfg.PRCommentTemplatePath)
	}
	if cfg.PRCommentModel != defaultOpenCodeModel {
		t.Fatalf("expected pr comment model %s, got %s", defaultOpenCodeModel, cfg.PRCommentModel)
	}
	if cfg.PRCommentMaxDiffLength != defaultPRCommentMaxDiffLen {
		t.Fatalf("expected pr comment max diff length %d, got %d", defaultPRCommentMaxDiffLen, cfg.PRCommentMaxDiffLength)
	}
	if cfg.IssueResponseTemplatePath != "/tmp/issue.tmpl" {
		t.Fatalf("expected issue response template path %s, got %s", "/tmp/issue.tmpl", cfg.IssueResponseTemplatePath)
	}
	if cfg.IssueResponseModel != defaultOpenCodeModel {
		t.Fatalf("expected issue response model %s, got %s", defaultOpenCodeModel, cfg.IssueResponseModel)
	}
	if cfg.IssueTriggerPrefix != defaultIssueTriggerPrefix {
		t.Fatalf("expected default issue trigger prefix %s, got %s", defaultIssueTriggerPrefix, cfg.IssueTriggerPrefix)
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	path := writeConfigFile(t, `{"opencode": {"port": "nope"}}`)
	t.Setenv("MORIBITO_CONFIG_PATH", path)
	if _, err := Load(); err == nil {
		t.Fatalf("expected error for invalid json")
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
		Addr:                      ":8080",
		GitHubWebhookPath:         "/webhook",
		PROpenTemplatePath:        "/tmp/pr.tmpl",
		IssueResponseTemplatePath: "/tmp/issue.tmpl",
	}
	if err := cfg.ValidateForWebhook(); err != nil {
		t.Fatalf("unexpected error: %v", err)
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
