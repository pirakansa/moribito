package webhook

// Payload types for GitHub webhook events.
// These structs capture only the fields needed for routing and logging.
// Extend them as needed when implementing actual business logic.

type (
	// installationPayload represents GitHub App installation events.
	installationPayload struct {
		Action       string `json:"action"`
		Installation struct {
			ID int64 `json:"id"`
		} `json:"installation"`
	}

	// pullRequestPayload represents pull request events.
	pullRequestPayload struct {
		Action       string `json:"action"`
		Installation struct {
			ID int64 `json:"id"`
		} `json:"installation"`
		PullRequest struct {
			Number int `json:"number"`
		} `json:"pull_request"`
		Repository repositoryInfo `json:"repository"`
	}

	// issueCommentPayload represents issue comment events.
	issueCommentPayload struct {
		Action       string `json:"action"`
		Installation struct {
			ID int64 `json:"id"`
		} `json:"installation"`
		Issue      issueInfo      `json:"issue"`
		Comment    commentInfo    `json:"comment"`
		Repository repositoryInfo `json:"repository"`
	}

	// checkRunPayload represents check run events.
	checkRunPayload struct {
		Action     string         `json:"action"`
		CheckRun   checkRunInfo   `json:"check_run"`
		Repository repositoryInfo `json:"repository"`
	}

	// repositoryInfo is a common struct for repository metadata.
	repositoryInfo struct {
		FullName string `json:"full_name"`
	}

	// issueInfo holds issue metadata.
	issueInfo struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
		Body   string `json:"body"`
		User   struct {
			Login string `json:"login"`
		} `json:"user"`
		HTMLURL     string `json:"html_url"`
		PullRequest *struct {
			URL string `json:"url"`
		} `json:"pull_request,omitempty"`
	}

	// commentInfo holds comment metadata.
	commentInfo struct {
		ID   int64  `json:"id"`
		Body string `json:"body"`
		User struct {
			Login string `json:"login"`
		} `json:"user"`
	}

	// checkRunInfo holds check run metadata.
	checkRunInfo struct {
		ID      int64  `json:"id"`
		Name    string `json:"name"`
		HeadSHA string `json:"head_sha"`
	}
)
