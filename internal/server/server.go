package server

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"app/internal/config"
	"app/internal/githubapp"
	"app/internal/webhook"
)

// Server provides HTTP handlers for the GitHub App skeleton.
type Server struct {
	cfg     config.Config
	logger  *log.Logger
	handler *webhook.Router
}

// New builds a new Server.
func New(cfg config.Config, logger *log.Logger, submitter webhook.Submitter) *Server {
	return &Server{
		cfg:     cfg,
		logger:  logger,
		handler: webhook.NewRouter(logger, submitter),
	}
}

// Handler returns the HTTP handler for the server.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc(s.cfg.GitHubWebhookPath, s.handleWebhook)
	return mux
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

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

	if err := s.handler.Handle(context.Background(), event, delivery, body); err != nil {
		s.writeError(w, http.StatusInternalServerError, err, requestID, delivery)
		return
	}
	s.logger.Printf("webhook received event=%s delivery=%s request_id=%s bytes=%d", event, delivery, requestID, len(body))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (s *Server) writeError(w http.ResponseWriter, status int, err error, requestID string, delivery string) {
	s.logger.Printf("http error status=%d request_id=%s delivery=%s err=%v", status, requestID, delivery, err)
	http.Error(w, http.StatusText(status), status)
}

func requestIDFromHeaders(r *http.Request) string {
	if value := strings.TrimSpace(r.Header.Get("X-GitHub-Request-Id")); value != "" {
		return value
	}
	if value := strings.TrimSpace(r.Header.Get("X-Request-Id")); value != "" {
		return value
	}
	return "unknown"
}
