package search

import (
	"testing"
	"time"

	"github.com/Ramyprojs/goclip/internal/clip"
)

func TestFuzzySearchEmptyQueryReturnsAllClips(t *testing.T) {
	t.Parallel()

	clips := []clip.Clip{
		{ID: 1, Content: "first clip"},
		{ID: 2, Content: "second clip"},
	}

	got := FuzzySearch("", clips)

	if len(got) != len(clips) {
		t.Fatalf("FuzzySearch() len = %d, want %d", len(got), len(clips))
	}

	for i := range clips {
		if got[i].ID != clips[i].ID {
			t.Fatalf("FuzzySearch()[%d].ID = %d, want %d", i, got[i].ID, clips[i].ID)
		}
	}
}

func TestFuzzySearchNoMatch(t *testing.T) {
	t.Parallel()

	clips := []clip.Clip{
		{ID: 1, Content: "alpha"},
		{ID: 2, Content: "beta"},
	}

	got := FuzzySearch("zeta", clips)
	if len(got) != 0 {
		t.Fatalf("FuzzySearch() len = %d, want 0", len(got))
	}
}

func TestFuzzySearchPartialMatchOrdersByPosition(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.March, 28, 19, 45, 0, 0, time.UTC)
	clips := []clip.Clip{
		{ID: 1, Content: "clip at the start", CopiedAt: now},
		{ID: 2, Content: "saved terminal clip appears later", CopiedAt: now.Add(time.Minute)},
		{ID: 3, Content: "nothing relevant here", CopiedAt: now.Add(2 * time.Minute)},
	}

	got := FuzzySearch("clip", clips)

	if len(got) != 2 {
		t.Fatalf("FuzzySearch() len = %d, want 2", len(got))
	}

	if got[0].ID != 1 {
		t.Fatalf("FuzzySearch()[0].ID = %d, want 1", got[0].ID)
	}

	if got[1].ID != 2 {
		t.Fatalf("FuzzySearch()[1].ID = %d, want 2", got[1].ID)
	}
}
