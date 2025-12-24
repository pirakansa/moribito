package githubapp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"
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
	if client == nil {
		client = http.DefaultClient
	}

	url := strings.TrimRight(baseURL, "/")
	url = url + path.Join("/app/installations", fmt.Sprintf("%d", installationID), "access_tokens")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return InstallationTokenResponse{}, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+appJWT)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return InstallationTokenResponse{}, fmt.Errorf("request token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return InstallationTokenResponse{}, fmt.Errorf("token request failed: status=%d", resp.StatusCode)
	}

	var payload InstallationTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return InstallationTokenResponse{}, fmt.Errorf("decode token: %w", err)
	}
	if strings.TrimSpace(payload.Token) == "" {
		return InstallationTokenResponse{}, fmt.Errorf("token missing in response")
	}

	return payload, nil
}
