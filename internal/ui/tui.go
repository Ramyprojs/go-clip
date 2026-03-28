package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	container lipgloss.Style
}

func newModel() model {
	return model{
		container: lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Foreground(lipgloss.Color("252")),
	}
}

// Init satisfies the Bubble Tea model interface.
func (m model) Init() tea.Cmd {
	return nil
}

// Update satisfies the Bubble Tea model interface.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

// View satisfies the Bubble Tea model interface.
func (m model) View() string {
	return m.container.Render("goclip TUI — loading...")
}
