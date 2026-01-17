package opencode

import (
	"context"
	"fmt"
	"net/http"
)

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
	if err := c.doJSONWithTimeout(ctx, http.MethodPost, "/session/"+sessionID+"/message", req, &msg, c.longTimeout); err != nil {
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

// ExecuteCommand executes a slash command in a session.
// Endpoint: POST /session/:id/command
func (c *Client) ExecuteCommand(ctx context.Context, sessionID string, req *CommandRequest) (*MessageWithParts, error) {
	var msg MessageWithParts
	if err := c.doJSONWithTimeout(ctx, http.MethodPost, "/session/"+sessionID+"/command", req, &msg, c.longTimeout); err != nil {
		return nil, fmt.Errorf("execute command in session %s: %w", sessionID, err)
	}
	return &msg, nil
}

// RunShell runs a shell command in a session.
// Endpoint: POST /session/:id/shell
func (c *Client) RunShell(ctx context.Context, sessionID string, req *ShellRequest) (*MessageWithParts, error) {
	var msg MessageWithParts
	if err := c.doJSONWithTimeout(ctx, http.MethodPost, "/session/"+sessionID+"/shell", req, &msg, c.longTimeout); err != nil {
		return nil, fmt.Errorf("run shell in session %s: %w", sessionID, err)
	}
	return &msg, nil
}
