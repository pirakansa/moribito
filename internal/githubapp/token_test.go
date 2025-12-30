package githubapp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchInstallationTokenSuccess(t *testing.T) {
	var gotMethod string
	var gotAuth string
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotAuth = r.Header.Get("Authorization")
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"token":"tok","expires_at":"2025-01-02T03:04:05Z"}`))
	}))
	t.Cleanup(server.Close)

	resp, err := FetchInstallationToken(context.Background(), server.Client(), server.URL, "jwt", 42)
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if gotMethod != http.MethodPost {
		t.Fatalf("expected POST, got %s", gotMethod)
	}
	if gotAuth != "Bearer jwt" {
		t.Fatalf("unexpected auth header: %s", gotAuth)
	}
	if gotPath != "/app/installations/42/access_tokens" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	if resp.Token != "tok" {
		t.Fatalf("unexpected token: %s", resp.Token)
	}
	if resp.ExpiresAt.IsZero() {
		t.Fatalf("expected expires_at to be set")
	}
}

func TestFetchInstallationTokenErrors(t *testing.T) {
	if _, err := FetchInstallationToken(context.Background(), nil, "http://example", "", 1); err == nil {
		t.Fatalf("expected error for missing jwt")
	}
	if _, err := FetchInstallationToken(context.Background(), nil, "http://example", "jwt", 0); err == nil {
		t.Fatalf("expected error for missing installation id")
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	t.Cleanup(server.Close)

	_, err := FetchInstallationToken(context.Background(), server.Client(), server.URL, "jwt", 1)
	if err == nil {
		t.Fatalf("expected error for non-2xx status")
	}
}

func TestFetchInstallationTokenMissingToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"token":"","expires_at":"2025-01-02T03:04:05Z"}`))
	}))
	t.Cleanup(server.Close)

	_, err := FetchInstallationToken(context.Background(), server.Client(), server.URL, "jwt", 1)
	if err == nil {
		t.Fatalf("expected error for empty token")
	}
}
