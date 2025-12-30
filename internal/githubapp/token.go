package githubapp

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/go-github/v80/github"
)

// InstallationTokenResponse contains the token and expiration from GitHub.
type InstallationTokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// FetchInstallationToken requests an installation access token from GitHub.
// The token allows API calls on behalf of the installation (org/user).
// Parameters:
//   - ctx: context for cancellation
//   - client: HTTP client (with timeouts configured)
//   - baseURL: GitHub API base URL (empty for github.com)
//   - appJWT: JWT signed with the App's private key
//   - installationID: target installation ID
func FetchInstallationToken(ctx context.Context, client *http.Client, baseURL, appJWT string, installationID int64) (InstallationTokenResponse, error) {
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
