package opencode

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Message represents a message in a session.
type Message struct {
	ID        string    `json:"id"`
	SessionID string    `json:"sessionID,omitempty"`
	Role      string    `json:"role"` // "user", "assistant", "system"
	CreatedAt time.Time `json:"createdAt,omitempty"`
}

// Part represents a part of a message (text, tool call, tool result, etc.).
type Part struct {
	Type      string    `json:"type"` // "text", "tool-invocation", "tool-result", etc.
	Text      string    `json:"text,omitempty"`
	ToolName  string    `json:"toolName,omitempty"`
	ToolInput any       `json:"toolInput,omitempty"`
	Output    any       `json:"output,omitempty"`
	State     string    `json:"state,omitempty"` // for tool invocations: "pending", "running", "completed", "error"
	Time      *PartTime `json:"time,omitempty"`
}

// PartTime contains timing information for a part.
type PartTime struct {
	Start int64 `json:"start,omitempty"`
	End   int64 `json:"end,omitempty"`
}

// MessageWithParts combines message info with its parts.
type MessageWithParts struct {
	Info  Message `json:"info"`
	Parts []Part  `json:"parts"`
}

// SendMessageRequest represents the request to send a message.
type SendMessageRequest struct {
	MessageID string `json:"messageID,omitempty"` // Optional: continue from a specific message
	Model     string `json:"model,omitempty"`     // Optional: model to use
	Agent     string `json:"agent,omitempty"`     // Optional: agent to use
	NoReply   bool   `json:"noReply,omitempty"`   // If true, don't wait for response
	System    string `json:"system,omitempty"`    // Optional: system prompt override
	Tools     []any  `json:"tools,omitempty"`     // Optional: tools to use
	Parts     []Part `json:"parts"`               // Message parts (at minimum, text)
}

// NewTextMessageRequest creates a simple text message request.
func NewTextMessageRequest(text string) *SendMessageRequest {
	return &SendMessageRequest{
		Parts: []Part{
			{
				Type: "text",
				Text: text,
			},
		},
	}
}

// NewTextMessageRequestWithModel creates a text message request with a specific model.
func NewTextMessageRequestWithModel(text, model string) *SendMessageRequest {
	return &SendMessageRequest{
		Model: model,
		Parts: []Part{
			{
				Type: "text",
				Text: text,
			},
		},
	}
}

// NewTextMessageRequestWithAgent creates a text message request with a specific agent.
func NewTextMessageRequestWithAgent(text, agent string) *SendMessageRequest {
	return &SendMessageRequest{
		Agent: agent,
		Parts: []Part{
			{
				Type: "text",
				Text: text,
			},
		},
	}
}

// ListMessages returns all messages in a session.
// Endpoint: GET /session/:id/message
func (c *Client) ListMessages(ctx context.Context, sessionID string, limit int) ([]MessageWithParts, error) {
	path := "/session/" + sessionID + "/message"
	if limit > 0 {
		path += fmt.Sprintf("?limit=%d", limit)
	}
	var messages []MessageWithParts
	if err := c.get(ctx, path, &messages); err != nil {
		return nil, fmt.Errorf("list messages in session %s: %w", sessionID, err)
	}
	return messages, nil
}

// GetMessage returns a specific message in a session.
// Endpoint: GET /session/:id/message/:messageID
func (c *Client) GetMessage(ctx context.Context, sessionID, messageID string) (*MessageWithParts, error) {
	var msg MessageWithParts
	if err := c.get(ctx, "/session/"+sessionID+"/message/"+messageID, &msg); err != nil {
		return nil, fmt.Errorf("get message %s in session %s: %w", messageID, sessionID, err)
	}
	return &msg, nil
}

// SendMessage sends a message and waits for the response.
// This is a synchronous call that blocks until the AI responds.
// Endpoint: POST /session/:id/message
//
// Note: This may take a long time depending on the complexity of the request.
// Use SendMessageAsync for non-blocking operation.
func (c *Client) SendMessage(ctx context.Context, sessionID string, req *SendMessageRequest) (*MessageWithParts, error) {
	var msg MessageWithParts
	// Use longer timeout for message requests as AI processing can take time
	if err := c.doJSONWithTimeout(ctx, http.MethodPost, "/session/"+sessionID+"/message", req, &msg, LongTimeout); err != nil {
		return nil, fmt.Errorf("send message to session %s: %w", sessionID, err)
	}
	return &msg, nil
}

// SendMessageAsync sends a message without waiting for a response.
// Use this for fire-and-forget scenarios or when you'll poll for results.
// Endpoint: POST /session/:id/prompt_async
func (c *Client) SendMessageAsync(ctx context.Context, sessionID string, req *SendMessageRequest) error {
	if err := c.post(ctx, "/session/"+sessionID+"/prompt_async", req, nil); err != nil {
		return fmt.Errorf("send async message to session %s: %w", sessionID, err)
	}
	return nil
}

// CommandRequest represents a request to execute a slash command.
type CommandRequest struct {
	MessageID string `json:"messageID,omitempty"`
	Agent     string `json:"agent,omitempty"`
	Model     string `json:"model,omitempty"`
	Command   string `json:"command"`   // The slash command (e.g., "compact", "clear")
	Arguments string `json:"arguments"` // Command arguments
}

// ExecuteCommand executes a slash command in a session.
// Endpoint: POST /session/:id/command
func (c *Client) ExecuteCommand(ctx context.Context, sessionID string, req *CommandRequest) (*MessageWithParts, error) {
	var msg MessageWithParts
	if err := c.doJSONWithTimeout(ctx, http.MethodPost, "/session/"+sessionID+"/command", req, &msg, LongTimeout); err != nil {
		return nil, fmt.Errorf("execute command in session %s: %w", sessionID, err)
	}
	return &msg, nil
}

// ShellRequest represents a request to run a shell command.
type ShellRequest struct {
	Agent   string `json:"agent"`
	Model   string `json:"model,omitempty"`
	Command string `json:"command"`
}

// RunShell runs a shell command in a session.
// Endpoint: POST /session/:id/shell
func (c *Client) RunShell(ctx context.Context, sessionID string, req *ShellRequest) (*MessageWithParts, error) {
	var msg MessageWithParts
	if err := c.doJSONWithTimeout(ctx, http.MethodPost, "/session/"+sessionID+"/shell", req, &msg, LongTimeout); err != nil {
		return nil, fmt.Errorf("run shell in session %s: %w", sessionID, err)
	}
	return &msg, nil
}

// ExtractTextFromResponse extracts all text parts from a message response.
func ExtractTextFromResponse(msg *MessageWithParts) string {
	if msg == nil {
		return ""
	}
	var text string
	for _, part := range msg.Parts {
		if part.Type == "text" && part.Text != "" {
			if text != "" {
				text += "\n"
			}
			text += part.Text
		}
	}
	return text
}

// HasToolCalls checks if a message contains tool invocations.
func HasToolCalls(msg *MessageWithParts) bool {
	if msg == nil {
		return false
	}
	for _, part := range msg.Parts {
		if part.Type == "tool-invocation" {
			return true
		}
	}
	return false
}

// GetToolCalls returns all tool invocation parts from a message.
func GetToolCalls(msg *MessageWithParts) []Part {
	if msg == nil {
		return nil
	}
	var calls []Part
	for _, part := range msg.Parts {
		if part.Type == "tool-invocation" {
			calls = append(calls, part)
		}
	}
	return calls
}
