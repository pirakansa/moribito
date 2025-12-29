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
	"github.com/pirakansa/moribito/internal/opencode"
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

	jobQueue := queue.New(100)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	jobQueue.Start(ctx, 2)

	// Create GitHub client factory for the review service
	clientFactory := githubapp.NewClientFactory(githubapp.ClientFactoryConfig{
		AppID:          cfg.AppID,
		PrivateKeyPath: cfg.PrivateKeyPath,
		BaseURL:        cfg.GitHubAPIBaseURL,
		HTTPClient:     &http.Client{Timeout: 30 * time.Second},
	})

	// Create OpenCode client for AI-powered reviews (optional)
	reviewOpts := createOpenCodeOptions(cfg, logger)

	reviewer := review.NewService(logger, clientFactory, reviewOpts...)
	srv := server.New(cfg, logger, jobQueue, reviewer)
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
// Returns empty options if OpenCode is not configured or unavailable.
func createOpenCodeOptions(cfg config.Config, logger *log.Logger) []review.ServiceOption {
	// Create OpenCode client
	ocClient := opencode.NewClient(cfg.OpenCodeHost, cfg.OpenCodePort)

	// Check if OpenCode server is available
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if ocClient.IsHealthy(ctx) {
		health, _ := ocClient.Health(ctx)
		logger.Printf("opencode: connected to %s (version: %s)", ocClient.BaseURL(), health.Version)
		return []review.ServiceOption{review.WithOpenCodeClient(ocClient)}
	}

	logger.Printf("opencode: server not available at %s, AI review disabled", ocClient.BaseURL())
	return nil
}
