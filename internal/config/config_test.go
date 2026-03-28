package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadReturnsDefaultsWhenConfigMissing(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.MaxHistory != defaultMaxHistory {
		t.Fatalf("MaxHistory = %d, want %d", cfg.MaxHistory, defaultMaxHistory)
	}

	if cfg.PreviewLength != defaultPreviewLength {
		t.Fatalf("PreviewLength = %d, want %d", cfg.PreviewLength, defaultPreviewLength)
	}

	wantDBPath := filepath.Join(homeDir, ".goclip", "history.db")
	if cfg.DBPath != wantDBPath {
		t.Fatalf("DBPath = %q, want %q", cfg.DBPath, wantDBPath)
	}
}

func TestLoadReadsConfigFile(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	configDir := filepath.Join(homeDir, ".goclip")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	configData := []byte("max_history: 250\npreview_length: 42\ndb_path: ~/.goclip/custom.db\n")
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, configData, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.MaxHistory != 250 {
		t.Fatalf("MaxHistory = %d, want 250", cfg.MaxHistory)
	}

	if cfg.PreviewLength != 42 {
		t.Fatalf("PreviewLength = %d, want 42", cfg.PreviewLength)
	}

	wantDBPath := filepath.Join(homeDir, ".goclip", "custom.db")
	if cfg.DBPath != wantDBPath {
		t.Fatalf("DBPath = %q, want %q", cfg.DBPath, wantDBPath)
	}
}
