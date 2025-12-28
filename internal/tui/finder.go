package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/whisller/pkit/internal/bookmark"
	"github.com/whisller/pkit/internal/tag"
	"github.com/whisller/pkit/pkg/models"
)

// PanelType represents which panel is currently focused
type PanelType int

const (
	PanelFilters PanelType = iota
	PanelList
)

// FilterSection represents different filter sections
type FilterSection int

const (
	FilterSources FilterSection = iota
	FilterTags
	FilterBookmarked
)

// KeyMap defines keyboard shortcuts
type KeyMap struct {
	Quit         key.Binding
	Select       key.Binding
	Get          key.Binding
	Bookmark     key.Binding
	Tag          key.Binding
	Up           key.Binding
	Down         key.Binding
	SwitchPanel  key.Binding
	ToggleFilter key.Binding
}

// DefaultKeyMap returns the default key bindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit: key.NewBinding(
			key.WithKeys("q", "esc", "ctrl+c"),
			key.WithHelp("q/esc", "quit"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Get: key.NewBinding(
			key.WithKeys("ctrl+g"),
			key.WithHelp("ctrl+g", "get prompt"),
		),
		Bookmark: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "bookmark"),
		),
		Tag: key.NewBinding(
			key.WithKeys("ctrl+t"),
			key.WithHelp("ctrl+t", "add tags"),
		),
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		SwitchPanel: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch panel"),
		),
		ToggleFilter: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "toggle filter"),
		),
	}
}

// PromptItem wraps a prompt for display in the list
type PromptItem struct {
	Prompt models.Prompt
}

func (i PromptItem) Title() string       { return i.Prompt.ID }
func (i PromptItem) Description() string { return i.Prompt.Description }
func (i PromptItem) FilterValue() string {
	return i.Prompt.ID + " " + i.Prompt.Name + " " + i.Prompt.Description
}

// FinderModel is the Bubbletea model for the interactive finder
type FinderModel struct {
	list             list.Model
	keys             KeyMap
	allPrompts       []models.Prompt
	filteredPrompts  []models.Prompt
	width            int
	height           int
	activePanel      PanelType
	filterCursor     int
	filterSection    FilterSection

	// Filter state
	availableSources []string
	selectedSources  map[string]bool
	availableTags    []string
	selectedTags     map[string]bool
	showBookmarked   bool
	bookmarkedIDs    map[string]bool

	// Action state
	selectedID     string
	actionGet      bool
	actionBookmark bool
	actionTag      bool
	quitting       bool
}

// NewFinderModel creates a new finder model
func NewFinderModel(prompts []models.Prompt) FinderModel {
	// Extract available sources
	sourceSet := make(map[string]bool)
	for _, p := range prompts {
		parts := strings.Split(p.ID, ":")
		if len(parts) > 0 {
			sourceSet[parts[0]] = true
		}
	}
	sources := make([]string, 0, len(sourceSet))
	for source := range sourceSet {
		sources = append(sources, source)
	}
	sort.Strings(sources)

	// Extract available tags from user tags
	tagSet := make(map[string]bool)
	tagMgr := tag.NewManager()
	allTags, _ := tagMgr.ListAllTags()
	for _, pt := range allTags {
		for _, t := range pt.Tags {
			tagSet[t] = true
		}
	}
	tags := make([]string, 0, len(tagSet))
	for t := range tagSet {
		tags = append(tags, t)
	}
	sort.Strings(tags)

	// Get bookmarked IDs
	bookmarkedIDs := make(map[string]bool)
	bookmarks, _ := bookmark.LoadBookmarks()
	for _, bm := range bookmarks {
		bookmarkedIDs[bm.PromptID] = true
	}

	// Initialize with all sources selected
	selectedSources := make(map[string]bool)
	for _, source := range sources {
		selectedSources[source] = true
	}

	// Create list (will be populated with filtered items)
	delegate := list.NewDefaultDelegate()
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "Prompts"
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true)
	l.SetShowStatusBar(true)

	m := FinderModel{
		list:             l,
		keys:             DefaultKeyMap(),
		allPrompts:       prompts,
		activePanel:      PanelFilters,
		filterCursor:     0,
		filterSection:    FilterSources,
		availableSources: sources,
		selectedSources:  selectedSources,
		availableTags:    tags,
		selectedTags:     make(map[string]bool),
		showBookmarked:   false,
		bookmarkedIDs:    bookmarkedIDs,
	}

	// Apply initial filtering
	m.applyFilters()

	return m
}

// applyFilters filters prompts based on selected filters
func (m *FinderModel) applyFilters() {
	// Get user tags for filtering
	tagMgr := tag.NewManager()
	promptTags, _ := tagMgr.ListAllTags()
	promptTagMap := make(map[string][]string)
	for _, pt := range promptTags {
		promptTagMap[pt.PromptID] = pt.Tags
	}

	filtered := make([]models.Prompt, 0)
	for _, p := range m.allPrompts {
		// Filter by source
		sourceID := strings.Split(p.ID, ":")[0]
		if !m.selectedSources[sourceID] {
			continue
		}

		// Filter by bookmarked
		if m.showBookmarked && !m.bookmarkedIDs[p.ID] {
			continue
		}

		// Filter by tags (if any tags selected)
		if len(m.selectedTags) > 0 {
			userTags := promptTagMap[p.ID]
			hasSelectedTag := false
			for _, t := range userTags {
				if m.selectedTags[t] {
					hasSelectedTag = true
					break
				}
			}
			if !hasSelectedTag {
				continue
			}
		}

		filtered = append(filtered, p)
	}

	m.filteredPrompts = filtered

	// Update list items
	items := make([]list.Item, len(filtered))
	for i, p := range filtered {
		items[i] = PromptItem{Prompt: p}
	}
	m.list.SetItems(items)
}

// Init initializes the model
func (m FinderModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m FinderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Split width: 30% for filters, 70% for list
		filterWidth := int(float64(msg.Width) * 0.3)
		listWidth := msg.Width - filterWidth - 2

		m.list.SetWidth(listWidth)
		m.list.SetHeight(msg.Height - 6)
		return m, nil

	case tea.KeyMsg:
		// Global shortcuts
		switch {
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, m.keys.SwitchPanel):
			// Switch between filters and list panels
			if m.activePanel == PanelFilters {
				m.activePanel = PanelList
			} else {
				m.activePanel = PanelFilters
			}
			return m, nil
		}

		// Panel-specific shortcuts
		if m.activePanel == PanelFilters {
			return m.updateFiltersPanel(msg)
		} else {
			return m.updateListPanel(msg)
		}
	}

	// Update list if in list panel
	if m.activePanel == PanelList {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	return m, nil
}

// updateFiltersPanel handles input when filters panel is active
func (m FinderModel) updateFiltersPanel(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Up):
		m.filterCursor--
		if m.filterCursor < 0 {
			m.filterCursor = m.getFilterItemCount() - 1
		}
		m.updateFilterSection()
		return m, nil

	case key.Matches(msg, m.keys.Down):
		m.filterCursor++
		if m.filterCursor >= m.getFilterItemCount() {
			m.filterCursor = 0
		}
		m.updateFilterSection()
		return m, nil

	case key.Matches(msg, m.keys.ToggleFilter):
		m.toggleCurrentFilter()
		m.applyFilters()
		return m, nil
	}

	return m, nil
}

// updateListPanel handles input when list panel is active
func (m FinderModel) updateListPanel(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Don't match any key bindings while filtering
	if m.list.FilterState() == list.Filtering {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	switch {
	case key.Matches(msg, m.keys.Select):
		if item, ok := m.list.SelectedItem().(PromptItem); ok {
			m.selectedID = item.Prompt.ID
			m.quitting = true
			return m, tea.Quit
		}

	case key.Matches(msg, m.keys.Get):
		if item, ok := m.list.SelectedItem().(PromptItem); ok {
			m.selectedID = item.Prompt.ID
			m.actionGet = true
			m.quitting = true
			return m, tea.Quit
		}

	case key.Matches(msg, m.keys.Bookmark):
		if item, ok := m.list.SelectedItem().(PromptItem); ok {
			m.selectedID = item.Prompt.ID
			m.actionBookmark = true
			m.quitting = true
			return m, tea.Quit
		}

	case key.Matches(msg, m.keys.Tag):
		if item, ok := m.list.SelectedItem().(PromptItem); ok {
			m.selectedID = item.Prompt.ID
			m.actionTag = true
			m.quitting = true
			return m, tea.Quit
		}
	}

	// Update list
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// getFilterItemCount returns total number of filterable items
func (m *FinderModel) getFilterItemCount() int {
	// Sources
	count := len(m.availableSources)

	// Tags (only add if we have tags)
	if len(m.availableTags) > 0 {
		count += len(m.availableTags)
	}

	// Bookmarked toggle
	count += 1

	return count
}

// updateFilterSection updates which filter section we're in based on cursor
func (m *FinderModel) updateFilterSection() {
	cursor := 0

	// Sources section
	if m.filterCursor < len(m.availableSources) {
		m.filterSection = FilterSources
		return
	}
	cursor += len(m.availableSources)

	// Tags section (only if we have tags)
	if len(m.availableTags) > 0 {
		if m.filterCursor < cursor+len(m.availableTags) {
			m.filterSection = FilterTags
			return
		}
		cursor += len(m.availableTags)
	}

	// Bookmarked section
	m.filterSection = FilterBookmarked
}

// toggleCurrentFilter toggles the currently selected filter
func (m *FinderModel) toggleCurrentFilter() {
	cursor := 0

	// Sources section
	if m.filterCursor < len(m.availableSources) {
		source := m.availableSources[m.filterCursor]
		m.selectedSources[source] = !m.selectedSources[source]
		return
	}
	cursor += len(m.availableSources)

	// Tags section (only if we have tags)
	if len(m.availableTags) > 0 {
		if m.filterCursor < cursor+len(m.availableTags) {
			tagIdx := m.filterCursor - cursor
			tag := m.availableTags[tagIdx]
			m.selectedTags[tag] = !m.selectedTags[tag]
			return
		}
		cursor += len(m.availableTags)
	}

	// Bookmarked toggle
	m.showBookmarked = !m.showBookmarked
}

// View renders the UI
func (m FinderModel) View() string {
	if m.quitting {
		return ""
	}

	// Styles
	activeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		Padding(0, 1)

	inactiveStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1)

	sectionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		MarginTop(1)

	filterWidth := int(float64(m.width) * 0.3)
	listWidth := m.width - filterWidth - 4

	// Build filters panel
	filtersPanel := m.renderFiltersPanel(filterWidth, sectionStyle)

	// Style filters panel based on active state
	if m.activePanel == PanelFilters {
		filtersPanel = activeStyle.Width(filterWidth).Render(filtersPanel)
	} else {
		filtersPanel = inactiveStyle.Width(filterWidth).Render(filtersPanel)
	}

	// Build list panel
	listPanel := m.list.View()

	// Style list panel based on active state
	if m.activePanel == PanelList {
		listPanel = activeStyle.Width(listWidth).Render(listPanel)
	} else {
		listPanel = inactiveStyle.Width(listWidth).Render(listPanel)
	}

	// Combine panels side by side
	mainView := lipgloss.JoinHorizontal(lipgloss.Top, filtersPanel, listPanel)

	// Help text
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Padding(1, 0)

	var help string
	if m.activePanel == PanelFilters {
		help = "↑/↓: navigate | space: toggle | tab: switch panel | q: quit"
	} else {
		help = "enter: select | ctrl+g: get | ctrl+s: bookmark | ctrl+t: tag | tab: filters | q: quit"
	}

	return fmt.Sprintf("%s\n%s", mainView, helpStyle.Render(help))
}

// renderFiltersPanel renders the left filters panel
func (m *FinderModel) renderFiltersPanel(width int, sectionStyle lipgloss.Style) string {
	var s strings.Builder

	s.WriteString("FILTERS\n\n")

	cursor := 0

	// Sources section
	s.WriteString(sectionStyle.Render("Sources"))
	s.WriteString("\n")
	for _, source := range m.availableSources {
		checkbox := "[ ]"
		if m.selectedSources[source] {
			checkbox = "[✓]"
		}

		line := fmt.Sprintf("%s %s", checkbox, source)
		if cursor == m.filterCursor && m.activePanel == PanelFilters {
			line = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render("→ " + line)
		} else {
			line = "  " + line
		}
		s.WriteString(line + "\n")
		cursor++
	}

	// Tags section
	if len(m.availableTags) > 0 {
		s.WriteString("\n")
		s.WriteString(sectionStyle.Render("Tags"))
		s.WriteString("\n")
		for _, tag := range m.availableTags {
			checkbox := "[ ]"
			if m.selectedTags[tag] {
				checkbox = "[✓]"
			}

			line := fmt.Sprintf("%s %s", checkbox, tag)
			if cursor == m.filterCursor && m.activePanel == PanelFilters {
				line = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render("→ " + line)
			} else {
				line = "  " + line
			}
			s.WriteString(line + "\n")
			cursor++
		}
	}

	// Bookmarked toggle
	s.WriteString("\n")
	s.WriteString(sectionStyle.Render("Other"))
	s.WriteString("\n")
	checkbox := "[ ]"
	if m.showBookmarked {
		checkbox = "[✓]"
	}
	line := fmt.Sprintf("%s Bookmarked only", checkbox)
	if cursor == m.filterCursor && m.activePanel == PanelFilters {
		line = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render("→ " + line)
	} else {
		line = "  " + line
	}
	s.WriteString(line + "\n")

	// Stats
	s.WriteString(fmt.Sprintf("\nShowing: %d/%d prompts", len(m.filteredPrompts), len(m.allPrompts)))

	return s.String()
}

// SelectedID returns the selected prompt ID
func (m FinderModel) SelectedID() string {
	return m.selectedID
}

// ShouldGet returns whether the get action was triggered
func (m FinderModel) ShouldGet() bool {
	return m.actionGet
}

// ShouldBookmark returns whether the bookmark action was triggered
func (m FinderModel) ShouldBookmark() bool {
	return m.actionBookmark
}

// ShouldTag returns whether the tag action was triggered
func (m FinderModel) ShouldTag() bool {
	return m.actionTag
}

// RunFinder runs the interactive finder and returns the selected prompt ID and action
func RunFinder(prompts []models.Prompt) (selectedID string, action string, err error) {
	p := tea.NewProgram(NewFinderModel(prompts), tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return "", "", fmt.Errorf("error running finder: %w", err)
	}

	m := finalModel.(FinderModel)
	selectedID = m.SelectedID()

	if m.ShouldGet() {
		action = "get"
	} else if m.ShouldBookmark() {
		action = "bookmark"
	} else if m.ShouldTag() {
		action = "tag"
	} else {
		action = "select"
	}

	return selectedID, action, nil
}
