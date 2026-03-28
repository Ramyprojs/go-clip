package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Ramyprojs/goclip/internal/clip"
	clipsearch "github.com/Ramyprojs/goclip/internal/search"
)

const (
	defaultWidth        = 84
	defaultHeight       = 24
	defaultPreviewWidth = 64
)

type styles struct {
	title        lipgloss.Style
	searchBar    lipgloss.Style
	searchLabel  lipgloss.Style
	searchValue  lipgloss.Style
	searchHint   lipgloss.Style
	listPane     lipgloss.Style
	item         lipgloss.Style
	selectedItem lipgloss.Style
	meta         lipgloss.Style
	selectedMeta lipgloss.Style
	preview      lipgloss.Style
	selectedText lipgloss.Style
	emptyState   lipgloss.Style
}

type model struct {
	width    int
	height   int
	query    string
	clips    []clip.Clip
	filtered []clip.Clip
	selected int
	styles   styles
}

func newModel(clips []clip.Clip) model {
	m := model{
		width:  defaultWidth,
		height: defaultHeight,
		clips:  append([]clip.Clip(nil), clips...),
		styles: defaultStyles(),
	}
	m.applyFilter()

	return m
}

// Init satisfies the Bubble Tea model interface.
func (m model) Init() tea.Cmd {
	return nil
}

// Update satisfies the Bubble Tea model interface.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if msg.Width > 0 {
			m.width = msg.Width
		}

		if msg.Height > 0 {
			m.height = msg.Height
		}
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp:
			m.moveSelection(-1)
		case tea.KeyDown:
			m.moveSelection(1)
		case tea.KeyBackspace, tea.KeyCtrlH:
			m.removeLastRune()
			m.applyFilter()
		default:
			if len(msg.Runes) > 0 {
				m.query += string(msg.Runes)
				m.applyFilter()
			}
		}
	}

	return m, nil
}

// View satisfies the Bubble Tea model interface.
func (m model) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.styles.title.Render("goclip"),
		m.renderSearchBar(),
		m.renderListPane(),
	)
}

func defaultStyles() styles {
	return styles{
		title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("87")).
			Padding(0, 1),
		searchBar: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(0, 1),
		searchLabel: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("252")),
		searchValue: lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")),
		searchHint: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),
		listPane: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(0, 1),
		item: lipgloss.NewStyle().
			Padding(0, 1),
		selectedItem: lipgloss.NewStyle().
			Padding(0, 1).
			BorderLeft(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("51")).
			Foreground(lipgloss.Color("87")).
			Bold(true),
		meta: lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")),
		selectedMeta: lipgloss.NewStyle().
			Foreground(lipgloss.Color("123")),
		preview: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
		selectedText: lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")),
		emptyState: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true).
			Padding(1, 0),
	}
}

func (m *model) applyFilter() {
	m.filtered = clipsearch.FuzzySearch(m.query, m.clips)
	if len(m.filtered) == 0 {
		m.selected = 0
		return
	}

	if m.selected >= len(m.filtered) {
		m.selected = len(m.filtered) - 1
	}

	if m.selected < 0 {
		m.selected = 0
	}
}

func (m *model) moveSelection(delta int) {
	if len(m.filtered) == 0 {
		return
	}

	m.selected += delta
	if m.selected < 0 {
		m.selected = 0
	}

	if m.selected >= len(m.filtered) {
		m.selected = len(m.filtered) - 1
	}
}

func (m *model) removeLastRune() {
	runes := []rune(m.query)
	if len(runes) == 0 {
		return
	}

	m.query = string(runes[:len(runes)-1])
}

func (m model) renderSearchBar() string {
	value := m.styles.searchHint.Render("type to filter clipboard history")
	if strings.TrimSpace(m.query) != "" {
		value = m.styles.searchValue.Render(m.query)
	}

	content := lipgloss.JoinHorizontal(
		lipgloss.Left,
		m.styles.searchLabel.Render("Search: "),
		value,
	)

	return m.styles.searchBar.Width(m.contentWidth()).Render(content)
}

func (m model) renderListPane() string {
	if len(m.filtered) == 0 {
		return m.styles.listPane.
			Width(m.contentWidth()).
			Render(m.styles.emptyState.Render("No clips match your search."))
	}

	items := make([]string, 0, len(m.visibleClips()))
	for index, entry := range m.visibleClips() {
		actualIndex := m.visibleStart() + index
		items = append(items, m.renderClip(actualIndex, entry))
	}

	return m.styles.listPane.Width(m.contentWidth()).Render(
		lipgloss.JoinVertical(lipgloss.Left, items...),
	)
}

func (m model) renderClip(index int, entry clip.Clip) string {
	meta := entry.CopiedAt.Format("2006-01-02 15:04:05")
	if entry.Source != "" {
		meta += "  " + entry.Source
	}

	body := lipgloss.JoinVertical(
		lipgloss.Left,
		m.styles.meta.Render(meta),
		m.styles.preview.Render(m.previewContent(entry.Content)),
	)

	if index == m.selected {
		body = lipgloss.JoinVertical(
			lipgloss.Left,
			m.styles.selectedMeta.Render(meta),
			m.styles.selectedText.Render(m.previewContent(entry.Content)),
		)

		return m.styles.selectedItem.Render(body)
	}

	return m.styles.item.Render(body)
}

func (m model) previewContent(content string) string {
	compact := strings.Join(strings.Fields(content), " ")
	if compact == "" {
		return "(empty clip)"
	}

	runes := []rune(compact)
	limit := m.previewWidth()
	if len(runes) <= limit {
		return compact
	}

	if limit <= 3 {
		return string(runes[:limit])
	}

	return string(runes[:limit-3]) + "..."
}

func (m model) previewWidth() int {
	width := defaultPreviewWidth
	if m.width > 0 {
		width = m.width - 12
	}

	if width < 24 {
		return 24
	}

	return width
}

func (m model) contentWidth() int {
	if m.width <= 0 {
		return defaultWidth
	}

	if m.width < 40 {
		return 40
	}

	return m.width - 4
}

func (m model) visibleClips() []clip.Clip {
	if len(m.filtered) == 0 {
		return nil
	}

	start := m.visibleStart()
	end := start + m.maxVisibleItems()
	if end > len(m.filtered) {
		end = len(m.filtered)
	}

	return m.filtered[start:end]
}

func (m model) visibleStart() int {
	maxItems := m.maxVisibleItems()
	if len(m.filtered) <= maxItems {
		return 0
	}

	start := m.selected - (maxItems / 2)
	if start < 0 {
		start = 0
	}

	maxStart := len(m.filtered) - maxItems
	if start > maxStart {
		start = maxStart
	}

	return start
}

func (m model) maxVisibleItems() int {
	if m.height <= 0 {
		return 8
	}

	available := m.height - 8
	if available < 3 {
		return 3
	}

	items := available / 3
	if items < 1 {
		return 1
	}

	return items
}
