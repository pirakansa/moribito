package webhook

import (
	"bytes"
	"context"
	"log"
	"testing"
)

func TestRouterHandleKnownEvent(t *testing.T) {
	logger := log.New(&bytes.Buffer{}, "test: ", 0)
	router := NewRouter(logger)
	called := false
	router.Register("ping", func(_ context.Context, _, _ string, _ []byte) error {
		called = true
		return nil
	})

	if err := router.Handle(context.Background(), "ping", "d1", []byte(`{}`)); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if !called {
		t.Fatalf("expected handler to be called")
	}
}

func TestRouterHandleUnknownEvent(t *testing.T) {
	logger := log.New(&bytes.Buffer{}, "test: ", 0)
	router := NewRouter(logger)
	if err := router.Handle(context.Background(), "unknown", "d1", []byte(`{}`)); err != nil {
		t.Fatalf("handle unknown: %v", err)
	}
}
