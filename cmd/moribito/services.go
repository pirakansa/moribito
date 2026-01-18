package main

import (
	"context"
	"log"
	"time"

	"github.com/pirakansa/moribito/internal/config"
	"github.com/pirakansa/moribito/internal/githubapp"
	"github.com/pirakansa/moribito/internal/issue"
	"github.com/pirakansa/moribito/internal/opencode"
	"github.com/pirakansa/moribito/internal/prompt"
	"github.com/pirakansa/moribito/internal/review"
)

// createOpenCodeOptions creates review service options for OpenCode integration.
// Returns options and the OpenCode client (nil if unavailable).
func createOpenCodeOptions(cfg config.Config, logger *log.Logger) ([]review.ServiceOption, *opencode.Client, error) {
	var opts []review.ServiceOption

	// Create OpenCode client
	ocClient := opencode.NewClient(cfg.OpenCodeHost, cfg.OpenCodePort, opencode.WithLongTimeout(cfg.OpenCodeLongTimeout))
	var activeClient *opencode.Client

	// Check if OpenCode server is available
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if ocClient.IsHealthy(ctx) {
		health, _ := ocClient.Health(ctx)
		logger.Printf("opencode: connected to %s (version: %s)", ocClient.BaseURL(), health.Version)
		activeClient = ocClient
	} else {
		logger.Printf("opencode: server not available at %s, AI features disabled", ocClient.BaseURL())
	}

	if !cfg.PROpenConfigured {
		return opts, activeClient, nil
	}

	prTemplate, err := prompt.LoadTemplateFromFile(cfg.PROpenTemplatePath)
	if err != nil {
		return nil, nil, err
	}
	opts = append(opts, review.WithPromptBuilder(prompt.NewBuilder(
		prompt.WithTemplate(prTemplate),
		prompt.WithMaxDiffLength(cfg.PROpenMaxDiffLength),
	)))
	logger.Printf("prompt: using PR open template %q", cfg.PROpenTemplatePath)
	logger.Printf("prompt: using max diff length %d", cfg.PROpenMaxDiffLength)
	if cfg.PROpenModel != "" {
		opts = append(opts, review.WithReviewModel(cfg.PROpenModel))
		logger.Printf("review: using model %q", cfg.PROpenModel)
	}
	if activeClient != nil {
		opts = append(opts, review.WithOpenCodeClient(activeClient))
	}

	return opts, activeClient, nil
}

// createIssueService creates the issue response service.
// AI responses are disabled when OpenCode is not available, but reactions still work.
func createIssueService(cfg config.Config, logger *log.Logger, factory *githubapp.DefaultClientFactory, ocClient *opencode.Client) (*issue.Service, error) {
	if !cfg.IssueCommentConfigured && !cfg.IssueLabelConfigured {
		return nil, nil
	}

	var opts []issue.ServiceOption
	if ocClient != nil {
		opts = append(opts, issue.WithOpenCodeClient(ocClient))
	}

	if cfg.IssueCommentConfigured {
		issueTemplate, err := prompt.LoadTemplateFromFile(cfg.IssueResponseTemplatePath)
		if err != nil {
			return nil, err
		}
		opts = append(opts, issue.WithPromptBuilder(prompt.NewBuilder(prompt.WithTemplate(issueTemplate))))
		logger.Printf("prompt: using issue response template %q", cfg.IssueResponseTemplatePath)
		if cfg.IssueResponseModel != "" {
			opts = append(opts, issue.WithResponseModel(cfg.IssueResponseModel))
			logger.Printf("issue: using model %q", cfg.IssueResponseModel)
		}

		// Set custom trigger prefix if configured
		if cfg.IssueTriggerPrefix != "" {
			opts = append(opts, issue.WithTriggerPrefix(cfg.IssueTriggerPrefix))
			logger.Printf("issue: using trigger prefix %q", cfg.IssueTriggerPrefix)
		}
	} else {
		opts = append(opts, issue.WithCommentEnabled(false))
	}
	if cfg.IssueLabelTemplatePath != "" {
		labelTemplate, err := prompt.LoadTemplateFromFile(cfg.IssueLabelTemplatePath)
		if err != nil {
			return nil, err
		}
		opts = append(opts, issue.WithLabelPromptBuilder(prompt.NewBuilder(
			prompt.WithTemplate(labelTemplate),
		)))
		logger.Printf("prompt: using issue label template %q", cfg.IssueLabelTemplatePath)
	}
	if cfg.IssueLabelModel != "" {
		opts = append(opts, issue.WithLabelResponseModel(cfg.IssueLabelModel))
		logger.Printf("issue: using label model %q", cfg.IssueLabelModel)
	}
	if len(cfg.IssueLabelTriggers) != 0 {
		opts = append(opts, issue.WithLabelTriggers(cfg.IssueLabelTriggers))
		logger.Printf("issue: using label triggers %v", cfg.IssueLabelTriggers)
	}

	return issue.NewService(logger, factory, opts...), nil
}

func createPRCommentService(cfg config.Config, logger *log.Logger, factory *githubapp.DefaultClientFactory, ocClient *opencode.Client) (*review.PRCommentService, error) {
	if !cfg.PRCommentConfigured && !cfg.PRLabelConfigured {
		return nil, nil
	}

	var opts []review.PRCommentOption
	if ocClient != nil {
		opts = append(opts, review.WithCommentOpenCodeClient(ocClient))
	}

	if cfg.PRCommentConfigured {
		commentTemplate, err := prompt.LoadTemplateFromFile(cfg.PRCommentTemplatePath)
		if err != nil {
			return nil, err
		}
		opts = append(opts, review.WithCommentPromptBuilder(prompt.NewBuilder(
			prompt.WithTemplate(commentTemplate),
			prompt.WithMaxDiffLength(cfg.PRCommentMaxDiffLength),
		)))
		logger.Printf("prompt: using PR comment template %q", cfg.PRCommentTemplatePath)
		logger.Printf("prompt: using PR comment max diff length %d", cfg.PRCommentMaxDiffLength)
		if cfg.PRCommentModel != "" {
			opts = append(opts, review.WithCommentModel(cfg.PRCommentModel))
			logger.Printf("pr-comment: using model %q", cfg.PRCommentModel)
		}
		if cfg.PRCommentTriggerPrefix != "" {
			opts = append(opts, review.WithCommentTriggerPrefix(cfg.PRCommentTriggerPrefix))
			logger.Printf("pr-comment: using trigger prefix %q", cfg.PRCommentTriggerPrefix)
		}
	} else {
		opts = append(opts, review.WithCommentEnabled(false))
	}
	if cfg.PRLabelTemplatePath != "" {
		labelTemplate, err := prompt.LoadTemplateFromFile(cfg.PRLabelTemplatePath)
		if err != nil {
			return nil, err
		}
		opts = append(opts, review.WithCommentLabelPromptBuilder(prompt.NewBuilder(
			prompt.WithTemplate(labelTemplate),
			prompt.WithMaxDiffLength(cfg.PRLabelMaxDiffLength),
		)))
		logger.Printf("prompt: using PR label template %q", cfg.PRLabelTemplatePath)
	}
	if cfg.PRLabelModel != "" {
		opts = append(opts, review.WithCommentLabelModel(cfg.PRLabelModel))
		logger.Printf("pr-comment: using label model %q", cfg.PRLabelModel)
	}
	if len(cfg.PRLabelTriggers) != 0 {
		opts = append(opts, review.WithCommentLabelTriggers(cfg.PRLabelTriggers))
		logger.Printf("pr-comment: using label triggers %v", cfg.PRLabelTriggers)
	}

	return review.NewPRCommentService(logger, factory, opts...), nil
}
