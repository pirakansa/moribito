package opencode

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClient_ListMessages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/message") || r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}
		messages := []MessageWithParts{
			{
				Info:  Message{ID: "msg-1", Role: "user"},
				Parts: []Part{{Type: "text", Text: "Hello"}},
			},
			{
				Info:  Message{ID: "msg-2", Role: "assistant"},
				Parts: []Part{{Type: "text", Text: "Hi there!"}},
			},
		}
		json.NewEncoder(w).Encode(messages)
	}))
	defer server.Close()

	c := newTestClient(t, server)

	t.Run("without limit", func(t *testing.T) {
		messages, err := c.ListMessages(context.Background(), "sess-123", 0)
		if err != nil {
			t.Fatalf("ListMessages() error = %v", err)
		}
		if len(messages) != 2 {
			t.Errorf("ListMessages() got %d messages, want 2", len(messages))
		}
	})

	t.Run("with limit", func(t *testing.T) {
		messages, err := c.ListMessages(context.Background(), "sess-123", 10)
		if err != nil {
			t.Fatalf("ListMessages() error = %v", err)
		}
		if len(messages) != 2 {
			t.Errorf("ListMessages() got %d messages, want 2", len(messages))
		}
	})
}

func TestClient_SendMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/message") || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}

		var req SendMessageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		// Simulate AI response
		response := MessageWithParts{
			Info: Message{ID: "msg-response", Role: "assistant"},
			Parts: []Part{
				{Type: "text", Text: "I reviewed your code. Here are my findings..."},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	req := NewTextMessageRequest("Please review this code")
	resp, err := c.SendMessage(context.Background(), "sess-123", req)
	if err != nil {
		t.Fatalf("SendMessage() error = %v", err)
	}
	if resp.Info.Role != "assistant" {
		t.Errorf("response role = %q, want %q", resp.Info.Role, "assistant")
	}
}

func TestClient_SendMessageAsync(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/prompt_async") || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	req := NewTextMessageRequest("Please review this code")
	err := c.SendMessageAsync(context.Background(), "sess-123", req)
	if err != nil {
		t.Fatalf("SendMessageAsync() error = %v", err)
	}
}

func TestClient_GetMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/message/") || r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}
		msg := MessageWithParts{
			Info:  Message{ID: "msg-123", Role: "assistant"},
			Parts: []Part{{Type: "text", Text: "Detail"}},
		}
		_ = json.NewEncoder(w).Encode(msg)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	msg, err := c.GetMessage(context.Background(), "sess-123", "msg-123")
	if err != nil {
		t.Fatalf("GetMessage() error = %v", err)
	}
	if msg.Info.ID != "msg-123" {
		t.Errorf("message ID = %q, want %q", msg.Info.ID, "msg-123")
	}
}

func TestClient_ExecuteCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/command") || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}

		var req CommandRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		response := MessageWithParts{
			Info:  Message{ID: "msg-cmd", Role: "assistant"},
			Parts: []Part{{Type: "text", Text: "Command executed: " + req.Command}},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	req := &CommandRequest{Command: "compact", Arguments: ""}
	resp, err := c.ExecuteCommand(context.Background(), "sess-123", req)
	if err != nil {
		t.Fatalf("ExecuteCommand() error = %v", err)
	}
	if !strings.Contains(ExtractTextFromResponse(resp), "compact") {
		t.Error("ExecuteCommand() response should contain command name")
	}
}

func TestClient_RunShell(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/shell") || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}

		var req ShellRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		response := MessageWithParts{
			Info:  Message{ID: "msg-shell", Role: "assistant"},
			Parts: []Part{{Type: "text", Text: "Ran: " + req.Command}},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	resp, err := c.RunShell(context.Background(), "sess-123", &ShellRequest{Command: "echo ok"})
	if err != nil {
		t.Fatalf("RunShell() error = %v", err)
	}
	if !strings.Contains(ExtractTextFromResponse(resp), "echo ok") {
		t.Error("RunShell() response should contain command")
	}
}

func TestNewTextMessageRequest(t *testing.T) {
	req := NewTextMessageRequest("Hello, world!")
	if len(req.Parts) != 1 {
		t.Fatalf("Parts length = %d, want 1", len(req.Parts))
	}
	if req.Parts[0].Type != "text" {
		t.Errorf("Part type = %q, want %q", req.Parts[0].Type, "text")
	}
	if req.Parts[0].Text != "Hello, world!" {
		t.Errorf("Part text = %q, want %q", req.Parts[0].Text, "Hello, world!")
	}
}

func TestNewTextMessageRequestWithModel(t *testing.T) {
	req := NewTextMessageRequestWithModel("Hello", "opencode/claude-3-opus")
	if req.Model == nil {
		t.Fatal("expected model to be set")
	}
	if req.Model.ProviderID != "opencode" {
		t.Errorf("ProviderID = %q, want %q", req.Model.ProviderID, "opencode")
	}
	if req.Model.ModelID != "claude-3-opus" {
		t.Errorf("ModelID = %q, want %q", req.Model.ModelID, "claude-3-opus")
	}
	if len(req.Parts) != 1 || req.Parts[0].Text != "Hello" {
		t.Error("Parts not set correctly")
	}
}

func TestNewTextMessageRequestWithAgent(t *testing.T) {
	req := NewTextMessageRequestWithAgent("Hello", "coder")
	if req.Agent != "coder" {
		t.Errorf("Agent = %q, want %q", req.Agent, "coder")
	}
	if len(req.Parts) != 1 || req.Parts[0].Text != "Hello" {
		t.Error("Parts not set correctly")
	}
}

func TestExtractTextFromResponse(t *testing.T) {
	tests := []struct {
		name string
		msg  *MessageWithParts
		want string
	}{
		{
			name: "nil message",
			msg:  nil,
			want: "",
		},
		{
			name: "single text part",
			msg: &MessageWithParts{
				Parts: []Part{{Type: "text", Text: "Hello"}},
			},
			want: "Hello",
		},
		{
			name: "multiple text parts",
			msg: &MessageWithParts{
				Parts: []Part{
					{Type: "text", Text: "Line 1"},
					{Type: "text", Text: "Line 2"},
				},
			},
			want: "Line 1\nLine 2",
		},
		{
			name: "mixed parts",
			msg: &MessageWithParts{
				Parts: []Part{
					{Type: "text", Text: "Before"},
					{Type: "tool-invocation", ToolName: "read_file"},
					{Type: "text", Text: "After"},
				},
			},
			want: "Before\nAfter",
		},
		{
			name: "no text parts",
			msg: &MessageWithParts{
				Parts: []Part{
					{Type: "tool-invocation", ToolName: "read_file"},
				},
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractTextFromResponse(tt.msg)
			if got != tt.want {
				t.Errorf("ExtractTextFromResponse() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHasToolCalls(t *testing.T) {
	tests := []struct {
		name string
		msg  *MessageWithParts
		want bool
	}{
		{
			name: "nil message",
			msg:  nil,
			want: false,
		},
		{
			name: "no tool calls",
			msg: &MessageWithParts{
				Parts: []Part{{Type: "text", Text: "Hello"}},
			},
			want: false,
		},
		{
			name: "has tool call",
			msg: &MessageWithParts{
				Parts: []Part{
					{Type: "text", Text: "Let me check"},
					{Type: "tool-invocation", ToolName: "read_file"},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasToolCalls(tt.msg); got != tt.want {
				t.Errorf("HasToolCalls() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetToolCalls(t *testing.T) {
	msg := &MessageWithParts{
		Parts: []Part{
			{Type: "text", Text: "Let me check"},
			{Type: "tool-invocation", ToolName: "read_file"},
			{Type: "tool-result", Output: "file contents"},
			{Type: "tool-invocation", ToolName: "write_file"},
		},
	}

	calls := GetToolCalls(msg)
	if len(calls) != 2 {
		t.Fatalf("GetToolCalls() got %d calls, want 2", len(calls))
	}
	if calls[0].ToolName != "read_file" {
		t.Errorf("calls[0].ToolName = %q, want %q", calls[0].ToolName, "read_file")
	}
	if calls[1].ToolName != "write_file" {
		t.Errorf("calls[1].ToolName = %q, want %q", calls[1].ToolName, "write_file")
	}
}
