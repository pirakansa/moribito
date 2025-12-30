// Package opencode provides a client for interacting with the OpenCode server API.
// OpenCode server exposes an OpenAPI 3.1 spec endpoint for programmatic interaction
// with AI agents that have access to MCP tools.
//
// See: https://opencode.ai/docs/server/
package opencode

import (
	"context"
	"fmt"
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

// GetProviders returns the list of available AI providers.
// Endpoint: GET /provider
func (c *Client) GetProviders(ctx context.Context) (*ProvidersResponse, error) {
	var resp ProvidersResponse
	if err := c.get(ctx, "/provider", &resp); err != nil {
		return nil, fmt.Errorf("get providers failed: %w", err)
	}
	return &resp, nil
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
