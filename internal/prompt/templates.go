// Package prompt provides customizable prompt templates for AI interactions.
// Templates can be customized for different review styles, languages, and requirements.
package prompt

// Template defines a prompt template with a name and content.
type Template struct {
	Name        string
	Description string
	Content     string
}

// Default templates for various AI interactions.
var (
	// DefaultPRReviewTemplate is the standard template for pull request reviews.
	DefaultPRReviewTemplate = Template{
		Name:        "pr-review",
		Description: "Standard pull request code review",
		Content: `Please review the following pull request.

## Pull Request Information
- **Title**: {{.Title}}
- **Base Branch**: {{.Base}}
- **Head Branch**: {{.Head}}
- **URL**: {{.URL}}

## Description
{{.Body}}

## Changes (Diff)
` + "```diff\n{{.Diff}}\n```" + `

## Review Instructions
Please provide a code review focusing on:
1. Code quality and best practices
2. Potential bugs or issues
3. Security concerns
4. Performance implications
5. Suggestions for improvement

Keep your review concise and actionable. Use markdown formatting for clarity.`,
	}

	// ConcisePRReviewTemplate provides a shorter, focused review.
	ConcisePRReviewTemplate = Template{
		Name:        "pr-review-concise",
		Description: "Concise pull request review focusing on critical issues",
		Content: `Review this PR briefly. Focus only on:
- Critical bugs or security issues
- Breaking changes

Title: {{.Title}}
Base: {{.Base}} ← Head: {{.Head}}
` + "```diff\n{{.Diff}}\n```" + `

Reply with "LGTM" if no issues found, or list only critical concerns.`,
	}

	// JapanesePRReviewTemplate provides review output in Japanese.
	JapanesePRReviewTemplate = Template{
		Name:        "pr-review-ja",
		Description: "Pull request review in Japanese",
		Content: `以下のプルリクエストをレビューしてください。

## プルリクエスト情報
- **タイトル**: {{.Title}}
- **ベースブランチ**: {{.Base}}
- **ヘッドブランチ**: {{.Head}}

## 説明
{{.Body}}

## 変更内容 (Diff)
` + "```diff\n{{.Diff}}\n```" + `

## レビュー指示
以下の観点でコードレビューを行ってください：
1. コード品質とベストプラクティス
2. バグや問題の可能性
3. セキュリティ上の懸念
4. パフォーマンスへの影響
5. 改善提案

簡潔で実用的なレビューを日本語で提供してください。`,
	}

	// SecurityFocusedTemplate emphasizes security review.
	SecurityFocusedTemplate = Template{
		Name:        "pr-review-security",
		Description: "Security-focused pull request review",
		Content: `Perform a security-focused review of this pull request.

## PR: {{.Title}}
Branch: {{.Head}} → {{.Base}}

## Changes
` + "```diff\n{{.Diff}}\n```" + `

## Security Review Checklist
Focus exclusively on:
1. Input validation and sanitization
2. Authentication/authorization issues
3. Injection vulnerabilities (SQL, XSS, command injection)
4. Sensitive data exposure
5. Cryptographic weaknesses
6. Dependency vulnerabilities
7. Access control issues

If no security issues found, state "No security concerns identified."
Otherwise, list each issue with severity (Critical/High/Medium/Low).`,
	}

	// DefaultIssueResponseTemplate is the standard template for issue comment responses.
	DefaultIssueResponseTemplate = Template{
		Name:        "issue-response",
		Description: "Standard issue comment response",
		Content: `You are a helpful assistant for a GitHub repository. Please respond to the following issue comment.

## Issue Information
- **Title**: {{.Title}}
- **Number**: #{{.Number}}
- **Author**: @{{.Author}}
- **URL**: {{.URL}}

## Issue Description
{{.Body}}

## Comment to Respond To
**@{{.CommentAuthor}}** wrote:
{{.Comment}}

## Instructions
Please provide a helpful, friendly, and accurate response. If the question is about code:
1. Provide clear explanations
2. Include code examples when helpful
3. Reference documentation if applicable
4. Ask clarifying questions if the request is unclear

Keep your response concise and actionable.`,
	}

	// JapaneseIssueResponseTemplate provides issue responses in Japanese.
	JapaneseIssueResponseTemplate = Template{
		Name:        "issue-response-ja",
		Description: "Issue comment response in Japanese",
		Content: `あなたはGitHubリポジトリのヘルプアシスタントです。以下のIssueコメントに返答してください。

## Issue情報
- **タイトル**: {{.Title}}
- **番号**: #{{.Number}}
- **作成者**: @{{.Author}}
- **URL**: {{.URL}}

## Issue本文
{{.Body}}

## 返答対象のコメント
**@{{.CommentAuthor}}** さんのコメント:
{{.Comment}}

## 指示
親切で正確な返答を日本語で提供してください。コードに関する質問の場合：
1. わかりやすい説明を心がける
2. 必要に応じてコード例を提示
3. 関連ドキュメントがあれば参照
4. 不明点があれば確認の質問をする

簡潔で実用的な返答を心がけてください。`,
	}

	// TechnicalIssueTemplate focuses on technical problem solving.
	TechnicalIssueTemplate = Template{
		Name:        "issue-technical",
		Description: "Technical issue analysis and troubleshooting",
		Content: `Analyze this technical issue and provide troubleshooting guidance.

## Issue: {{.Title}} (#{{.Number}})

## Description
{{.Body}}

## User's Comment
@{{.CommentAuthor}}: {{.Comment}}

## Analysis Instructions
1. Identify the likely root cause
2. List potential solutions in order of likelihood
3. Provide step-by-step debugging instructions
4. Mention any relevant logs or diagnostics to check
5. Suggest preventive measures if applicable

Format your response with clear sections and code blocks where helpful.`,
	}
)

// CommentFooterTemplate is the footer added to AI review comments.
const CommentFooterTemplate = `
---
*This review was generated by [M.O.R.I.B.I.T.O.](https://github.com/pirakansa/moribito) powered by OpenCode.*`

// BuiltinTemplates returns all built-in templates.
func BuiltinTemplates() []Template {
	return []Template{
		DefaultPRReviewTemplate,
		ConcisePRReviewTemplate,
		JapanesePRReviewTemplate,
		SecurityFocusedTemplate,
		DefaultIssueResponseTemplate,
		JapaneseIssueResponseTemplate,
		TechnicalIssueTemplate,
	}
}

// GetTemplateByName returns a template by its name.
// Returns the default template if not found.
func GetTemplateByName(name string) Template {
	for _, t := range BuiltinTemplates() {
		if t.Name == name {
			return t
		}
	}
	return DefaultPRReviewTemplate
}

// GetIssueTemplateByName returns an issue template by name.
// Returns the default issue template if not found.
func GetIssueTemplateByName(name string) Template {
	for _, t := range BuiltinTemplates() {
		if t.Name == name {
			return t
		}
	}
	return DefaultIssueResponseTemplate
}
