package db

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/Ramyprojs/goclip/internal/clip"
	"go.etcd.io/bbolt"
)

func TestStoreSaveClip(t *testing.T) {
	t.Parallel()

	store, err := OpenDB(filepath.Join(t.TempDir(), "history.db"))
	if err != nil {
		t.Fatalf("OpenDB() error = %v", err)
	}
	t.Cleanup(func() {
		if err := store.CloseDB(); err != nil {
			t.Fatalf("CloseDB() error = %v", err)
		}
	})

	entry := clip.Clip{
		Content:  "hello world",
		CopiedAt: time.Date(2026, time.March, 28, 15, 4, 5, 123456789, time.UTC),
		Source:   "test",
	}

	if err := store.SaveClip(entry); err != nil {
		t.Fatalf("SaveClip() error = %v", err)
	}

	var count int
	if err := store.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(clipsBucket))
		if bucket == nil {
			return errors.New("clips bucket not found")
		}

		return bucket.ForEach(func(_, _ []byte) error {
			count++
			return nil
		})
	}); err != nil {
		t.Fatalf("View() error = %v", err)
	}

	if count != 1 {
		t.Fatalf("saved clip count = %d, want 1", count)
	}
}
