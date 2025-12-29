// Package opencode provides types for OpenCode server API responses.
package opencode

// HealthResponse represents the response from the health endpoint.
type HealthResponse struct {
	Healthy bool   `json:"healthy"`
	Version string `json:"version"`
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

// Agent represents an available agent configuration.
type Agent struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// MCPStatus represents the status of an MCP server.
type MCPStatus struct {
	Name      string `json:"name"`
	Connected bool   `json:"connected"`
	Error     string `json:"error,omitempty"`
}
