package opencode

import (
	"context"
	"fmt"
)

// ListSessions returns all sessions.
// Endpoint: GET /session
func (c *Client) ListSessions(ctx context.Context) ([]Session, error) {
	var sessions []Session
	if err := c.get(ctx, "/session", &sessions); err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	return sessions, nil
}

// CreateSession creates a new session.
// Endpoint: POST /session
func (c *Client) CreateSession(ctx context.Context, req *CreateSessionRequest) (*Session, error) {
	if req == nil {
		req = &CreateSessionRequest{}
	}
	var session Session
	if err := c.post(ctx, "/session", req, &session); err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	return &session, nil
}

// GetSession returns a session by ID.
// Endpoint: GET /session/:id
func (c *Client) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	var session Session
	if err := c.get(ctx, "/session/"+sessionID, &session); err != nil {
		return nil, fmt.Errorf("get session %s: %w", sessionID, err)
	}
	return &session, nil
}

// UpdateSession updates a session's properties.
// Endpoint: PATCH /session/:id
func (c *Client) UpdateSession(ctx context.Context, sessionID string, req *UpdateSessionRequest) (*Session, error) {
	var session Session
	if err := c.patch(ctx, "/session/"+sessionID, req, &session); err != nil {
		return nil, fmt.Errorf("update session %s: %w", sessionID, err)
	}
	return &session, nil
}

// DeleteSession deletes a session and all its data.
// Endpoint: DELETE /session/:id
func (c *Client) DeleteSession(ctx context.Context, sessionID string) error {
	var result bool
	if err := c.delete(ctx, "/session/"+sessionID, &result); err != nil {
		return fmt.Errorf("delete session %s: %w", sessionID, err)
	}
	return nil
}

// GetSessionStatus returns the status of all sessions.
// Endpoint: GET /session/status
func (c *Client) GetSessionStatus(ctx context.Context) (map[string]SessionStatus, error) {
	var status map[string]SessionStatus
	if err := c.get(ctx, "/session/status", &status); err != nil {
		return nil, fmt.Errorf("get session status: %w", err)
	}
	return status, nil
}

// GetSessionChildren returns a session's child sessions.
// Endpoint: GET /session/:id/children
func (c *Client) GetSessionChildren(ctx context.Context, sessionID string) ([]Session, error) {
	var children []Session
	if err := c.get(ctx, "/session/"+sessionID+"/children", &children); err != nil {
		return nil, fmt.Errorf("get session children %s: %w", sessionID, err)
	}
	return children, nil
}

// GetSessionTodos returns the todo list for a session.
// Endpoint: GET /session/:id/todo
func (c *Client) GetSessionTodos(ctx context.Context, sessionID string) ([]Todo, error) {
	var todos []Todo
	if err := c.get(ctx, "/session/"+sessionID+"/todo", &todos); err != nil {
		return nil, fmt.Errorf("get session todos %s: %w", sessionID, err)
	}
	return todos, nil
}

// GetSessionDiff returns the diff for a session.
// Endpoint: GET /session/:id/diff
func (c *Client) GetSessionDiff(ctx context.Context, sessionID string, messageID string) ([]FileDiff, error) {
	path := "/session/" + sessionID + "/diff"
	if messageID != "" {
		path += "?messageID=" + messageID
	}
	var diffs []FileDiff
	if err := c.get(ctx, path, &diffs); err != nil {
		return nil, fmt.Errorf("get session diff %s: %w", sessionID, err)
	}
	return diffs, nil
}

// AbortSession aborts a running session.
// Endpoint: POST /session/:id/abort
func (c *Client) AbortSession(ctx context.Context, sessionID string) error {
	var result bool
	if err := c.post(ctx, "/session/"+sessionID+"/abort", nil, &result); err != nil {
		return fmt.Errorf("abort session %s: %w", sessionID, err)
	}
	return nil
}

// ForkSession forks an existing session at a message.
// Endpoint: POST /session/:id/fork
func (c *Client) ForkSession(ctx context.Context, sessionID string, req *ForkSessionRequest) (*Session, error) {
	if req == nil {
		req = &ForkSessionRequest{}
	}
	var session Session
	if err := c.post(ctx, "/session/"+sessionID+"/fork", req, &session); err != nil {
		return nil, fmt.Errorf("fork session %s: %w", sessionID, err)
	}
	return &session, nil
}

// ShareSession shares a session publicly.
// Endpoint: POST /session/:id/share
func (c *Client) ShareSession(ctx context.Context, sessionID string) (*Session, error) {
	var session Session
	if err := c.post(ctx, "/session/"+sessionID+"/share", nil, &session); err != nil {
		return nil, fmt.Errorf("share session %s: %w", sessionID, err)
	}
	return &session, nil
}

// UnshareSession removes public sharing from a session.
// Endpoint: DELETE /session/:id/share
func (c *Client) UnshareSession(ctx context.Context, sessionID string) (*Session, error) {
	var session Session
	if err := c.delete(ctx, "/session/"+sessionID+"/share", &session); err != nil {
		return nil, fmt.Errorf("unshare session %s: %w", sessionID, err)
	}
	return &session, nil
}

// SummarizeSession summarizes the session.
// Endpoint: POST /session/:id/summarize
func (c *Client) SummarizeSession(ctx context.Context, sessionID string, req *SummarizeSessionRequest) error {
	var result bool
	if err := c.post(ctx, "/session/"+sessionID+"/summarize", req, &result); err != nil {
		return fmt.Errorf("summarize session %s: %w", sessionID, err)
	}
	return nil
}

// RespondToPermission responds to a permission request in a session.
// Endpoint: POST /session/:id/permissions/:permissionID
func (c *Client) RespondToPermission(ctx context.Context, sessionID, permissionID string, resp *PermissionResponse) error {
	var result bool
	if err := c.post(ctx, "/session/"+sessionID+"/permissions/"+permissionID, resp, &result); err != nil {
		return fmt.Errorf("respond to permission %s/%s: %w", sessionID, permissionID, err)
	}
	return nil
}

// RevertMessage reverts a message in a session.
// Endpoint: POST /session/:id/revert
func (c *Client) RevertMessage(ctx context.Context, sessionID string, req *RevertMessageRequest) error {
	var result bool
	if err := c.post(ctx, "/session/"+sessionID+"/revert", req, &result); err != nil {
		return fmt.Errorf("revert message in session %s: %w", sessionID, err)
	}
	return nil
}

// UnrevertMessages restores all reverted messages in a session.
// Endpoint: POST /session/:id/unrevert
func (c *Client) UnrevertMessages(ctx context.Context, sessionID string) error {
	var result bool
	if err := c.post(ctx, "/session/"+sessionID+"/unrevert", nil, &result); err != nil {
		return fmt.Errorf("unrevert messages in session %s: %w", sessionID, err)
	}
	return nil
}
