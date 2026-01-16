package prompt

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadTemplateFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pr.tmpl")

	if err := os.WriteFile(path, []byte("Title: {{.Title}}"), 0o644); err != nil {
		t.Fatalf("write template: %v", err)
	}

	tmpl, err := LoadTemplateFromFile(path)
	if err != nil {
		t.Fatalf("LoadTemplateFromFile failed: %v", err)
	}
	if tmpl.Name != "pr.tmpl" {
		t.Fatalf("expected template name pr.tmpl, got %s", tmpl.Name)
	}
	if tmpl.Content != "Title: {{.Title}}" {
		t.Fatalf("unexpected template content: %q", tmpl.Content)
	}
}

func TestLoadTemplateFromFileEmptyPath(t *testing.T) {
	if _, err := LoadTemplateFromFile(""); err == nil {
		t.Fatalf("expected error for empty template path")
	}
}
