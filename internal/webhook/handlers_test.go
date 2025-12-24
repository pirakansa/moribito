package webhook

import (
	"bytes"
	"context"
	"log"
	"testing"
)

func TestHandleInstallationInvalidJSON(t *testing.T) {
	logger := log.New(&bytes.Buffer{}, "test: ", 0)
	handler := HandleInstallation(logger)
	if err := handler(context.Background(), "installation", "d1", []byte("{")); err == nil {
		t.Fatalf("expected error for invalid json")
	}
}

func TestHandleInstallationRepositories(t *testing.T) {
	logger := log.New(&bytes.Buffer{}, "test: ", 0)
	handler := HandleInstallationRepositories(logger)
	body := []byte(`{"action":"created","installation":{"id":123}}`)
	if err := handler(context.Background(), "installation_repositories", "d1", body); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
