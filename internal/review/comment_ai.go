package review

import (
	"context"
	"fmt"

	"github.com/pirakansa/moribito/internal/opencode"
)

func (s *PRCommentService) requestAIResponseWithModel(ctx context.Context, promptText string, event PRCommentEvent, model string) (string, error) {
	session, err := s.opencodeClient.CreateSession(ctx, &opencode.CreateSessionRequest{
		Title: fmt.Sprintf("PR Comment: %s/%s#%d", event.Owner, event.Repo, event.Number),
	})
	if err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}
	defer func() {
		_ = s.opencodeClient.DeleteSession(context.Background(), session.ID)
	}()

	req := opencode.NewTextMessageRequest(promptText)
	if model != "" {
		req = opencode.NewTextMessageRequestWithModel(promptText, model)
	}
	resp, err := s.opencodeClient.SendMessage(ctx, session.ID, req)
	if err != nil {
		return "", fmt.Errorf("send message: %w", err)
	}

	text := opencode.ExtractTextFromResponse(resp)
	if text == "" {
		s.logger.Printf("pr-comment: AI returned empty response")
		return "", nil
	}

	return text, nil
}
