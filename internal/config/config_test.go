package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	t.Setenv("APP_ADDR", "")
	t.Setenv("GITHUB_API_BASE_URL", "")
	t.Setenv("GITHUB_WEBHOOK_PATH", "")

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
