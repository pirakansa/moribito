// Package config provides configuration loading and validation for the GitHub App.
// Configuration is read from environment variables.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
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
	OpenCodeHost string // OpenCode server hostname (default: 127.0.0.1)
	OpenCodePort int    // OpenCode server port (default: 4096)
}

const (
	defaultAddr             = ":8080"
	defaultGitHubAPIBaseURL = "https://api.github.com"
	defaultWebhookPath      = "/webhook"
	defaultOpenCodeHost     = "127.0.0.1"
	defaultOpenCodePort     = 4096
)

// Load reads configuration from environment variables.
// Returns partial config even if some optional values are missing;
// use ValidateForToken or ValidateForWebhook to check required fields.
//
// Environment variables:
//   - APP_ADDR: listen address (default: ":8080")
//   - GITHUB_APP_ID: GitHub App ID (optional, required for tokens)
//   - GITHUB_INSTALLATION_ID: installation ID (optional, required for tokens)
//   - GITHUB_PRIVATE_KEY_PATH: path to private key PEM file
//   - GITHUB_WEBHOOK_SECRET: webhook signature secret (optional)
//   - GITHUB_API_BASE_URL: API base URL (default: "https://api.github.com")
//   - GITHUB_WEBHOOK_PATH: webhook endpoint (default: "/webhook")
//   - OPENCODE_HOST: OpenCode server hostname (default: "127.0.0.1")
//   - OPENCODE_PORT: OpenCode server port (default: 4096)
func Load() (Config, error) {
	cfg := Config{
		Addr:              envOrDefault("APP_ADDR", defaultAddr),
		PrivateKeyPath:    strings.TrimSpace(os.Getenv("GITHUB_PRIVATE_KEY_PATH")),
		WebhookSecret:     strings.TrimSpace(os.Getenv("GITHUB_WEBHOOK_SECRET")),
		GitHubAPIBaseURL:  envOrDefault("GITHUB_API_BASE_URL", defaultGitHubAPIBaseURL),
		GitHubWebhookPath: envOrDefault("GITHUB_WEBHOOK_PATH", defaultWebhookPath),
		OpenCodeHost:      envOrDefault("OPENCODE_HOST", defaultOpenCodeHost),
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
	cfg.OpenCodePort, err = parseIntEnvDefault("OPENCODE_PORT", defaultOpenCodePort)
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

func parseIntEnvDefault(key string, def int) (int, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return def, nil
	}
	val, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}
	return val, nil
}
