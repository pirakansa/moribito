package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	t.Setenv("APP_ADDR", "")
	t.Setenv("GITHUB_API_BASE_URL", "")
	t.Setenv("GITHUB_WEBHOOK_PATH", "")
	t.Setenv("OPENCODE_HOST", "")
	t.Setenv("OPENCODE_PORT", "")
	t.Setenv("OPENCODE_LONG_TIMEOUT_SECONDS", "")
	t.Setenv("QUEUE_WORKERS", "")
	t.Setenv("QUEUE_BUFFER", "")
	t.Setenv("PR_REVIEW_TEMPLATE_PATH", "/tmp/pr.tmpl")
	t.Setenv("PR_REVIEW_MODEL", "")
	t.Setenv("PR_REVIEW_MAX_DIFF_LENGTH", "")
	t.Setenv("ISSUE_RESPONSE_TEMPLATE_PATH", "/tmp/issue.tmpl")
	t.Setenv("ISSUE_RESPONSE_MODEL", "")
	t.Setenv("ISSUE_TRIGGER_PREFIX", "")

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
	if cfg.PRReviewTemplatePath != "/tmp/pr.tmpl" {
		t.Fatalf("expected pr review template path %s, got %s", "/tmp/pr.tmpl", cfg.PRReviewTemplatePath)
	}
	if cfg.PRReviewModel != defaultOpenCodeModel {
		t.Fatalf("expected pr review model %s, got %s", defaultOpenCodeModel, cfg.PRReviewModel)
	}
	if cfg.PRReviewMaxDiffLength != defaultPRReviewMaxDiffLen {
		t.Fatalf("expected pr review max diff length %d, got %d", defaultPRReviewMaxDiffLen, cfg.PRReviewMaxDiffLength)
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

func TestLoadInvalidInt(t *testing.T) {
	t.Setenv("GITHUB_APP_ID", "nope")
	if _, err := Load(); err == nil {
		t.Fatalf("expected error for invalid app id")
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
		PRReviewTemplatePath:      "/tmp/pr.tmpl",
		IssueResponseTemplatePath: "/tmp/issue.tmpl",
	}
	if err := cfg.ValidateForWebhook(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
