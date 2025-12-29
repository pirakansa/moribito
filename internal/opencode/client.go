// Package opencode provides a client for interacting with the OpenCode server API.
// OpenCode server exposes an OpenAPI 3.1 spec endpoint for programmatic interaction
// with AI agents that have access to MCP tools.
//
// See: https://opencode.ai/docs/server/
package opencode

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	// DefaultHost is the default hostname for the OpenCode server.
	DefaultHost = "127.0.0.1"
	// DefaultPort is the default port for the OpenCode server.
	DefaultPort = 4096
	// DefaultTimeout is the default HTTP request timeout.
	DefaultTimeout = 30 * time.Second
	// LongTimeout is used for message requests that may take longer.
	LongTimeout = 10 * time.Minute
)

// Client provides methods to interact with the OpenCode server API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// ClientOption configures the Client.
type ClientOption func(*Client)

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithTimeout sets the HTTP client timeout.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// NewClient creates a new OpenCode server client.
// Parameters:
//   - host: server hostname (use DefaultHost if empty)
//   - port: server port (use DefaultPort if 0)
//   - opts: optional configuration options
func NewClient(host string, port int, opts ...ClientOption) *Client {
	if host == "" {
		host = DefaultHost
	}
	if port == 0 {
		port = DefaultPort
	}

	c := &Client{
		baseURL: fmt.Sprintf("http://%s:%d", host, port),
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// HealthResponse represents the response from the health endpoint.
type HealthResponse struct {
	Healthy bool   `json:"healthy"`
	Version string `json:"version"`
}

// Health checks the server health and returns version information.
// Endpoint: GET /global/health
func (c *Client) Health(ctx context.Context) (*HealthResponse, error) {
	var resp HealthResponse
	if err := c.get(ctx, "/global/health", &resp); err != nil {
		return nil, fmt.Errorf("health check failed: %w", err)
	}
	return &resp, nil
}

// IsHealthy returns true if the server is reachable and healthy.
func (c *Client) IsHealthy(ctx context.Context) bool {
	resp, err := c.Health(ctx)
	return err == nil && resp.Healthy
}

// Provider represents an AI provider configuration.
type Provider struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Models []string `json:"models,omitempty"`
}

// ProvidersResponse represents the response from the providers endpoint.
type ProvidersResponse struct {
	All       []Provider        `json:"all"`
	Default   map[string]string `json:"default"`
	Connected []string          `json:"connected"`
}

// GetProviders returns the list of available AI providers.
// Endpoint: GET /provider
func (c *Client) GetProviders(ctx context.Context) (*ProvidersResponse, error) {
	var resp ProvidersResponse
	if err := c.get(ctx, "/provider", &resp); err != nil {
		return nil, fmt.Errorf("get providers failed: %w", err)
	}
	return &resp, nil
}

// Agent represents an available agent configuration.
type Agent struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// GetAgents returns the list of available agents.
// Endpoint: GET /agent
func (c *Client) GetAgents(ctx context.Context) ([]Agent, error) {
	var agents []Agent
	if err := c.get(ctx, "/agent", &agents); err != nil {
		return nil, fmt.Errorf("get agents failed: %w", err)
	}
	return agents, nil
}

// MCPStatus represents the status of an MCP server.
type MCPStatus struct {
	Name      string `json:"name"`
	Connected bool   `json:"connected"`
	Error     string `json:"error,omitempty"`
}

// GetMCPStatus returns the status of all MCP servers.
// Endpoint: GET /mcp
func (c *Client) GetMCPStatus(ctx context.Context) (map[string]MCPStatus, error) {
	var status map[string]MCPStatus
	if err := c.get(ctx, "/mcp", &status); err != nil {
		return nil, fmt.Errorf("get mcp status failed: %w", err)
	}
	return status, nil
}

// BaseURL returns the base URL of the OpenCode server.
func (c *Client) BaseURL() string {
	return c.baseURL
}

// get performs a GET request and decodes the JSON response.
func (c *Client) get(ctx context.Context, path string, result any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

// post performs a POST request with JSON body and decodes the response.
func (c *Client) post(ctx context.Context, path string, body any, result any) error {
	return c.doJSON(ctx, http.MethodPost, path, body, result)
}

// patch performs a PATCH request with JSON body and decodes the response.
func (c *Client) patch(ctx context.Context, path string, body any, result any) error {
	return c.doJSON(ctx, http.MethodPatch, path, body, result)
}

// delete performs a DELETE request and decodes the response.
func (c *Client) delete(ctx context.Context, path string, result any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	if result != nil && resp.StatusCode == http.StatusOK {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

// doJSON performs an HTTP request with JSON body and decodes the response.
func (c *Client) doJSON(ctx context.Context, method, path string, body any, result any) error {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	if result != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

// doJSONWithTimeout performs an HTTP request with a custom timeout.
func (c *Client) doJSONWithTimeout(ctx context.Context, method, path string, body any, result any, timeout time.Duration) error {
	// Create a new client with the specified timeout for this request
	client := &http.Client{
		Timeout: timeout,
	}

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	if result != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}
