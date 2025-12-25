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
	"github.com/pirakansa/moribito/internal/queue"
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

	srv := server.New(cfg, logger, jobQueue)
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
