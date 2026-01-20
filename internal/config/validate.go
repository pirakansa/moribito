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
	for repoName, repoCfg := range c.Repositories {
		if err := repoCfg.validateForWebhook(repoName); err != nil {
			return err
		}
	}
	return nil
}

func (c RepositoryConfig) validateForWebhook(repoName string) error {
	if c.PROpenConfigured {
		if !c.PROpenTemplateSet {
			return fmt.Errorf("repositories.%s.prOpen.templatePath is required", repoName)
		}
		if !c.PROpenModelSet {
			return fmt.Errorf("repositories.%s.prOpen.model is required", repoName)
		}
	}
	if c.PRCommentConfigured {
		if !c.PRCommentTemplateSet {
			return fmt.Errorf("repositories.%s.prComment.templatePath is required", repoName)
		}
		if !c.PRCommentModelSet {
			return fmt.Errorf("repositories.%s.prComment.model is required", repoName)
		}
	}
	if c.IssueCommentConfigured {
		if !c.IssueCommentTemplateSet {
			return fmt.Errorf("repositories.%s.issueComment.templatePath is required", repoName)
		}
		if !c.IssueCommentModelSet {
			return fmt.Errorf("repositories.%s.issueComment.model is required", repoName)
		}
	}
	if c.IssueLabelConfigured {
		if !c.IssueLabelTemplateSet {
			return fmt.Errorf("repositories.%s.issueLabel.templatePath is required", repoName)
		}
		if !c.IssueLabelModelSet {
			return fmt.Errorf("repositories.%s.issueLabel.model is required", repoName)
		}
	}
	if c.PRLabelConfigured {
		if !c.PRLabelTemplateSet {
			return fmt.Errorf("repositories.%s.prLabel.templatePath is required", repoName)
		}
		if !c.PRLabelModelSet {
			return fmt.Errorf("repositories.%s.prLabel.model is required", repoName)
		}
	}
	return nil
}
