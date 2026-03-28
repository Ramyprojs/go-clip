package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/Ramyprojs/goclip/internal/clip"
	"go.etcd.io/bbolt"
)

const clipsBucket = "clips"

// Store provides access to the local clipboard history database.
type Store struct {
	db *bbolt.DB
}

// OpenDB opens the clipboard history database and ensures the clips bucket exists.
func OpenDB(path string) (*Store, error) {
	resolvedPath, err := resolveDBPath(path)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(filepath.Dir(resolvedPath), 0o755); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	boltDB, err := bbolt.Open(resolvedPath, 0o600, &bbolt.Options{Timeout: time.Second})
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := boltDB.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(clipsBucket))
		return err
	}); err != nil {
		_ = boltDB.Close()
		return nil, fmt.Errorf("initialize clips bucket: %w", err)
	}

	return &Store{db: boltDB}, nil
}

// CloseDB closes the underlying clipboard history database.
func (s *Store) CloseDB() error {
	if s == nil {
		return errors.New("db store is nil")
	}

	if s.db == nil {
		return errors.New("db is not open")
	}

	if err := s.db.Close(); err != nil {
		return fmt.Errorf("close db: %w", err)
	}

	s.db = nil
	return nil
}

// SaveClip persists a clipboard entry in the clips bucket.
func (s *Store) SaveClip(entry clip.Clip) error {
	if s == nil {
		return errors.New("db store is nil")
	}

	if s.db == nil {
		return errors.New("db is not open")
	}

	if entry.CopiedAt.IsZero() {
		entry.CopiedAt = time.Now().UTC()
	}

	if entry.ID == 0 {
		entry.ID = uint64(entry.CopiedAt.UnixNano())
	}

	payload, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal clip: %w", err)
	}

	key := []byte(strconv.FormatInt(entry.CopiedAt.UnixNano(), 10))
	if err := s.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(clipsBucket))
		if bucket == nil {
			return errors.New("clips bucket not found")
		}

		return bucket.Put(key, payload)
	}); err != nil {
		return fmt.Errorf("save clip: %w", err)
	}

	return nil
}

// GetAllClips returns all stored clipboard entries sorted from newest to oldest.
func (s *Store) GetAllClips() ([]clip.Clip, error) {
	if s == nil {
		return nil, errors.New("db store is nil")
	}

	if s.db == nil {
		return nil, errors.New("db is not open")
	}

	clips := make([]clip.Clip, 0)
	if err := s.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(clipsBucket))
		if bucket == nil {
			return errors.New("clips bucket not found")
		}

		return bucket.ForEach(func(_, value []byte) error {
			var entry clip.Clip
			if err := json.Unmarshal(value, &entry); err != nil {
				return fmt.Errorf("unmarshal clip: %w", err)
			}

			clips = append(clips, entry)
			return nil
		})
	}); err != nil {
		return nil, fmt.Errorf("get all clips: %w", err)
	}

	sort.Slice(clips, func(i, j int) bool {
		if clips[i].CopiedAt.Equal(clips[j].CopiedAt) {
			return clips[i].ID > clips[j].ID
		}

		return clips[i].CopiedAt.After(clips[j].CopiedAt)
	})

	return clips, nil
}

// DeleteClip removes a clipboard entry by its ID.
func (s *Store) DeleteClip(id uint64) error {
	if s == nil {
		return errors.New("db store is nil")
	}

	if s.db == nil {
		return errors.New("db is not open")
	}

	if err := s.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(clipsBucket))
		if bucket == nil {
			return errors.New("clips bucket not found")
		}

		cursor := bucket.Cursor()
		for key, value := cursor.First(); key != nil; key, value = cursor.Next() {
			var entry clip.Clip
			if err := json.Unmarshal(value, &entry); err != nil {
				return fmt.Errorf("unmarshal clip: %w", err)
			}

			if entry.ID == id {
				if err := bucket.Delete(key); err != nil {
					return fmt.Errorf("delete clip: %w", err)
				}

				return nil
			}
		}

		return fmt.Errorf("clip with id %d not found", id)
	}); err != nil {
		return fmt.Errorf("delete clip: %w", err)
	}

	return nil
}

// ClearAll removes all clipboard entries from the clips bucket.
func (s *Store) ClearAll() error {
	if s == nil {
		return errors.New("db store is nil")
	}

	if s.db == nil {
		return errors.New("db is not open")
	}

	if err := s.db.Update(func(tx *bbolt.Tx) error {
		if err := tx.DeleteBucket([]byte(clipsBucket)); err != nil {
			return fmt.Errorf("delete clips bucket: %w", err)
		}

		if _, err := tx.CreateBucket([]byte(clipsBucket)); err != nil {
			return fmt.Errorf("recreate clips bucket: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("clear all clips: %w", err)
	}

	return nil
}

func resolveDBPath(path string) (string, error) {
	if path != "" {
		return path, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}

	return filepath.Join(homeDir, ".goclip", "history.db"), nil
}
