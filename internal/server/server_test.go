package server

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/pirakansa/moribito/internal/config"
	"github.com/pirakansa/moribito/internal/opencode"
)

func TestHealthWithoutOpenCode(t *testing.T) {
	cfg := config.Config{GitHubWebhookPath: "/webhook"}
	srv := New(cfg, testLogger(), nil, nil, nil, nil, 3, 200)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, rec.Code)
	}

	var resp healthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.OpenCode.Connected {
		t.Fatalf("expected opencode disconnected")
	}
	if resp.Queue.Workers != 3 || resp.Queue.Buffer != 200 {
		t.Fatalf("unexpected queue settings: %+v", resp.Queue)
	}
}

func TestHealthWithOpenCode(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/global/health", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"healthy":true,"version":"1.2.3"}`))
	})
	mux.HandleFunc("/session/status", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"a":{"running":true},"b":{"running":false}}`))
	})
	ocServer := httptest.NewServer(mux)
	defer ocServer.Close()

	ocURL, err := url.Parse(ocServer.URL)
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}
	host := ocURL.Hostname()
	port, err := strconv.Atoi(ocURL.Port())
	if err != nil {
		t.Fatalf("parse port: %v", err)
	}

	ocClient := opencode.NewClient(host, port)
	cfg := config.Config{GitHubWebhookPath: "/webhook"}
	srv := New(cfg, testLogger(), nil, nil, nil, ocClient, 2, 100)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp healthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !resp.OpenCode.Connected {
		t.Fatalf("expected opencode connected")
	}
	if resp.OpenCode.Version != "1.2.3" {
		t.Fatalf("expected version 1.2.3, got %s", resp.OpenCode.Version)
	}
	if resp.OpenCode.RunningSessions != 1 {
		t.Fatalf("expected running sessions 1, got %d", resp.OpenCode.RunningSessions)
	}
}

func testLogger() *log.Logger {
	return log.New(io.Discard, "test: ", 0)
}
