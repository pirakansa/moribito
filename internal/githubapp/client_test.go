package githubapp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClientInvalidBaseURL(t *testing.T) {
	if _, err := NewClient("://bad", nil); err == nil {
		t.Fatalf("expected error for invalid base url")
	}
}

func TestDoJSONSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer token" {
			t.Fatalf("unexpected auth header: %s", got)
		}
		if got := r.Header.Get("Accept"); got == "" {
			t.Fatalf("missing accept header")
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"message": "ok",
		})
	}))
	defer server.Close()

	client, err := NewClient(server.URL, server.Client())
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	var out map[string]string
	if err := client.DoJSON(context.Background(), http.MethodGet, "/test", "token", nil, &out); err != nil {
		t.Fatalf("do json: %v", err)
	}
	if out["message"] != "ok" {
		t.Fatalf("unexpected response: %v", out)
	}
}

func TestDoJSONError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("forbidden"))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, server.Client())
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	if err := client.DoJSON(context.Background(), http.MethodGet, "/test", "token", nil, nil); err == nil {
		t.Fatalf("expected error")
	}
}
