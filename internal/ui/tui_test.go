package ui

import (
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

func typeQuery(t *testing.T, m model, query string) model {
	t.Helper()

	for _, r := range query {
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(model)
	}

	return m
}
