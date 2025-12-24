package githubapp

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestFetchInstallationTokenSuccess(t *testing.T) {
	var gotMethod string
	var gotAuth string
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		gotMethod = r.Method
		gotAuth = r.Header.Get("Authorization")
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"token":"tok","expires_at":"2025-01-02T03:04:05Z"}`)),
			Header:     http.Header{"Content-Type": []string{"application/json"}},
		}, nil
	})
	client := &http.Client{Transport: transport}
	resp, err := FetchInstallationToken(context.Background(), client, "https://example.test", "jwt", 42)
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if gotMethod != http.MethodPost {
		t.Fatalf("expected POST, got %s", gotMethod)
	}
	if gotAuth != "Bearer jwt" {
		t.Fatalf("unexpected auth header: %s", gotAuth)
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

	transport := roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusForbidden,
			Body:       io.NopCloser(strings.NewReader("")),
			Header:     http.Header{},
		}, nil
	})
	client := &http.Client{Transport: transport}
	_, err := FetchInstallationToken(context.Background(), client, "https://example.test", "jwt", 1)
	if err == nil {
		t.Fatalf("expected error for non-2xx status")
	}
}

func TestFetchInstallationTokenMissingToken(t *testing.T) {
	transport := roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"token":"","expires_at":"2025-01-02T03:04:05Z"}`)),
			Header:     http.Header{"Content-Type": []string{"application/json"}},
		}, nil
	})
	client := &http.Client{Transport: transport}
	_, err := FetchInstallationToken(context.Background(), client, "https://example.test", "jwt", 1)
	if err == nil {
		t.Fatalf("expected error for empty token")
	}
}
