package opencode

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClient_CreateSession(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/session" || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}

		var req CreateSessionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		session := Session{
			ID:        "sess-12345",
			Title:     req.Title,
			ParentID:  req.ParentID,
			CreatedAt: time.Now(),
		}
		json.NewEncoder(w).Encode(session)
	}))
	defer server.Close()

	c := newTestClient(t, server)

	t.Run("create with title", func(t *testing.T) {
		session, err := c.CreateSession(context.Background(), &CreateSessionRequest{
			Title: "PR Review Session",
		})
		if err != nil {
			t.Fatalf("CreateSession() error = %v", err)
		}
		if session.ID != "sess-12345" {
			t.Errorf("session.ID = %q, want %q", session.ID, "sess-12345")
		}
		if session.Title != "PR Review Session" {
			t.Errorf("session.Title = %q, want %q", session.Title, "PR Review Session")
		}
	})

	t.Run("create with nil request", func(t *testing.T) {
		session, err := c.CreateSession(context.Background(), nil)
		if err != nil {
			t.Fatalf("CreateSession() error = %v", err)
		}
		if session.ID == "" {
			t.Error("session.ID should not be empty")
		}
	})
}

func TestClient_ListSessions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/session" || r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}
		sessions := []Session{
			{ID: "sess-1", Title: "Session 1"},
			{ID: "sess-2", Title: "Session 2"},
		}
		json.NewEncoder(w).Encode(sessions)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	sessions, err := c.ListSessions(context.Background())
	if err != nil {
		t.Fatalf("ListSessions() error = %v", err)
	}
	if len(sessions) != 2 {
		t.Errorf("ListSessions() got %d sessions, want 2", len(sessions))
	}
}

func TestClient_GetSession(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/session/") || r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}
		id := strings.TrimPrefix(r.URL.Path, "/session/")
		session := Session{ID: id, Title: "Test Session"}
		json.NewEncoder(w).Encode(session)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	session, err := c.GetSession(context.Background(), "sess-123")
	if err != nil {
		t.Fatalf("GetSession() error = %v", err)
	}
	if session.ID != "sess-123" {
		t.Errorf("session.ID = %q, want %q", session.ID, "sess-123")
	}
}

func TestClient_DeleteSession(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/session/") || r.Method != http.MethodDelete {
			http.NotFound(w, r)
			return
		}
		json.NewEncoder(w).Encode(true)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.DeleteSession(context.Background(), "sess-123")
	if err != nil {
		t.Fatalf("DeleteSession() error = %v", err)
	}
}

func TestClient_AbortSession(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/abort") || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		json.NewEncoder(w).Encode(true)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.AbortSession(context.Background(), "sess-123")
	if err != nil {
		t.Fatalf("AbortSession() error = %v", err)
	}
}

func TestClient_GetSessionTodos(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/todo") || r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}
		todos := []Todo{
			{ID: "1", Title: "Fix bug", Status: "completed"},
			{ID: "2", Title: "Add tests", Status: "in-progress"},
		}
		json.NewEncoder(w).Encode(todos)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	todos, err := c.GetSessionTodos(context.Background(), "sess-123")
	if err != nil {
		t.Fatalf("GetSessionTodos() error = %v", err)
	}
	if len(todos) != 2 {
		t.Errorf("GetSessionTodos() got %d todos, want 2", len(todos))
	}
}

func TestClient_GetSessionDiff(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/diff") || r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}
		diffs := []FileDiff{
			{Path: "main.go", Diff: "+// new line"},
		}
		json.NewEncoder(w).Encode(diffs)
	}))
	defer server.Close()

	c := newTestClient(t, server)

	t.Run("without messageID", func(t *testing.T) {
		diffs, err := c.GetSessionDiff(context.Background(), "sess-123", "")
		if err != nil {
			t.Fatalf("GetSessionDiff() error = %v", err)
		}
		if len(diffs) != 1 {
			t.Errorf("GetSessionDiff() got %d diffs, want 1", len(diffs))
		}
	})

	t.Run("with messageID", func(t *testing.T) {
		diffs, err := c.GetSessionDiff(context.Background(), "sess-123", "msg-456")
		if err != nil {
			t.Fatalf("GetSessionDiff() error = %v", err)
		}
		if len(diffs) != 1 {
			t.Errorf("GetSessionDiff() got %d diffs, want 1", len(diffs))
		}
	})
}

func TestClient_ForkSession(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/fork") || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		session := Session{ID: "sess-forked", ParentID: "sess-123"}
		json.NewEncoder(w).Encode(session)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	session, err := c.ForkSession(context.Background(), "sess-123", nil)
	if err != nil {
		t.Fatalf("ForkSession() error = %v", err)
	}
	if session.ID != "sess-forked" {
		t.Errorf("session.ID = %q, want %q", session.ID, "sess-forked")
	}
}
