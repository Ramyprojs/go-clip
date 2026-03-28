package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRemoveApplicationDataRemovesCustomDatabaseOutsideDataDir(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	dataDir := filepath.Join(tempDir, ".goclip")
	customDBPath := filepath.Join(tempDir, "history.db")

	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatalf("mkdir data dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dataDir, "config.yaml"), []byte("preview_length: 42\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	if err := os.WriteFile(customDBPath, []byte("db"), 0o644); err != nil {
		t.Fatalf("write custom db: %v", err)
	}

	if err := removeApplicationData(dataDir, customDBPath); err != nil {
		t.Fatalf("removeApplicationData() error = %v", err)
	}

	if _, err := os.Stat(dataDir); !os.IsNotExist(err) {
		t.Fatalf("data dir still exists, stat err = %v", err)
	}

	if _, err := os.Stat(customDBPath); !os.IsNotExist(err) {
		t.Fatalf("custom db still exists, stat err = %v", err)
	}
}

func TestRemoveCustomDatabaseKeepsDatabaseInsideDataDir(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	dataDir := filepath.Join(tempDir, ".goclip")
	databasePath := filepath.Join(dataDir, "history.db")

	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatalf("mkdir data dir: %v", err)
	}

	if err := os.WriteFile(databasePath, []byte("db"), 0o644); err != nil {
		t.Fatalf("write db: %v", err)
	}

	if err := removeCustomDatabase(databasePath, dataDir); err != nil {
		t.Fatalf("removeCustomDatabase() error = %v", err)
	}

	if _, err := os.Stat(databasePath); err != nil {
		t.Fatalf("database inside data dir should remain, stat err = %v", err)
	}
}
