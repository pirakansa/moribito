package githubapp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	defaultUserAgent = "moribito/0.1"
	maxErrorBody     = 4096
)

// Client wraps GitHub API requests.
type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
	userAgent  string
}

// NewClient creates a GitHub API client.
func NewClient(baseURL string, httpClient *http.Client) (*Client, error) {
	if strings.TrimSpace(baseURL) == "" {
		return nil, fmt.Errorf("base url is required")
	}
	parsed, err := url.Parse(strings.TrimRight(baseURL, "/"))
	if err != nil {
		return nil, fmt.Errorf("parse base url: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("base url must include scheme and host")
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		baseURL:    parsed,
		httpClient: httpClient,
		userAgent:  defaultUserAgent,
	}, nil
}

// NewRequest builds an API request for the given endpoint.
func (c *Client) NewRequest(ctx context.Context, method, endpoint, token string, body io.Reader) (*http.Request, error) {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return nil, fmt.Errorf("endpoint is required")
	}
	if !strings.HasPrefix(endpoint, "/") {
		endpoint = "/" + endpoint
	}

	target := c.baseURL.ResolveReference(&url.URL{Path: endpoint})
	req, err := http.NewRequestWithContext(ctx, method, target.String(), body)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", c.userAgent)
	if strings.TrimSpace(token) != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return req, nil
}

// DoJSON sends a JSON request and decodes a JSON response.
func (c *Client) DoJSON(ctx context.Context, method, endpoint, token string, payload any, out any) error {
	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("marshal json: %w", err)
		}
		body = bytes.NewReader(data)
	}

	req, err := c.NewRequest(ctx, method, endpoint, token, body)
	if err != nil {
		return err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyText, _ := readBodyLimit(resp.Body, maxErrorBody)
		if bodyText != "" {
			return fmt.Errorf("api error status=%d body=%s", resp.StatusCode, bodyText)
		}
		return fmt.Errorf("api error status=%d", resp.StatusCode)
	}

	if out == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode json: %w", err)
	}
	return nil
}

func readBodyLimit(r io.Reader, limit int64) (string, error) {
	data, err := io.ReadAll(io.LimitReader(r, limit))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}
