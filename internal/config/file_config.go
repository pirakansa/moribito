package config

type fileConfig struct {
	Server struct {
		Addr        string `json:"addr"`
		WebhookPath string `json:"webhookPath"`
	} `json:"server"`
	GitHub struct {
		AppID          int64  `json:"appID"`
		InstallationID int64  `json:"installationID"`
		PrivateKeyPath string `json:"privateKeyPath"`
		WebhookSecret  string `json:"webhookSecret"`
		APIBaseURL     string `json:"apiBaseURL"`
	} `json:"github"`
	Queue struct {
		Workers *int `json:"workers"`
		Buffer  *int `json:"buffer"`
	} `json:"queue"`
	Repositories map[string]repositoryFileConfig `json:"repositories"`
}

type repositoryFileConfig struct {
	OpenCode struct {
		Host               string `json:"host"`
		Port               *int   `json:"port"`
		LongTimeoutSeconds *int   `json:"longTimeoutSeconds"`
	} `json:"opencode"`
	Queue struct {
		Workers *int `json:"workers"`
	} `json:"queue"`
	PROpen struct {
		TemplatePath  string `json:"templatePath"`
		Model         string `json:"model"`
		MaxDiffLength *int   `json:"maxDiffLength"`
	} `json:"prOpen"`
	PRComment struct {
		TemplatePath  string `json:"templatePath"`
		Model         string `json:"model"`
		MaxDiffLength *int   `json:"maxDiffLength"`
		TriggerPrefix string `json:"triggerPrefix"`
	} `json:"prComment"`
	IssueComment struct {
		TemplatePath  string `json:"templatePath"`
		Model         string `json:"model"`
		TriggerPrefix string `json:"triggerPrefix"`
	} `json:"issueComment"`
	IssueLabel struct {
		TemplatePath string   `json:"templatePath"`
		Model        string   `json:"model"`
		Labels       []string `json:"labels"`
	} `json:"issueLabel"`
	PRLabel struct {
		TemplatePath  string   `json:"templatePath"`
		Model         string   `json:"model"`
		MaxDiffLength *int     `json:"maxDiffLength"`
		Labels        []string `json:"labels"`
	} `json:"prLabel"`
}
