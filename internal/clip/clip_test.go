package clip

import (
	"testing"
	"time"
)

func TestClipString(t *testing.T) {
	t.Parallel()

	copiedAt := time.Date(2026, time.March, 28, 12, 34, 56, 0, time.UTC)
	clip := Clip{
		ID:       42,
		Content:  "hello world",
		CopiedAt: copiedAt,
		Source:   "terminal",
	}

	got := clip.String()
	want := "#42 [2026-03-28T12:34:56Z] hello world (terminal)"
	if got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
}
