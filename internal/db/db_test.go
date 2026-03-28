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

func TestStoreGetAllClips(t *testing.T) {
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

	baseTime := time.Date(2026, time.March, 28, 15, 30, 0, 0, time.UTC)
	entries := []clip.Clip{
		{Content: "oldest", CopiedAt: baseTime, Source: "test"},
		{Content: "newest", CopiedAt: baseTime.Add(2 * time.Minute), Source: "test"},
		{Content: "middle", CopiedAt: baseTime.Add(time.Minute), Source: "test"},
	}

	for _, entry := range entries {
		if err := store.SaveClip(entry); err != nil {
			t.Fatalf("SaveClip() error = %v", err)
		}
	}

	got, err := store.GetAllClips()
	if err != nil {
		t.Fatalf("GetAllClips() error = %v", err)
	}

	if len(got) != 3 {
		t.Fatalf("GetAllClips() len = %d, want 3", len(got))
	}

	wantOrder := []string{"newest", "middle", "oldest"}
	for i, want := range wantOrder {
		if got[i].Content != want {
			t.Fatalf("GetAllClips()[%d].Content = %q, want %q", i, got[i].Content, want)
		}
	}
}

func TestStoreDeleteClip(t *testing.T) {
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

	baseTime := time.Date(2026, time.March, 28, 16, 0, 0, 0, time.UTC)
	entries := []clip.Clip{
		{ID: 101, Content: "keep me", CopiedAt: baseTime, Source: "test"},
		{ID: 202, Content: "delete me", CopiedAt: baseTime.Add(time.Minute), Source: "test"},
	}

	for _, entry := range entries {
		if err := store.SaveClip(entry); err != nil {
			t.Fatalf("SaveClip() error = %v", err)
		}
	}

	if err := store.DeleteClip(202); err != nil {
		t.Fatalf("DeleteClip() error = %v", err)
	}

	got, err := store.GetAllClips()
	if err != nil {
		t.Fatalf("GetAllClips() error = %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("GetAllClips() len = %d, want 1", len(got))
	}

	if got[0].ID != 101 {
		t.Fatalf("remaining clip ID = %d, want 101", got[0].ID)
	}
}

func TestStoreClearAll(t *testing.T) {
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

	baseTime := time.Date(2026, time.March, 28, 16, 15, 0, 0, time.UTC)
	entries := []clip.Clip{
		{Content: "first", CopiedAt: baseTime, Source: "test"},
		{Content: "second", CopiedAt: baseTime.Add(time.Minute), Source: "test"},
	}

	for _, entry := range entries {
		if err := store.SaveClip(entry); err != nil {
			t.Fatalf("SaveClip() error = %v", err)
		}
	}

	if err := store.ClearAll(); err != nil {
		t.Fatalf("ClearAll() error = %v", err)
	}

	got, err := store.GetAllClips()
	if err != nil {
		t.Fatalf("GetAllClips() error = %v", err)
	}

	if len(got) != 0 {
		t.Fatalf("GetAllClips() len = %d, want 0", len(got))
	}
}
