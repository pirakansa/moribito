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

type pullRequestPayload struct {
	Action      string `json:"action"`
	PullRequest struct {
		Number int `json:"number"`
	} `json:"pull_request"`
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
}

type issueCommentPayload struct {
	Action string `json:"action"`
	Issue  struct {
		Number int `json:"number"`
	} `json:"issue"`
	Comment struct {
		ID int64 `json:"id"`
	} `json:"comment"`
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
}

type checkRunPayload struct {
	Action   string `json:"action"`
	CheckRun struct {
		ID      int64  `json:"id"`
		Name    string `json:"name"`
		HeadSHA string `json:"head_sha"`
	} `json:"check_run"`
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
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

// HandlePullRequest logs pull request events.
func HandlePullRequest(logger *log.Logger) Handler {
	return func(_ context.Context, event, delivery string, body []byte) error {
		var payload pullRequestPayload
		if err := json.Unmarshal(body, &payload); err != nil {
			return fmt.Errorf("decode pull_request payload: %w", err)
		}
		logger.Printf("event=%s delivery=%s action=%s repo=%s number=%d", event, delivery, payload.Action, payload.Repository.FullName, payload.PullRequest.Number)
		return nil
	}
}

// HandleIssueComment logs issue comment events.
func HandleIssueComment(logger *log.Logger) Handler {
	return func(_ context.Context, event, delivery string, body []byte) error {
		var payload issueCommentPayload
		if err := json.Unmarshal(body, &payload); err != nil {
			return fmt.Errorf("decode issue_comment payload: %w", err)
		}
		logger.Printf("event=%s delivery=%s action=%s repo=%s issue=%d comment_id=%d", event, delivery, payload.Action, payload.Repository.FullName, payload.Issue.Number, payload.Comment.ID)
		return nil
	}
}

// HandleCheckRun logs check run events.
func HandleCheckRun(logger *log.Logger) Handler {
	return func(_ context.Context, event, delivery string, body []byte) error {
		var payload checkRunPayload
		if err := json.Unmarshal(body, &payload); err != nil {
			return fmt.Errorf("decode check_run payload: %w", err)
		}
		logger.Printf("event=%s delivery=%s action=%s repo=%s check_run=%d name=%s", event, delivery, payload.Action, payload.Repository.FullName, payload.CheckRun.ID, payload.CheckRun.Name)
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
