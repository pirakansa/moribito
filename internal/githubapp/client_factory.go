package githubapp

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// ClientFactoryConfig holds configuration for creating GitHub clients.
type ClientFactoryConfig struct {
	AppID          int64
	PrivateKeyPath string
	BaseURL        string
	HTTPClient     *http.Client
}

// DefaultClientFactory creates GitHub API clients using installation tokens.
type DefaultClientFactory struct {
	cfg        ClientFactoryConfig
	tokenCache *TokenCache
}

// NewClientFactory creates a new DefaultClientFactory.
func NewClientFactory(cfg ClientFactoryConfig) *DefaultClientFactory {
	return &DefaultClientFactory{
		cfg:        cfg,
		tokenCache: NewTokenCache(2 * time.Minute),
	}
}

// NewClient creates a GitHub API client for the given installation.
func (f *DefaultClientFactory) NewClient(ctx context.Context, installationID int64) (GitHubClient, error) {
	token, err := f.getInstallationToken(ctx, installationID)
	if err != nil {
		return nil, fmt.Errorf("get installation token: %w", err)
	}
	return NewClient(f.cfg.BaseURL, f.cfg.HTTPClient, token)
}

// getInstallationToken fetches or retrieves a cached installation token.
func (f *DefaultClientFactory) getInstallationToken(ctx context.Context, installationID int64) (string, error) {
	return f.tokenCache.Get(ctx, time.Now(), func(ctx context.Context) (InstallationTokenResponse, error) {
		appJWT, err := CreateAppJWT(f.cfg.AppID, f.cfg.PrivateKeyPath, time.Now())
		if err != nil {
			return InstallationTokenResponse{}, fmt.Errorf("create app jwt: %w", err)
		}
		return FetchInstallationToken(ctx, f.cfg.HTTPClient, f.cfg.BaseURL, appJWT, installationID)
	})
}
