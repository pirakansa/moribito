package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pirakansa/moribito/internal/config"
	"github.com/pirakansa/moribito/internal/githubapp"
	"github.com/pirakansa/moribito/internal/issue"
	"github.com/pirakansa/moribito/internal/opencode"
	"github.com/pirakansa/moribito/internal/prompt"
	"github.com/pirakansa/moribito/internal/queue"
	"github.com/pirakansa/moribito/internal/review"
	"github.com/pirakansa/moribito/internal/server"
)

// main starts the GitHub App skeleton server.
func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	printToken := flag.Bool("print-installation-token", false, "print installation token and exit")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger := log.New(os.Stderr, "app: ", log.LstdFlags|log.LUTC)

	if *printToken {
		return printInstallationToken(cfg)
	}
	if err := cfg.ValidateForWebhook(); err != nil {
		return err
	}

	jobQueue := queue.New(cfg.QueueBuffer)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	jobQueue.Start(ctx, cfg.QueueWorkers)

	// Create GitHub client factory for the review service
	clientFactory := githubapp.NewClientFactory(githubapp.ClientFactoryConfig{
		AppID:          cfg.AppID,
		PrivateKeyPath: cfg.PrivateKeyPath,
		BaseURL:        cfg.GitHubAPIBaseURL,
		HTTPClient:     &http.Client{Timeout: 30 * time.Second},
	})

	// Create OpenCode client for AI-powered reviews (optional)
	reviewOpts, ocClient, err := createOpenCodeOptions(cfg, logger)
	if err != nil {
		return err
	}

	reviewer := review.NewService(logger, clientFactory, reviewOpts...)

	// Create Issue service for AI-powered issue responses
	issueService, err := createIssueService(cfg, logger, clientFactory, ocClient)
	if err != nil {
		return err
	}

	healthClient := opencode.NewClient(cfg.OpenCodeHost, cfg.OpenCodePort)
	srv := server.New(cfg, logger, jobQueue, reviewer, issueService, healthClient, cfg.QueueWorkers, cfg.QueueBuffer)
	httpServer := &http.Server{
		Addr:              cfg.Addr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Printf("listening on %s", cfg.Addr)
		errCh <- httpServer.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown: %w", err)
		}
		jobQueue.Close()
		jobQueue.Wait()
		logger.Println("shutdown complete")
		return nil
	case err := <-errCh:
		if err == nil || err == http.ErrServerClosed {
			return nil
		}
		return fmt.Errorf("listen: %w", err)
	}
}

func printInstallationToken(cfg config.Config) error {
	if err := cfg.ValidateForToken(); err != nil {
		return err
	}

	appJWT, err := githubapp.CreateAppJWT(cfg.AppID, cfg.PrivateKeyPath, time.Now())
	if err != nil {
		return err
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	token, err := githubapp.FetchInstallationToken(context.Background(), client, cfg.GitHubAPIBaseURL, appJWT, cfg.InstallationID)
	if err != nil {
		return err
	}

	fmt.Println(token.Token)
	return nil
}

// createOpenCodeOptions creates review service options for OpenCode integration.
// Returns options and the OpenCode client (nil if unavailable).
func createOpenCodeOptions(cfg config.Config, logger *log.Logger) ([]review.ServiceOption, *opencode.Client, error) {
	var opts []review.ServiceOption

	prTemplate, err := prompt.LoadTemplateFromFile(cfg.PRReviewTemplatePath)
	if err != nil {
		return nil, nil, err
	}
	opts = append(opts, review.WithPromptBuilder(prompt.NewBuilder(prompt.WithTemplate(prTemplate))))
	logger.Printf("prompt: using PR review template %q", cfg.PRReviewTemplatePath)

	// Create OpenCode client
	ocClient := opencode.NewClient(cfg.OpenCodeHost, cfg.OpenCodePort)

	// Check if OpenCode server is available
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if ocClient.IsHealthy(ctx) {
		health, _ := ocClient.Health(ctx)
		logger.Printf("opencode: connected to %s (version: %s)", ocClient.BaseURL(), health.Version)
		opts = append(opts, review.WithOpenCodeClient(ocClient))
		return opts, ocClient, nil
	}

	logger.Printf("opencode: server not available at %s, AI features disabled", ocClient.BaseURL())
	return opts, nil, nil
}

// createIssueService creates the issue response service.
// AI responses are disabled when OpenCode is not available, but reactions still work.
func createIssueService(cfg config.Config, logger *log.Logger, factory *githubapp.DefaultClientFactory, ocClient *opencode.Client) (*issue.Service, error) {
	var opts []issue.ServiceOption
	if ocClient != nil {
		opts = append(opts, issue.WithOpenCodeClient(ocClient))
	}

	issueTemplate, err := prompt.LoadTemplateFromFile(cfg.IssueResponseTemplatePath)
	if err != nil {
		return nil, err
	}
	opts = append(opts, issue.WithPromptBuilder(prompt.NewBuilder(prompt.WithTemplate(issueTemplate))))
	logger.Printf("prompt: using issue response template %q", cfg.IssueResponseTemplatePath)

	// Set custom trigger prefix if configured
	if cfg.IssueTriggerPrefix != "" {
		opts = append(opts, issue.WithTriggerPrefix(cfg.IssueTriggerPrefix))
		logger.Printf("issue: using trigger prefix %q", cfg.IssueTriggerPrefix)
	}

	return issue.NewService(logger, factory, opts...), nil
}
