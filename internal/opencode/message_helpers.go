package opencode

import "strings"

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
		Model: modelRefFromString(model),
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

func modelRefFromString(model string) *ModelRef {
	trimmed := strings.TrimSpace(model)
	if trimmed == "" {
		return nil
	}
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) == 2 {
		return &ModelRef{ProviderID: parts[0], ModelID: parts[1]}
	}
	return &ModelRef{ModelID: trimmed}
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
