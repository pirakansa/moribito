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

type repoServices struct {
	reviewer    review.Reviewer
	issue       *issue.Service
	prCommenter review.PRCommenter
}

type repoServiceResolver struct {
	services map[string]*repoServices
}

func (r repoServiceResolver) Resolve(repoFullName string) (review.Reviewer, review.PRCommenter, *issue.Service, bool) {
	service, ok := r.services[repoFullName]
	if !ok {
		return nil, nil, nil, false
	}
	return service.reviewer, service.prCommenter, service.issue, true
}

func createRepositoryServices(cfg config.Config, logger *log.Logger, factory *githubapp.DefaultClientFactory) (map[string]*repoServices, []*opencode.Client, error) {
	services := make(map[string]*repoServices)
	var healthClients []*opencode.Client

	for repoName, repoCfg := range cfg.Repositories {
		service := &repoServices{}
		var ocClient *opencode.Client
		if repoHasAIConfig(repoCfg) {
			client, err := createOpenCodeClient(repoName, repoCfg, logger)
			if err != nil {
				return nil, nil, err
			}
			ocClient = client
			if ocClient != nil {
				healthClients = append(healthClients, ocClient)
			}
		}

		reviewOpts, err := createReviewOptions(repoName, repoCfg, logger)
		if err != nil {
			return nil, nil, err
		}
		if repoCfg.PROpenConfigured {
			if ocClient != nil {
				reviewOpts = append(reviewOpts, review.WithOpenCodeClient(ocClient))
			}
			service.reviewer = review.NewService(logger, factory, reviewOpts...)
		}

		issueService, err := createIssueService(repoName, repoCfg, logger, factory, ocClient)
		if err != nil {
			return nil, nil, err
		}
		service.issue = issueService

		prCommentService, err := createPRCommentService(repoName, repoCfg, logger, factory, ocClient)
		if err != nil {
			return nil, nil, err
		}
		service.prCommenter = prCommentService

		services[repoName] = service
	}

	return services, healthClients, nil
}

func repoHasAIConfig(cfg config.RepositoryConfig) bool {
	return cfg.PROpenConfigured || cfg.PRCommentConfigured || cfg.IssueCommentConfigured || cfg.IssueLabelConfigured || cfg.PRLabelConfigured
}

// createOpenCodeClient creates an OpenCode client for a repository.
// Returns nil if the server is unavailable.
func createOpenCodeClient(repoName string, cfg config.RepositoryConfig, logger *log.Logger) (*opencode.Client, error) {
	ocClient := opencode.NewClient(cfg.OpenCodeHost, cfg.OpenCodePort, opencode.WithLongTimeout(cfg.OpenCodeLongTimeout))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if ocClient.IsHealthy(ctx) {
		health, _ := ocClient.Health(ctx)
		logger.Printf("opencode: repo=%s connected to %s (version: %s)", repoName, ocClient.BaseURL(), health.Version)
		return ocClient, nil
	}

	logger.Printf("opencode: repo=%s server not available at %s, AI features disabled", repoName, ocClient.BaseURL())
	return nil, nil
}

func createReviewOptions(repoName string, cfg config.RepositoryConfig, logger *log.Logger) ([]review.ServiceOption, error) {
	if !cfg.PROpenConfigured {
		return nil, nil
	}

	prTemplate, err := prompt.LoadTemplateFromFile(cfg.PROpenTemplatePath)
	if err != nil {
		return nil, err
	}
	opts := []review.ServiceOption{
		review.WithPromptBuilder(prompt.NewBuilder(
			prompt.WithTemplate(prTemplate),
			prompt.WithMaxDiffLength(cfg.PROpenMaxDiffLength),
		)),
	}
	logger.Printf("prompt: repo=%s using PR open template %q", repoName, cfg.PROpenTemplatePath)
	logger.Printf("prompt: repo=%s using max diff length %d", repoName, cfg.PROpenMaxDiffLength)
	if cfg.PROpenModel != "" {
		opts = append(opts, review.WithReviewModel(cfg.PROpenModel))
		logger.Printf("review: repo=%s using model %q", repoName, cfg.PROpenModel)
	}

	return opts, nil
}

// createIssueService creates the issue response service.
// AI responses are disabled when OpenCode is not available, but reactions still work.
func createIssueService(repoName string, cfg config.RepositoryConfig, logger *log.Logger, factory *githubapp.DefaultClientFactory, ocClient *opencode.Client) (*issue.Service, error) {
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
		logger.Printf("prompt: repo=%s using issue response template %q", repoName, cfg.IssueResponseTemplatePath)
		if cfg.IssueResponseModel != "" {
			opts = append(opts, issue.WithResponseModel(cfg.IssueResponseModel))
			logger.Printf("issue: repo=%s using model %q", repoName, cfg.IssueResponseModel)
		}

		// Set custom trigger prefix if configured
		if cfg.IssueTriggerPrefix != "" {
			opts = append(opts, issue.WithTriggerPrefix(cfg.IssueTriggerPrefix))
			logger.Printf("issue: repo=%s using trigger prefix %q", repoName, cfg.IssueTriggerPrefix)
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
		logger.Printf("prompt: repo=%s using issue label template %q", repoName, cfg.IssueLabelTemplatePath)
	}
	if cfg.IssueLabelModel != "" {
		opts = append(opts, issue.WithLabelResponseModel(cfg.IssueLabelModel))
		logger.Printf("issue: repo=%s using label model %q", repoName, cfg.IssueLabelModel)
	}
	if len(cfg.IssueLabelTriggers) != 0 {
		opts = append(opts, issue.WithLabelTriggers(cfg.IssueLabelTriggers))
		logger.Printf("issue: repo=%s using label triggers %v", repoName, cfg.IssueLabelTriggers)
	}

	return issue.NewService(logger, factory, opts...), nil
}

func createPRCommentService(repoName string, cfg config.RepositoryConfig, logger *log.Logger, factory *githubapp.DefaultClientFactory, ocClient *opencode.Client) (*review.PRCommentService, error) {
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
		logger.Printf("prompt: repo=%s using PR comment template %q", repoName, cfg.PRCommentTemplatePath)
		logger.Printf("prompt: repo=%s using PR comment max diff length %d", repoName, cfg.PRCommentMaxDiffLength)
		if cfg.PRCommentModel != "" {
			opts = append(opts, review.WithCommentModel(cfg.PRCommentModel))
			logger.Printf("pr-comment: repo=%s using model %q", repoName, cfg.PRCommentModel)
		}
		if cfg.PRCommentTriggerPrefix != "" {
			opts = append(opts, review.WithCommentTriggerPrefix(cfg.PRCommentTriggerPrefix))
			logger.Printf("pr-comment: repo=%s using trigger prefix %q", repoName, cfg.PRCommentTriggerPrefix)
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
		logger.Printf("prompt: repo=%s using PR label template %q", repoName, cfg.PRLabelTemplatePath)
	}
	if cfg.PRLabelModel != "" {
		opts = append(opts, review.WithCommentLabelModel(cfg.PRLabelModel))
		logger.Printf("pr-comment: repo=%s using label model %q", repoName, cfg.PRLabelModel)
	}
	if len(cfg.PRLabelTriggers) != 0 {
		opts = append(opts, review.WithCommentLabelTriggers(cfg.PRLabelTriggers))
		logger.Printf("pr-comment: repo=%s using label triggers %v", repoName, cfg.PRLabelTriggers)
	}

	return review.NewPRCommentService(logger, factory, opts...), nil
}
