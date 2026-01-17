package config

import (
	"fmt"
	"strings"
)

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
