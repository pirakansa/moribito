package issue

import (
	"context"
	"fmt"

	"github.com/pirakansa/moribito/internal/opencode"
)

func (s *Service) requestAIResponse(ctx context.Context, promptText string) (string, error) {
	// Create session for this issue response
	session, err := s.opencodeClient.CreateSession(ctx, &opencode.CreateSessionRequest{
		Title: "Issue Response",
	})
	if err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}
	defer func() {
		// Clean up session after use
		_ = s.opencodeClient.DeleteSession(context.Background(), session.ID)
	}()

	// Send prompt and get response
	req := opencode.NewTextMessageRequest(promptText)
	if s.model != "" {
		req = opencode.NewTextMessageRequestWithModel(promptText, s.model)
	}
	msg, err := s.opencodeClient.SendMessage(ctx, session.ID, req)
	if err != nil {
		return "", fmt.Errorf("send message: %w", err)
	}

	// Extract text from response
	response := opencode.ExtractTextFromResponse(msg)
	if response == "" {
		return "", fmt.Errorf("empty response from AI")
	}

	return response, nil
}
