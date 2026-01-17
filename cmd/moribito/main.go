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

	healthClient := opencode.NewClient(cfg.OpenCodeHost, cfg.OpenCodePort, opencode.WithLongTimeout(cfg.OpenCodeLongTimeout))
	prCommentService, err := createPRCommentService(cfg, logger, clientFactory, ocClient)
	if err != nil {
		return err
	}

	srv := server.New(cfg, logger, jobQueue, reviewer, issueService, prCommentService, healthClient, cfg.QueueWorkers, cfg.QueueBuffer)
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
