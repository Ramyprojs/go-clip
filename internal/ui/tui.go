package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	sysclip "golang.design/x/clipboard"

	"github.com/Ramyprojs/goclip/internal/clip"
	"github.com/Ramyprojs/goclip/internal/config"
	"github.com/Ramyprojs/goclip/internal/db"
	clipsearch "github.com/Ramyprojs/goclip/internal/search"
)

const (
	defaultWidth      = 84
	defaultHeight     = 24
	browseStatusHints = "A add | Enter copy | D delete | Q quit"
	addModeHints      = "Enter save | Esc cancel | Ctrl+C quit"
)

type styles struct {
	title         lipgloss.Style
	searchBar     lipgloss.Style
	searchLabel   lipgloss.Style
	searchValue   lipgloss.Style
	searchHint    lipgloss.Style
	listPane      lipgloss.Style
	item          lipgloss.Style
	selectedItem  lipgloss.Style
	meta          lipgloss.Style
	selectedMeta  lipgloss.Style
	preview       lipgloss.Style
	selectedText  lipgloss.Style
	emptyState    lipgloss.Style
	statusBar     lipgloss.Style
	statusMessage lipgloss.Style
	statusHints   lipgloss.Style
}

type clipStore interface {
	DeleteClip(id uint64) error
	GetAllClips() ([]clip.Clip, error)
	SaveClip(entry clip.Clip) error
}

type clipboardWriter interface {
	WriteText(content string) error
}

type systemClipboard struct{}

type model struct {
	width      int
	height     int
	query      string
	draft      string
	status     string
	clips      []clip.Clip
	filtered   []clip.Clip
	selected   int
	adding     bool
	store      clipStore
	clipboard  clipboardWriter
	preview    int
	maxHistory int
	styles     styles
}

func newModel(clips []clip.Clip) model {
	cfg := config.DefaultConfig()
	return newModelWithRuntime(clips, nil, nil, cfg.PreviewLength, cfg.MaxHistory)
}

func newModelWithRuntime(clips []clip.Clip, store clipStore, clipboard clipboardWriter, previewLength int, maxHistory int) model {
	if previewLength <= 0 {
		previewLength = config.DefaultConfig().PreviewLength
	}

	m := model{
		width:      defaultWidth,
		height:     defaultHeight,
		status:     "Ready.",
		clips:      append([]clip.Clip(nil), clips...),
		store:      store,
		clipboard:  clipboard,
		preview:    previewLength,
		maxHistory: maxHistory,
		styles:     defaultStyles(),
	}
	m.applyFilter()

	return m
}

func newSystemClipboard() (clipboardWriter, error) {
	if err := sysclip.Init(); err != nil {
		return nil, fmt.Errorf("initialize clipboard: %w", err)
	}

	return systemClipboard{}, nil
}

func (systemClipboard) WriteText(content string) error {
	sysclip.Write(sysclip.FmtText, []byte(content))
	return nil
}

// StartTUI loads clipboard history and launches the Bubble Tea interface.
func StartTUI(cfg config.Config) error {
	store, err := db.OpenDB(cfg.DBPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = store.CloseDB()
	}()

	clips, err := store.GetAllClips()
	if err != nil {
		return err
	}

	clipboard, clipboardErr := newSystemClipboard()
	model := newModelWithRuntime(clips, store, clipboard, cfg.PreviewLength, cfg.MaxHistory)
	if clipboardErr != nil {
		model.status = fmt.Sprintf("Clipboard unavailable: %v", clipboardErr)
	}

	if _, err := tea.NewProgram(model, tea.WithAltScreen()).Run(); err != nil {
		return fmt.Errorf("run tui: %w", err)
	}

	return nil
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
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}

		if m.adding {
			m.handleAddModeKey(msg)
			return m, nil
		}

		if len(msg.Runes) == 1 {
			switch msg.Runes[0] {
			case 'a', 'A':
				m.startAddMode()
				return m, nil
			case 'Q':
				fallthrough
			case 'q':
				return m, tea.Quit
			case 'D':
				fallthrough
			case 'd':
				m.deleteSelected()
				return m, nil
			}
		}

		switch msg.Type {
		case tea.KeyUp:
			m.moveSelection(-1)
		case tea.KeyDown:
			m.moveSelection(1)
		case tea.KeyEnter:
			m.copySelected()
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
		m.renderStatusBar(),
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
		statusBar: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(0, 1),
		statusMessage: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
		statusHints: lipgloss.NewStyle().
			Foreground(lipgloss.Color("87")),
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

func (m *model) copySelected() {
	entry, ok := m.currentSelection()
	if !ok {
		m.status = "No clip selected."
		return
	}

	if m.clipboard == nil {
		m.status = "Clipboard unavailable."
		return
	}

	if err := m.clipboard.WriteText(entry.Content); err != nil {
		m.status = fmt.Sprintf("Copy failed: %v", err)
		return
	}

	m.status = "Copied selected clip."
}

func (m *model) handleAddModeKey(msg tea.KeyMsg) {
	switch msg.Type {
	case tea.KeyEsc:
		m.cancelAddMode()
	case tea.KeyEnter:
		m.saveDraft()
	case tea.KeyBackspace, tea.KeyCtrlH:
		m.removeLastDraftRune()
	default:
		if len(msg.Runes) > 0 {
			m.draft += string(msg.Runes)
		}
	}
}

func (m *model) startAddMode() {
	m.adding = true
	m.draft = ""
	m.status = "Add mode: type a new clip and press Enter to save."
}

func (m *model) cancelAddMode() {
	m.adding = false
	m.draft = ""
	m.status = "Add cancelled."
}

func (m *model) removeLastDraftRune() {
	runes := []rune(m.draft)
	if len(runes) == 0 {
		return
	}

	m.draft = string(runes[:len(runes)-1])
}

func (m *model) saveDraft() {
	if strings.TrimSpace(m.draft) == "" {
		m.status = "Clip content cannot be empty."
		return
	}

	if m.store == nil {
		m.status = "Add unavailable."
		return
	}

	entry := clip.Clip{
		Content: strings.TrimSpace(m.draft),
		Source:  "ui",
	}

	if err := m.store.SaveClip(entry); err != nil {
		m.status = fmt.Sprintf("Save failed: %v", err)
		return
	}

	if err := m.trimToMaxHistory(); err != nil {
		m.status = fmt.Sprintf("Saved clip, but trimming failed: %v", err)
	}

	clips, err := m.store.GetAllClips()
	if err != nil {
		m.adding = false
		m.draft = ""
		m.status = fmt.Sprintf("Saved clip, but refresh failed: %v", err)
		return
	}

	m.clips = clips
	m.query = ""
	m.adding = false
	m.draft = ""
	m.applyFilter()
	m.status = "Saved new clip."
}

func (m *model) deleteSelected() {
	entry, ok := m.currentSelection()
	if !ok {
		m.status = "No clip selected."
		return
	}

	if m.store == nil {
		m.status = "Delete unavailable."
		return
	}

	if err := m.store.DeleteClip(entry.ID); err != nil {
		m.status = fmt.Sprintf("Delete failed: %v", err)
		return
	}

	clips, err := m.store.GetAllClips()
	if err != nil {
		m.removeClipByID(entry.ID)
		m.applyFilter()
		m.status = fmt.Sprintf("Deleted clip, but refresh failed: %v", err)
		return
	}

	m.clips = clips
	m.applyFilter()
	m.status = "Deleted selected clip."
}

func (m *model) currentSelection() (clip.Clip, bool) {
	if len(m.filtered) == 0 || m.selected < 0 || m.selected >= len(m.filtered) {
		return clip.Clip{}, false
	}

	return m.filtered[m.selected], true
}

func (m *model) removeClipByID(id uint64) {
	next := make([]clip.Clip, 0, len(m.clips))
	for _, entry := range m.clips {
		if entry.ID == id {
			continue
		}

		next = append(next, entry)
	}

	m.clips = next
}

func (m *model) trimToMaxHistory() error {
	if m.maxHistory <= 0 || m.store == nil {
		return nil
	}

	clips, err := m.store.GetAllClips()
	if err != nil {
		return fmt.Errorf("load clips for max history: %w", err)
	}

	for i := m.maxHistory; i < len(clips); i++ {
		if err := m.store.DeleteClip(clips[i].ID); err != nil {
			return fmt.Errorf("trim history: %w", err)
		}
	}

	return nil
}

func (m model) renderSearchBar() string {
	label := "Search: "
	value := m.styles.searchHint.Render("type to filter clipboard history")

	if m.adding {
		label = "New Clip: "
		value = m.styles.searchHint.Render("type a new clip and press Enter")
		if strings.TrimSpace(m.draft) != "" {
			value = m.styles.searchValue.Render(m.draft)
		}
	} else if strings.TrimSpace(m.query) != "" {
		value = m.styles.searchValue.Render(m.query)
	}

	content := lipgloss.JoinHorizontal(
		lipgloss.Left,
		m.styles.searchLabel.Render(label),
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

func (m model) renderStatusBar() string {
	left := m.styles.statusMessage.Render(m.status)
	right := m.styles.statusHints.Render(m.statusHints())
	width := m.contentWidth()
	innerWidth := width - 4
	if innerWidth < 24 {
		innerWidth = 24
	}

	content := left
	remaining := innerWidth - lipgloss.Width(left) - lipgloss.Width(right)
	if remaining > 1 {
		content = left + strings.Repeat(" ", remaining) + right
	} else {
		content = lipgloss.JoinVertical(lipgloss.Left, left, right)
	}

	return m.styles.statusBar.Width(width).Render(content)
}

func (m model) statusHints() string {
	if m.adding {
		return addModeHints
	}

	return browseStatusHints
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
	width := m.preview
	if width <= 0 {
		width = config.DefaultConfig().PreviewLength
	}

	if m.width > 0 {
		maxWidth := m.width - 12
		if maxWidth < 8 {
			maxWidth = 8
		}

		if width > maxWidth {
			width = maxWidth
		}
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

	available := m.height - 11
	if available < 3 {
		return 3
	}

	items := available / 3
	if items < 1 {
		return 1
	}

	return items
}
