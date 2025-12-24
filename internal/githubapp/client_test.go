package githubapp

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestNewClientInvalidBaseURL(t *testing.T) {
	if _, err := NewClient("://bad", nil); err == nil {
		t.Fatalf("expected error for invalid base url")
	}
}

func TestDoJSONSuccess(t *testing.T) {
	var gotAuth string
	var gotAccept string
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		gotAuth = r.Header.Get("Authorization")
		gotAccept = r.Header.Get("Accept")
		payload, _ := json.Marshal(map[string]string{
			"message": "ok",
		})
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(string(payload))),
			Header:     http.Header{"Content-Type": []string{"application/json"}},
		}, nil
	})

	client, err := NewClient("https://example.test", &http.Client{Transport: transport})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	var out map[string]string
	if err := client.DoJSON(context.Background(), http.MethodGet, "/test", "token", nil, &out); err != nil {
		t.Fatalf("do json: %v", err)
	}
	if gotAuth != "Bearer token" {
		t.Fatalf("unexpected auth header: %s", gotAuth)
	}
	if gotAccept == "" {
		t.Fatalf("missing accept header")
	}
	if out["message"] != "ok" {
		t.Fatalf("unexpected response: %v", out)
	}
}

func TestDoJSONError(t *testing.T) {
	transport := roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusForbidden,
			Body:       io.NopCloser(strings.NewReader("forbidden")),
			Header:     http.Header{},
		}, nil
	})

	client, err := NewClient("https://example.test", &http.Client{Transport: transport})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	if err := client.DoJSON(context.Background(), http.MethodGet, "/test", "token", nil, nil); err == nil {
		t.Fatalf("expected error")
	}
}
