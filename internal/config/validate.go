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
	if c.PROpenConfigured {
		if !c.PROpenTemplateSet {
			return fmt.Errorf("prOpen.templatePath is required")
		}
		if !c.PROpenModelSet {
			return fmt.Errorf("prOpen.model is required")
		}
	}
	if c.PRCommentConfigured {
		if !c.PRCommentTemplateSet {
			return fmt.Errorf("prComment.templatePath is required")
		}
		if !c.PRCommentModelSet {
			return fmt.Errorf("prComment.model is required")
		}
	}
	if c.IssueCommentConfigured {
		if !c.IssueCommentTemplateSet {
			return fmt.Errorf("issueComment.templatePath is required")
		}
		if !c.IssueCommentModelSet {
			return fmt.Errorf("issueComment.responseModel is required")
		}
	}
	if c.IssueLabelConfigured {
		if !c.IssueLabelTemplateSet {
			return fmt.Errorf("issueLabel.templatePath is required")
		}
		if !c.IssueLabelModelSet {
			return fmt.Errorf("issueLabel.model is required")
		}
	}
	if c.PRLabelConfigured {
		if !c.PRLabelTemplateSet {
			return fmt.Errorf("prLabel.templatePath is required")
		}
		if !c.PRLabelModelSet {
			return fmt.Errorf("prLabel.model is required")
		}
	}
	return nil
}
