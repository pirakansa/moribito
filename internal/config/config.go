package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds configuration for the GitHub App server.
type Config struct {
	Addr              string
	AppID             int64
	InstallationID    int64
	PrivateKeyPath    string
	WebhookSecret     string
	GitHubAPIBaseURL  string
	GitHubWebhookPath string
}

const (
	defaultAddr             = ":8080"
	defaultGitHubAPIBaseURL = "https://api.github.com"
	defaultWebhookPath      = "/webhook"
)

// Load reads configuration from environment variables.
func Load() (Config, error) {
	cfg := Config{
		Addr:              envOrDefault("APP_ADDR", defaultAddr),
		PrivateKeyPath:    strings.TrimSpace(os.Getenv("GITHUB_PRIVATE_KEY_PATH")),
		WebhookSecret:     strings.TrimSpace(os.Getenv("GITHUB_WEBHOOK_SECRET")),
		GitHubAPIBaseURL:  envOrDefault("GITHUB_API_BASE_URL", defaultGitHubAPIBaseURL),
		GitHubWebhookPath: envOrDefault("GITHUB_WEBHOOK_PATH", defaultWebhookPath),
	}

	var err error
	cfg.AppID, err = parseIntEnv("GITHUB_APP_ID")
	if err != nil {
		return Config{}, err
	}
	cfg.InstallationID, err = parseIntEnv("GITHUB_INSTALLATION_ID")
	if err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// ValidateForToken ensures the minimum config for fetching installation tokens.
func (c Config) ValidateForToken() error {
	if c.AppID == 0 {
		return fmt.Errorf("GITHUB_APP_ID is required")
	}
	if c.InstallationID == 0 {
		return fmt.Errorf("GITHUB_INSTALLATION_ID is required")
	}
	if c.PrivateKeyPath == "" {
		return fmt.Errorf("GITHUB_PRIVATE_KEY_PATH is required")
	}
	return nil
}

// ValidateForWebhook ensures the minimum config for running the webhook server.
func (c Config) ValidateForWebhook() error {
	if strings.TrimSpace(c.Addr) == "" {
		return fmt.Errorf("APP_ADDR is required")
	}
	if strings.TrimSpace(c.GitHubWebhookPath) == "" {
		return fmt.Errorf("GITHUB_WEBHOOK_PATH is required")
	}
	return nil
}

func envOrDefault(key, def string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return def
	}
	return value
}

func parseIntEnv(key string) (int64, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return 0, nil
	}
	val, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}
	return val, nil
}
