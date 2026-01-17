package opencode

import "time"

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

// ModelRef identifies a model in the OpenCode server.
type ModelRef struct {
	ProviderID string `json:"providerID,omitempty"`
	ModelID    string `json:"modelID,omitempty"`
}

// SendMessageRequest represents the request to send a message.
type SendMessageRequest struct {
	MessageID string    `json:"messageID,omitempty"` // Optional: continue from a specific message
	Model     *ModelRef `json:"model,omitempty"`     // Optional: model to use
	Agent     string    `json:"agent,omitempty"`     // Optional: agent to use
	NoReply   bool      `json:"noReply,omitempty"`   // If true, don't wait for response
	System    string    `json:"system,omitempty"`    // Optional: system prompt override
	Tools     []any     `json:"tools,omitempty"`     // Optional: tools to use
	Parts     []Part    `json:"parts"`               // Message parts (at minimum, text)
}

// CommandRequest represents a request to execute a slash command.
type CommandRequest struct {
	MessageID string    `json:"messageID,omitempty"`
	Agent     string    `json:"agent,omitempty"`
	Model     *ModelRef `json:"model,omitempty"`
	Command   string    `json:"command"`   // The slash command (e.g., "compact", "clear")
	Arguments string    `json:"arguments"` // Command arguments
}

// ShellRequest represents a request to run a shell command.
type ShellRequest struct {
	Agent   string    `json:"agent"`
	Model   *ModelRef `json:"model,omitempty"`
	Command string    `json:"command"`
}
