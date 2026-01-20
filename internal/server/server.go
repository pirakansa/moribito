// Package server provides the HTTP server for the GitHub App.
// It handles webhook endpoints, health checks, and request validation.
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/pirakansa/moribito/internal/config"
	"github.com/pirakansa/moribito/internal/githubapp"
	"github.com/pirakansa/moribito/internal/opencode"
	"github.com/pirakansa/moribito/internal/webhook"
)

// Server provides HTTP handlers for the GitHub App.
// It manages webhook signature verification, event routing,
// and health check endpoints.
type Server struct {
	cfg             config.Config
	logger          *log.Logger
	handler         *webhook.Router
	opencodeClients []*opencode.Client
	queueWorkers    int
	queueBuffer     int
}

// New creates a Server with the given configuration and dependencies.
func New(cfg config.Config, logger *log.Logger, submitter webhook.Submitter, resolver webhook.RepoServiceResolver, ocClients []*opencode.Client, queueWorkers, queueBuffer int) *Server {
	return &Server{
		cfg:             cfg,
		logger:          logger,
		handler:         webhook.NewRouter(logger, submitter, resolver),
		opencodeClients: ocClients,
		queueWorkers:    queueWorkers,
		queueBuffer:     queueBuffer,
	}
}

// Handler returns an http.Handler with all routes configured.
// Routes:
//   - GET /healthz: health check endpoint
//   - POST <GitHubWebhookPath>: webhook receiver
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc(s.cfg.GitHubWebhookPath, s.handleWebhook)
	return mux
}

type healthResponse struct {
	Queue    healthQueue    `json:"queue"`
	OpenCode healthOpenCode `json:"opencode"`
}

type healthQueue struct {
	Workers int `json:"workers"`
	Buffer  int `json:"buffer"`
}

type healthOpenCode struct {
	Connected       bool   `json:"connected"`
	Version         string `json:"version,omitempty"`
	RunningSessions int    `json:"running_sessions"`
}

// handleHealth returns OpenCode and queue status in JSON format.
func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	resp := healthResponse{
		Queue: healthQueue{
			Workers: s.queueWorkers,
			Buffer:  s.queueBuffer,
		},
		OpenCode: healthOpenCode{
			Connected:       false,
			RunningSessions: 0,
		},
	}

	status := http.StatusOK
	if len(s.opencodeClients) == 0 {
		status = http.StatusServiceUnavailable
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		var version string
		connected := true

		for _, client := range s.opencodeClients {
			if client == nil {
				connected = false
				continue
			}
			health, err := client.Health(ctx)
			if err != nil || !health.Healthy {
				connected = false
				continue
			}
			if version == "" {
				version = health.Version
			} else if version != health.Version {
				version = "mixed"
			}
			sessions, err := client.GetSessionStatus(ctx)
			if err != nil {
				connected = false
				continue
			}
			for _, session := range sessions {
				if session.Running {
					resp.OpenCode.RunningSessions++
				}
			}
		}

		resp.OpenCode.Connected = connected
		resp.OpenCode.Version = version
		if !connected {
			status = http.StatusServiceUnavailable
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}

// handleWebhook processes incoming GitHub webhook events.
// It validates the request method, verifies the signature (if configured),
// and dispatches the event to the appropriate handler.
func (s *Server) handleWebhook(w http.ResponseWriter, r *http.Request) {
	requestID := requestIDFromHeaders(r)
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	event := r.Header.Get("X-GitHub-Event")
	delivery := r.Header.Get("X-GitHub-Delivery")
	if strings.TrimSpace(event) == "" {
		event = "unknown"
	}
	if strings.TrimSpace(delivery) == "" {
		delivery = "unknown"
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, fmt.Errorf("read body: %w", err), requestID, delivery)
		return
	}

	if s.cfg.WebhookSecret != "" {
		signature := r.Header.Get("X-Hub-Signature-256")
		if signature == "" {
			signature = r.Header.Get("X-Hub-Signature")
		}
		if !githubapp.VerifyWebhookSignature(s.cfg.WebhookSecret, body, signature) {
			s.writeError(w, http.StatusUnauthorized, fmt.Errorf("invalid webhook signature"), requestID, delivery)
			return
		}
	}

	result, err := s.handler.Handle(context.Background(), event, delivery, body)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err, requestID, delivery)
		return
	}
	if !result.Skipped {
		s.logger.Printf("webhook received event=%s delivery=%s request_id=%s bytes=%d", event, delivery, requestID, len(body))
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// writeError logs the error and responds with the appropriate HTTP status.
func (s *Server) writeError(w http.ResponseWriter, status int, err error, requestID, delivery string) {
	s.logger.Printf("http error status=%d request_id=%s delivery=%s err=%v",
		status, requestID, delivery, err)
	http.Error(w, http.StatusText(status), status)
}

// requestIDFromHeaders extracts a request ID from common header fields.
// Falls back to "unknown" if no ID is found.
func requestIDFromHeaders(r *http.Request) string {
	if value := strings.TrimSpace(r.Header.Get("X-GitHub-Request-Id")); value != "" {
		return value
	}
	if value := strings.TrimSpace(r.Header.Get("X-Request-Id")); value != "" {
		return value
	}
	return "unknown"
}
