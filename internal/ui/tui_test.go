package ui

import (
	"errors"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Ramyprojs/goclip/internal/clip"
)

func TestModelFiltersClipsAsUserTypes(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.March, 28, 21, 45, 0, 0, time.UTC)
	m := newModel([]clip.Clip{
		{ID: 1, Content: "clip starts here", CopiedAt: now},
		{ID: 2, Content: "nothing useful", CopiedAt: now.Add(time.Minute)},
		{ID: 3, Content: "saved clip appears later", CopiedAt: now.Add(2 * time.Minute)},
	})

	m = typeQuery(t, m, "clip")

	if len(m.filtered) != 2 {
		t.Fatalf("filtered len = %d, want 2", len(m.filtered))
	}

	if m.filtered[0].ID != 1 {
		t.Fatalf("filtered[0].ID = %d, want 1", m.filtered[0].ID)
	}

	if m.filtered[1].ID != 3 {
		t.Fatalf("filtered[1].ID = %d, want 3", m.filtered[1].ID)
	}
}

func TestModelMovesSelectionDown(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.March, 28, 22, 0, 0, 0, time.UTC)
	m := newModel([]clip.Clip{
		{ID: 1, Content: "first", CopiedAt: now},
		{ID: 2, Content: "second", CopiedAt: now.Add(time.Minute)},
	})

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(model)

	if m.selected != 1 {
		t.Fatalf("selected = %d, want 1", m.selected)
	}
}

func TestModelCopiesSelectedClipOnEnter(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.March, 28, 22, 15, 0, 0, time.UTC)
	mockClipboard := &fakeClipboard{}
	m := newModelWithRuntime([]clip.Clip{
		{ID: 1, Content: "copy me", CopiedAt: now},
	}, nil, mockClipboard, 60)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)

	if mockClipboard.content != "copy me" {
		t.Fatalf("clipboard content = %q, want %q", mockClipboard.content, "copy me")
	}

	if m.status != "Copied selected clip." {
		t.Fatalf("status = %q, want %q", m.status, "Copied selected clip.")
	}
}

func TestModelDeletesSelectedClipOnUppercaseD(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.March, 28, 22, 30, 0, 0, time.UTC)
	store := &fakeStore{
		clips: []clip.Clip{
			{ID: 3, Content: "newest", CopiedAt: now.Add(2 * time.Minute)},
			{ID: 2, Content: "middle", CopiedAt: now.Add(time.Minute)},
			{ID: 1, Content: "oldest", CopiedAt: now},
		},
	}
	m := newModelWithRuntime(store.clips, store, nil, 60)
	m.selected = 1
	m.applyFilter()

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'D'}})
	m = updated.(model)

	if len(store.deleted) != 1 || store.deleted[0] != 2 {
		t.Fatalf("deleted IDs = %v, want [2]", store.deleted)
	}

	if len(m.clips) != 2 {
		t.Fatalf("clip count = %d, want 2", len(m.clips))
	}

	if m.status != "Deleted selected clip." {
		t.Fatalf("status = %q, want %q", m.status, "Deleted selected clip.")
	}
}

func TestModelQuitKeybindings(t *testing.T) {
	t.Parallel()

	m := newModel(nil)

	tests := []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune{'Q'}},
		{Type: tea.KeyCtrlC},
	}

	for _, key := range tests {
		updated, cmd := m.Update(key)
		m = updated.(model)
		if cmd == nil {
			t.Fatalf("Update(%v) cmd = nil, want tea.Quit", key)
		}

		if _, ok := cmd().(tea.QuitMsg); !ok {
			t.Fatalf("Update(%v) did not return tea.Quit", key)
		}
	}
}

func TestViewShowsStatusBarHints(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.March, 28, 22, 45, 0, 0, time.UTC)
	m := newModel([]clip.Clip{
		{ID: 1, Content: "first", CopiedAt: now},
	})

	view := m.View()
	if !strings.Contains(view, "Enter copy") {
		t.Fatalf("View() does not contain status hints: %q", view)
	}
}

func typeQuery(t *testing.T, m model, query string) model {
	t.Helper()

	for _, r := range query {
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(model)
	}

	return m
}

type fakeClipboard struct {
	content string
	err     error
}

func (f *fakeClipboard) WriteText(content string) error {
	if f.err != nil {
		return f.err
	}

	f.content = content
	return nil
}

type fakeStore struct {
	clips   []clip.Clip
	deleted []uint64
	getErr  error
	delErr  error
}

func (f *fakeStore) DeleteClip(id uint64) error {
	if f.delErr != nil {
		return f.delErr
	}

	f.deleted = append(f.deleted, id)

	next := make([]clip.Clip, 0, len(f.clips))
	found := false
	for _, entry := range f.clips {
		if entry.ID == id {
			found = true
			continue
		}

		next = append(next, entry)
	}

	if !found {
		return errors.New("clip not found")
	}

	f.clips = next
	return nil
}

func (f *fakeStore) GetAllClips() ([]clip.Clip, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}

	return append([]clip.Clip(nil), f.clips...), nil
}
