package prompt

import (
	"strings"
	"testing"
)

func TestBuiltinTemplates(t *testing.T) {
	templates := BuiltinTemplates()
	if len(templates) == 0 {
		t.Fatal("expected at least one builtin template")
	}

	// Check all templates have required fields
	for _, tmpl := range templates {
		if tmpl.Name == "" {
			t.Error("template name should not be empty")
		}
		if tmpl.Content == "" {
			t.Errorf("template %s content should not be empty", tmpl.Name)
		}
	}
}

func TestGetTemplateByName(t *testing.T) {
	tests := []struct {
		name     string
		wantName string
	}{
		{"pr-review", "pr-review"},
		{"pr-review-concise", "pr-review-concise"},
		{"pr-review-ja", "pr-review-ja"},
		{"pr-review-security", "pr-review-security"},
		{"nonexistent", "pr-review"}, // Should return default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl := GetTemplateByName(tt.name)
			if tmpl.Name != tt.wantName {
				t.Errorf("GetTemplateByName(%q) = %q, want %q", tt.name, tmpl.Name, tt.wantName)
			}
		})
	}
}

func TestDefaultTemplateContainsPlaceholders(t *testing.T) {
	tmpl := DefaultPRReviewTemplate

	placeholders := []string{"{{.Title}}", "{{.Body}}", "{{.Head}}", "{{.Base}}", "{{.Diff}}"}
	for _, ph := range placeholders {
		if !strings.Contains(tmpl.Content, ph) {
			t.Errorf("default template should contain placeholder %s", ph)
		}
	}
}

func TestJapaneseTemplateIsInJapanese(t *testing.T) {
	tmpl := JapanesePRReviewTemplate

	// Check for Japanese characters
	if !strings.Contains(tmpl.Content, "レビュー") {
		t.Error("Japanese template should contain Japanese text")
	}
}

func TestSecurityTemplateHasSecurityFocus(t *testing.T) {
	tmpl := SecurityFocusedTemplate

	securityTerms := []string{"security", "vulnerabilities", "injection", "authentication"}
	found := 0
	for _, term := range securityTerms {
		if strings.Contains(strings.ToLower(tmpl.Content), term) {
			found++
		}
	}

	if found < 2 {
		t.Error("security template should contain security-related terms")
	}
}
