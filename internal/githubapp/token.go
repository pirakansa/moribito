package githubapp

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/go-github/v61/github"
)

// InstallationTokenResponse represents GitHub's token response.
type InstallationTokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// FetchInstallationToken requests an installation access token.
func FetchInstallationToken(ctx context.Context, client *http.Client, baseURL string, appJWT string, installationID int64) (InstallationTokenResponse, error) {
	if installationID == 0 {
		return InstallationTokenResponse{}, fmt.Errorf("installation id is required")
	}
	if strings.TrimSpace(appJWT) == "" {
		return InstallationTokenResponse{}, fmt.Errorf("app jwt is required")
	}
	appClient, err := newGitHubAppClient(baseURL, client, appJWT)
	if err != nil {
		return InstallationTokenResponse{}, err
	}

	token, _, err := appClient.Apps.CreateInstallationToken(ctx, installationID, &github.InstallationTokenOptions{})
	if err != nil {
		return InstallationTokenResponse{}, fmt.Errorf("create installation token: %w", err)
	}
	if token == nil || strings.TrimSpace(token.GetToken()) == "" {
		return InstallationTokenResponse{}, fmt.Errorf("token missing in response")
	}
	expiresAt := time.Time{}
	if token.ExpiresAt != nil {
		expiresAt = token.ExpiresAt.Time
	}

	return InstallationTokenResponse{
		Token:     token.GetToken(),
		ExpiresAt: expiresAt,
	}, nil
}
