package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	t.Setenv("APP_ADDR", "")
	t.Setenv("GITHUB_API_BASE_URL", "")
	t.Setenv("GITHUB_WEBHOOK_PATH", "")
	t.Setenv("OPENCODE_HOST", "")
	t.Setenv("OPENCODE_PORT", "")
	t.Setenv("PROMPT_TEMPLATE", "")

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
	if cfg.PromptTemplate != defaultPromptTemplate {
		t.Fatalf("expected default prompt template %s, got %s", defaultPromptTemplate, cfg.PromptTemplate)
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
		Addr:              ":8080",
		GitHubWebhookPath: "/webhook",
	}
	if err := cfg.ValidateForWebhook(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
