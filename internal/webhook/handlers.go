package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
)

type installationPayload struct {
	Action       string `json:"action"`
	Installation struct {
		ID int64 `json:"id"`
	} `json:"installation"`
}

// HandleInstallation logs installation events.
func HandleInstallation(logger *log.Logger) Handler {
	return func(_ context.Context, event, delivery string, body []byte) error {
		payload, err := decodeInstallationPayload(body)
		if err != nil {
			return err
		}
		logger.Printf("event=%s delivery=%s action=%s installation_id=%d", event, delivery, payload.Action, payload.Installation.ID)
		return nil
	}
}

// HandleInstallationRepositories logs installation repositories events.
func HandleInstallationRepositories(logger *log.Logger) Handler {
	return func(_ context.Context, event, delivery string, body []byte) error {
		payload, err := decodeInstallationPayload(body)
		if err != nil {
			return err
		}
		logger.Printf("event=%s delivery=%s action=%s installation_id=%d", event, delivery, payload.Action, payload.Installation.ID)
		return nil
	}
}

func decodeInstallationPayload(body []byte) (installationPayload, error) {
	var payload installationPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return installationPayload{}, fmt.Errorf("decode installation payload: %w", err)
	}
	return payload, nil
}
