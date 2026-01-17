package opencode

import "time"

// Session represents an OpenCode session.
type Session struct {
	ID        string     `json:"id"`
	ParentID  string     `json:"parentID,omitempty"`
	Title     string     `json:"title,omitempty"`
	CreatedAt time.Time  `json:"createdAt,omitempty"`
	UpdatedAt time.Time  `json:"updatedAt,omitempty"`
	Share     *ShareInfo `json:"share,omitempty"`
}

// ShareInfo contains session sharing information.
type ShareInfo struct {
	URL string `json:"url,omitempty"`
}

// SessionStatus represents the current status of a session.
type SessionStatus struct {
	Running bool   `json:"running"`
	Error   string `json:"error,omitempty"`
}

// Todo represents a task item in a session.
type Todo struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"` // "not-started", "in-progress", "completed"
}

// FileDiff represents a file difference in a session.
type FileDiff struct {
	Path    string `json:"path"`
	Content string `json:"content,omitempty"`
	Diff    string `json:"diff,omitempty"`
}

// CreateSessionRequest represents the request to create a new session.
type CreateSessionRequest struct {
	ParentID string `json:"parentID,omitempty"`
	Title    string `json:"title,omitempty"`
}

// UpdateSessionRequest represents the request to update a session.
type UpdateSessionRequest struct {
	Title string `json:"title,omitempty"`
}

// ForkSessionRequest represents the request to fork a session.
type ForkSessionRequest struct {
	MessageID string `json:"messageID,omitempty"`
}

// SummarizeSessionRequest represents the request to summarize a session.
type SummarizeSessionRequest struct {
	ProviderID string `json:"providerID"`
	ModelID    string `json:"modelID"`
}

// PermissionResponse represents a response to a permission request.
type PermissionResponse struct {
	Response string `json:"response"` // "allow" or "deny"
	Remember bool   `json:"remember,omitempty"`
}

// RevertMessageRequest represents the request to revert a message.
type RevertMessageRequest struct {
	MessageID string `json:"messageID"`
	PartID    string `json:"partID,omitempty"`
}
