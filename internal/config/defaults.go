package config

import "time"

const (
	defaultAddr                = ":8080"
	defaultGitHubAPIBaseURL    = "https://api.github.com"
	defaultWebhookPath         = "/webhook"
	defaultOpenCodeHost        = "127.0.0.1"
	defaultOpenCodePort        = 4096
	defaultOpenCodeModel       = "opencode/big-pickle"
	defaultOpenCodeLongTimeout = 10 * time.Minute
	defaultPROpenMaxDiffLen    = 50000
	defaultPRCommentMaxDiffLen = 50000
	defaultPRLabelMaxDiffLen   = 50000
	defaultIssueTriggerPrefix  = "@moribito"
	defaultQueueWorkers        = 2
	defaultQueueBuffer         = 100
)

func defaultConfig() Config {
	return Config{
		Addr:                   defaultAddr,
		GitHubAPIBaseURL:       defaultGitHubAPIBaseURL,
		GitHubWebhookPath:      defaultWebhookPath,
		OpenCodeHost:           defaultOpenCodeHost,
		OpenCodePort:           defaultOpenCodePort,
		OpenCodeLongTimeout:    defaultOpenCodeLongTimeout,
		QueueWorkers:           defaultQueueWorkers,
		QueueBuffer:            defaultQueueBuffer,
		PROpenModel:            defaultOpenCodeModel,
		PROpenMaxDiffLength:    defaultPROpenMaxDiffLen,
		PRCommentModel:         defaultOpenCodeModel,
		PRCommentMaxDiffLength: defaultPRCommentMaxDiffLen,
		PRCommentTriggerPrefix: defaultIssueTriggerPrefix,
		IssueResponseModel:     defaultOpenCodeModel,
		IssueTriggerPrefix:     defaultIssueTriggerPrefix,
		PRLabelMaxDiffLength:   defaultPRLabelMaxDiffLen,
	}
}
