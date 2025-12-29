package opencode

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		port     int
		wantBase string
	}{
		{
			name:     "default values",
			host:     "",
			port:     0,
			wantBase: "http://127.0.0.1:4096",
		},
		{
			name:     "custom host and port",
			host:     "localhost",
			port:     8080,
			wantBase: "http://localhost:8080",
		},
		{
			name:     "custom host only",
			host:     "opencode.local",
			port:     0,
			wantBase: "http://opencode.local:4096",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient(tt.host, tt.port)
			if c.BaseURL() != tt.wantBase {
				t.Errorf("BaseURL() = %q, want %q", c.BaseURL(), tt.wantBase)
			}
		})
	}
}

func TestClient_Health(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/global/health" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: %s", r.Method)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		resp := HealthResponse{
			Healthy: true,
			Version: "1.0.0",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	resp, err := c.Health(context.Background())
	if err != nil {
		t.Fatalf("Health() error = %v", err)
	}
	if !resp.Healthy {
		t.Error("Health() healthy = false, want true")
	}
	if resp.Version != "1.0.0" {
		t.Errorf("Health() version = %q, want %q", resp.Version, "1.0.0")
	}
}

func TestClient_IsHealthy(t *testing.T) {
	tests := []struct {
		name       string
		handler    http.HandlerFunc
		wantHealth bool
	}{
		{
			name: "healthy server",
			handler: func(w http.ResponseWriter, r *http.Request) {
				resp := HealthResponse{Healthy: true, Version: "1.0.0"}
				json.NewEncoder(w).Encode(resp)
			},
			wantHealth: true,
		},
		{
			name: "unhealthy server",
			handler: func(w http.ResponseWriter, r *http.Request) {
				resp := HealthResponse{Healthy: false, Version: "1.0.0"}
				json.NewEncoder(w).Encode(resp)
			},
			wantHealth: false,
		},
		{
			name: "server error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "internal error", http.StatusInternalServerError)
			},
			wantHealth: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			c := newTestClient(t, server)
			if got := c.IsHealthy(context.Background()); got != tt.wantHealth {
				t.Errorf("IsHealthy() = %v, want %v", got, tt.wantHealth)
			}
		})
	}
}

func TestClient_GetProviders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/provider" {
			http.NotFound(w, r)
			return
		}
		resp := ProvidersResponse{
			All: []Provider{
				{ID: "anthropic", Name: "Anthropic"},
				{ID: "openai", Name: "OpenAI"},
			},
			Default:   map[string]string{"chat": "anthropic"},
			Connected: []string{"anthropic"},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	resp, err := c.GetProviders(context.Background())
	if err != nil {
		t.Fatalf("GetProviders() error = %v", err)
	}
	if len(resp.All) != 2 {
		t.Errorf("GetProviders() got %d providers, want 2", len(resp.All))
	}
	if len(resp.Connected) != 1 || resp.Connected[0] != "anthropic" {
		t.Errorf("GetProviders() connected = %v, want [anthropic]", resp.Connected)
	}
}

func TestClient_GetAgents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/agent" {
			http.NotFound(w, r)
			return
		}
		agents := []Agent{
			{ID: "coder", Name: "Coder", Description: "Code generation agent"},
			{ID: "reviewer", Name: "Reviewer", Description: "Code review agent"},
		}
		json.NewEncoder(w).Encode(agents)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	agents, err := c.GetAgents(context.Background())
	if err != nil {
		t.Fatalf("GetAgents() error = %v", err)
	}
	if len(agents) != 2 {
		t.Errorf("GetAgents() got %d agents, want 2", len(agents))
	}
}

func TestClient_GetMCPStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/mcp" {
			http.NotFound(w, r)
			return
		}
		status := map[string]MCPStatus{
			"github": {Name: "github", Connected: true},
			"slack":  {Name: "slack", Connected: false, Error: "connection failed"},
		}
		json.NewEncoder(w).Encode(status)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	status, err := c.GetMCPStatus(context.Background())
	if err != nil {
		t.Fatalf("GetMCPStatus() error = %v", err)
	}
	if len(status) != 2 {
		t.Errorf("GetMCPStatus() got %d statuses, want 2", len(status))
	}
	if !status["github"].Connected {
		t.Error("GetMCPStatus() github should be connected")
	}
	if status["slack"].Connected {
		t.Error("GetMCPStatus() slack should not be connected")
	}
}

func TestClientOption_WithTimeout(t *testing.T) {
	c := NewClient("", 0, WithTimeout(5*time.Second))
	if c.httpClient.Timeout != 5*time.Second {
		t.Errorf("timeout = %v, want %v", c.httpClient.Timeout, 5*time.Second)
	}
}

func TestClientOption_WithHTTPClient(t *testing.T) {
	customClient := &http.Client{Timeout: 10 * time.Second}
	c := NewClient("", 0, WithHTTPClient(customClient))
	if c.httpClient != customClient {
		t.Error("httpClient was not set correctly")
	}
}

// newTestClient creates a client pointing to the test server.
func newTestClient(t *testing.T, server *httptest.Server) *Client {
	t.Helper()
	// Parse the test server URL to get host and port
	// httptest.Server.URL is like "http://127.0.0.1:12345"
	c := &Client{
		baseURL:    server.URL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
	return c
}
