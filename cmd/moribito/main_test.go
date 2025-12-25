package main

import (
	"testing"

	"github.com/pirakansa/moribito/internal/config"
)

func TestPrintInstallationTokenMissingConfig(t *testing.T) {
	err := printInstallationToken(config.Config{})
	if err == nil {
		t.Fatalf("expected error for missing config")
	}
}
