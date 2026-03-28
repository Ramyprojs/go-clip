package clip

import (
	"fmt"
	"time"
)

// Clip represents a single clipboard entry stored by goclip.
type Clip struct {
	ID       uint64
	Content  string
	CopiedAt time.Time
	Source   string
}

// String returns a human-readable representation of a clipboard entry.
func (c Clip) String() string {
	return fmt.Sprintf("#%d [%s] %s (%s)", c.ID, c.CopiedAt.Format(time.RFC3339), c.Content, c.Source)
}
