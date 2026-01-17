package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/pirakansa/moribito/internal/config"
	"github.com/pirakansa/moribito/internal/githubapp"
)

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
