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

func TestClient_UpdateSession(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/session/") || r.Method != http.MethodPatch {
			http.NotFound(w, r)
			return
		}
		var req UpdateSessionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		session := Session{ID: "sess-123", Title: req.Title}
		_ = json.NewEncoder(w).Encode(session)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	session, err := c.UpdateSession(context.Background(), "sess-123", &UpdateSessionRequest{Title: "Updated"})
	if err != nil {
		t.Fatalf("UpdateSession() error = %v", err)
	}
	if session.Title != "Updated" {
		t.Errorf("session.Title = %q, want %q", session.Title, "Updated")
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

func TestClient_GetSessionStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/session/status" || r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}
		status := map[string]SessionStatus{
			"sess-1": {Running: true},
		}
		_ = json.NewEncoder(w).Encode(status)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	status, err := c.GetSessionStatus(context.Background())
	if err != nil {
		t.Fatalf("GetSessionStatus() error = %v", err)
	}
	if !status["sess-1"].Running {
		t.Fatalf("expected sess-1 running status true")
	}
}

func TestClient_GetSessionChildren(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/children") || r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}
		children := []Session{{ID: "sess-child"}}
		_ = json.NewEncoder(w).Encode(children)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	children, err := c.GetSessionChildren(context.Background(), "sess-123")
	if err != nil {
		t.Fatalf("GetSessionChildren() error = %v", err)
	}
	if len(children) != 1 {
		t.Fatalf("GetSessionChildren() got %d children, want 1", len(children))
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

func TestClient_ShareSession(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/share") || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		session := Session{ID: "sess-123", Share: &ShareInfo{URL: "https://share"}}
		_ = json.NewEncoder(w).Encode(session)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	session, err := c.ShareSession(context.Background(), "sess-123")
	if err != nil {
		t.Fatalf("ShareSession() error = %v", err)
	}
	if session.Share == nil || session.Share.URL == "" {
		t.Fatalf("expected share URL to be set")
	}
}

func TestClient_UnshareSession(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/share") || r.Method != http.MethodDelete {
			http.NotFound(w, r)
			return
		}
		session := Session{ID: "sess-123", Share: nil}
		_ = json.NewEncoder(w).Encode(session)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	session, err := c.UnshareSession(context.Background(), "sess-123")
	if err != nil {
		t.Fatalf("UnshareSession() error = %v", err)
	}
	if session.Share != nil {
		t.Fatalf("expected share to be nil")
	}
}

func TestClient_SummarizeSession(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/summarize") || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(true)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.SummarizeSession(context.Background(), "sess-123", &SummarizeSessionRequest{ProviderID: "p", ModelID: "m"})
	if err != nil {
		t.Fatalf("SummarizeSession() error = %v", err)
	}
}

func TestClient_RespondToPermission(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/permissions/") || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(true)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.RespondToPermission(context.Background(), "sess-123", "perm-1", &PermissionResponse{Response: "allow"})
	if err != nil {
		t.Fatalf("RespondToPermission() error = %v", err)
	}
}

func TestClient_RevertMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/revert") || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(true)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.RevertMessage(context.Background(), "sess-123", &RevertMessageRequest{MessageID: "msg-1"})
	if err != nil {
		t.Fatalf("RevertMessage() error = %v", err)
	}
}

func TestClient_UnrevertMessages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/unrevert") || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(true)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.UnrevertMessages(context.Background(), "sess-123")
	if err != nil {
		t.Fatalf("UnrevertMessages() error = %v", err)
	}
}
