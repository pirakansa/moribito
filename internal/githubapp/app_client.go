package githubapp

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/go-github/v80/github"
)

type appAuthTransport struct {
	base  http.RoundTripper
	token string
}

func (t *appAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.Header = req.Header.Clone()
	if strings.TrimSpace(t.token) != "" {
		clone.Header.Set("Authorization", "Bearer "+t.token)
	}
	return t.base.RoundTrip(clone)
}

func newGitHubAppClient(baseURL string, httpClient *http.Client, appJWT string) (*github.Client, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	baseTransport := httpClient.Transport
	if baseTransport == nil {
		baseTransport = http.DefaultTransport
	}
	clientCopy := *httpClient
	clientCopy.Transport = &appAuthTransport{
		base:  baseTransport,
		token: appJWT,
	}

	client := github.NewClient(&clientCopy)
	if strings.TrimSpace(baseURL) == "" {
		return client, nil
	}

	normalized := strings.TrimRight(baseURL, "/") + "/"
	parsed, err := url.Parse(normalized)
	if err != nil {
		return nil, fmt.Errorf("parse base url: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("base url must include scheme and host")
	}
	client.BaseURL = parsed
	return client, nil
}
