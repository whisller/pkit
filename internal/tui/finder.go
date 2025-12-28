package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
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

// InputMode represents the current input mode
type InputMode int

const (
	ModeNormal InputMode = iota
	ModeAddingTag
	ModeAddingAlias
	ModeAddingNotes
	ModeRemovingTag
)

// KeyMap defines keyboard shortcuts
type KeyMap struct {
	Quit         key.Binding
	Select       key.Binding
	Get          key.Binding
	Bookmark     key.Binding
	Tag          key.Binding
	Alias        key.Binding
	RemoveTag    key.Binding
	Notes        key.Binding
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
			key.WithHelp("ctrl+s", "toggle bookmark"),
		),
		Tag: key.NewBinding(
			key.WithKeys("ctrl+t"),
			key.WithHelp("ctrl+t", "add tags"),
		),
		Alias: key.NewBinding(
			key.WithKeys("ctrl+a"),
			key.WithHelp("ctrl+a", "add alias"),
		),
		RemoveTag: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("ctrl+r", "remove tags"),
		),
		Notes: key.NewBinding(
			key.WithKeys("ctrl+n"),
			key.WithHelp("ctrl+n", "add notes"),
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

	// Input mode
	inputMode       InputMode
	textInput       textinput.Model
	statusMessage   string
	statusTimeout   time.Time
	currentPromptID string // ID of prompt being operated on
	promptTags      []string // Tags for current prompt (for removal)
	tagRemoveCursor int

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

	// Create text input for dialogs
	ti := textinput.New()
	ti.Placeholder = "Enter value..."
	ti.CharLimit = 200
	ti.Width = 50

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
		inputMode:        ModeNormal,
		textInput:        ti,
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

// reloadData reloads bookmarks and tags from disk
func (m *FinderModel) reloadData() {
	// Reload bookmarks
	bookmarkedIDs := make(map[string]bool)
	bookmarks, _ := bookmark.LoadBookmarks()
	for _, bm := range bookmarks {
		bookmarkedIDs[bm.PromptID] = true
	}
	m.bookmarkedIDs = bookmarkedIDs

	// Reload tags
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
	m.availableTags = tags

	// Reapply filters
	m.applyFilters()
}

// setStatus sets a status message with timeout
func (m *FinderModel) setStatus(msg string, duration time.Duration) {
	m.statusMessage = msg
	m.statusTimeout = time.Now().Add(duration)
}

// Init initializes the model
func (m FinderModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m FinderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Clear status message if timeout expired
	if !m.statusTimeout.IsZero() && time.Now().After(m.statusTimeout) {
		m.statusMessage = ""
		m.statusTimeout = time.Time{}
	}

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
		// Handle input modes
		if m.inputMode != ModeNormal {
			return m.handleInputMode(msg)
		}

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
	if m.activePanel == PanelList && m.inputMode == ModeNormal {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	return m, nil
}

// handleInputMode handles input when in an input mode
func (m FinderModel) handleInputMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+c":
		// Cancel input
		m.inputMode = ModeNormal
		m.textInput.SetValue("")
		return m, nil

	case "enter":
		// Process input based on mode
		return m.processInput()
	}

	// Handle tag removal mode differently (list navigation)
	if m.inputMode == ModeRemovingTag {
		switch msg.String() {
		case "up", "k":
			m.tagRemoveCursor--
			if m.tagRemoveCursor < 0 {
				m.tagRemoveCursor = len(m.promptTags) - 1
			}
			return m, nil
		case "down", "j":
			m.tagRemoveCursor++
			if m.tagRemoveCursor >= len(m.promptTags) {
				m.tagRemoveCursor = 0
			}
			return m, nil
		case " ":
			// Remove selected tag
			if m.tagRemoveCursor < len(m.promptTags) {
				return m.removeTag(m.tagRemoveCursor)
			}
			return m, nil
		}
		return m, nil
	}

	// Update text input
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// processInput processes the input based on current mode
func (m FinderModel) processInput() (tea.Model, tea.Cmd) {
	value := strings.TrimSpace(m.textInput.Value())
	if value == "" {
		m.inputMode = ModeNormal
		m.textInput.SetValue("")
		return m, nil
	}

	switch m.inputMode {
	case ModeAddingTag:
		return m.addTags(value)
	case ModeAddingAlias:
		return m.addAlias(value)
	case ModeAddingNotes:
		return m.addNotes(value)
	}

	m.inputMode = ModeNormal
	m.textInput.SetValue("")
	return m, nil
}

// addTags adds tags to the current prompt
func (m FinderModel) addTags(tagsStr string) (tea.Model, tea.Cmd) {
	// Parse tags (comma-separated)
	tags := strings.Split(tagsStr, ",")
	parsedTags := make([]string, 0, len(tags))
	for _, t := range tags {
		t = strings.TrimSpace(t)
		if t != "" {
			parsedTags = append(parsedTags, strings.ToLower(t))
		}
	}

	if len(parsedTags) == 0 {
		m.setStatus("No tags entered", 2*time.Second)
		m.inputMode = ModeNormal
		m.textInput.SetValue("")
		return m, nil
	}

	// Add tags
	tagMgr := tag.NewManager()
	if err := tagMgr.AddTags(m.currentPromptID, parsedTags); err != nil {
		m.setStatus(fmt.Sprintf("Error: %v", err), 3*time.Second)
	} else {
		m.setStatus(fmt.Sprintf("✓ Added tags: %s", strings.Join(parsedTags, ", ")), 3*time.Second)
		m.reloadData()
	}

	m.inputMode = ModeNormal
	m.textInput.SetValue("")
	return m, nil
}

// addAlias adds an alias to the current prompt
func (m FinderModel) addAlias(aliasName string) (tea.Model, tea.Cmd) {
	// Validate alias name
	aliasName = strings.ToLower(strings.TrimSpace(aliasName))
	if aliasName == "" {
		m.setStatus("Alias cannot be empty", 2*time.Second)
		m.inputMode = ModeNormal
		m.textInput.SetValue("")
		return m, nil
	}

	// Add bookmark with alias (alias is stored in bookmarks)
	mgr := bookmark.NewManager()
	bm := models.Bookmark{
		PromptID: m.currentPromptID,
		Notes:    fmt.Sprintf("Alias: %s", aliasName),
	}

	if err := mgr.AddBookmark(bm); err != nil {
		m.setStatus(fmt.Sprintf("Error: %v", err), 3*time.Second)
	} else {
		m.setStatus(fmt.Sprintf("✓ Added alias: %s", aliasName), 3*time.Second)
		m.reloadData()
	}

	m.inputMode = ModeNormal
	m.textInput.SetValue("")
	return m, nil
}

// addNotes adds notes to the bookmark
func (m FinderModel) addNotes(notes string) (tea.Model, tea.Cmd) {
	mgr := bookmark.NewManager()

	// Check if already bookmarked
	if !m.bookmarkedIDs[m.currentPromptID] {
		// Create bookmark with notes
		bm := models.Bookmark{
			PromptID: m.currentPromptID,
			Notes:    notes,
		}
		if err := mgr.AddBookmark(bm); err != nil {
			m.setStatus(fmt.Sprintf("Error: %v", err), 3*time.Second)
		} else {
			m.setStatus("✓ Bookmarked with notes", 3*time.Second)
			m.reloadData()
		}
	} else {
		// Update existing bookmark notes
		if err := mgr.UpdateBookmark(m.currentPromptID, func(bm *models.Bookmark) error {
			bm.Notes = notes
			return nil
		}); err != nil {
			m.setStatus(fmt.Sprintf("Error: %v", err), 3*time.Second)
		} else {
			m.setStatus("✓ Updated notes", 3*time.Second)
		}
	}

	m.inputMode = ModeNormal
	m.textInput.SetValue("")
	return m, nil
}

// removeTag removes a tag at the given index
func (m FinderModel) removeTag(idx int) (tea.Model, tea.Cmd) {
	if idx < 0 || idx >= len(m.promptTags) {
		return m, nil
	}

	tagToRemove := m.promptTags[idx]
	tagMgr := tag.NewManager()

	if err := tagMgr.RemoveTags(m.currentPromptID, []string{tagToRemove}); err != nil {
		m.setStatus(fmt.Sprintf("Error: %v", err), 3*time.Second)
	} else {
		m.setStatus(fmt.Sprintf("✓ Removed tag: %s", tagToRemove), 3*time.Second)
		m.reloadData()

		// Reload tags for this prompt
		remainingTags, _ := tagMgr.GetTags(m.currentPromptID)
		m.promptTags = remainingTags

		// Exit if no more tags
		if len(m.promptTags) == 0 {
			m.inputMode = ModeNormal
			m.setStatus("✓ All tags removed", 2*time.Second)
			return m, nil
		}

		// Adjust cursor
		if m.tagRemoveCursor >= len(m.promptTags) {
			m.tagRemoveCursor = len(m.promptTags) - 1
		}
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

	// Get selected prompt for operations
	var selectedPrompt *PromptItem
	if item, ok := m.list.SelectedItem().(PromptItem); ok {
		selectedPrompt = &item
	}

	switch {
	case key.Matches(msg, m.keys.Select):
		if selectedPrompt != nil {
			m.selectedID = selectedPrompt.Prompt.ID
			m.quitting = true
			return m, tea.Quit
		}

	case key.Matches(msg, m.keys.Get):
		if selectedPrompt != nil {
			m.selectedID = selectedPrompt.Prompt.ID
			m.actionGet = true
			m.quitting = true
			return m, tea.Quit
		}

	case key.Matches(msg, m.keys.Bookmark):
		// Toggle bookmark
		if selectedPrompt != nil {
			return m.toggleBookmark(selectedPrompt.Prompt.ID)
		}

	case key.Matches(msg, m.keys.Tag):
		// Add tags
		if selectedPrompt != nil {
			m.currentPromptID = selectedPrompt.Prompt.ID
			m.inputMode = ModeAddingTag
			m.textInput.Placeholder = "Enter tags (comma-separated)..."
			m.textInput.SetValue("")
			m.textInput.Focus()
			return m, nil
		}

	case key.Matches(msg, m.keys.Alias):
		// Add alias
		if selectedPrompt != nil {
			m.currentPromptID = selectedPrompt.Prompt.ID
			m.inputMode = ModeAddingAlias
			m.textInput.Placeholder = "Enter alias name..."
			m.textInput.SetValue("")
			m.textInput.Focus()
			return m, nil
		}

	case key.Matches(msg, m.keys.RemoveTag):
		// Remove tags
		if selectedPrompt != nil {
			tagMgr := tag.NewManager()
			tags, err := tagMgr.GetTags(selectedPrompt.Prompt.ID)
			if err != nil || len(tags) == 0 {
				m.setStatus("No tags to remove", 2*time.Second)
				return m, nil
			}
			m.currentPromptID = selectedPrompt.Prompt.ID
			m.promptTags = tags
			m.tagRemoveCursor = 0
			m.inputMode = ModeRemovingTag
			return m, nil
		}

	case key.Matches(msg, m.keys.Notes):
		// Add notes
		if selectedPrompt != nil {
			m.currentPromptID = selectedPrompt.Prompt.ID
			m.inputMode = ModeAddingNotes
			m.textInput.Placeholder = "Enter notes..."
			m.textInput.SetValue("")
			m.textInput.Focus()
			return m, nil
		}
	}

	// Update list
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// toggleBookmark toggles bookmark status for a prompt
func (m FinderModel) toggleBookmark(promptID string) (tea.Model, tea.Cmd) {
	mgr := bookmark.NewManager()

	if m.bookmarkedIDs[promptID] {
		// Remove bookmark
		if err := mgr.RemoveBookmark(promptID); err != nil {
			m.setStatus(fmt.Sprintf("Error: %v", err), 3*time.Second)
		} else {
			m.setStatus("✓ Bookmark removed", 2*time.Second)
			m.reloadData()
		}
	} else {
		// Add bookmark
		bm := models.Bookmark{
			PromptID: promptID,
			Notes:    "Bookmarked via finder",
		}
		if err := mgr.AddBookmark(bm); err != nil {
			m.setStatus(fmt.Sprintf("Error: %v", err), 3*time.Second)
		} else {
			m.setStatus("✓ Bookmarked", 2*time.Second)
			m.reloadData()
		}
	}

	return m, nil
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

	// Handle input modes
	if m.inputMode != ModeNormal {
		return m.renderInputDialog()
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

	// Status message
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("42")).
		Bold(true)

	var statusView string
	if m.statusMessage != "" {
		statusView = statusStyle.Render(m.statusMessage) + "\n"
	}

	// Help text
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Padding(1, 0)

	var help string
	if m.activePanel == PanelFilters {
		help = "↑/↓: navigate | space: toggle | tab: switch panel | q: quit"
	} else {
		help = "enter: select | ctrl+g: get | ctrl+s: bookmark | ctrl+t: tags | ctrl+a: alias | ctrl+r: remove tags | tab: filters"
	}

	return fmt.Sprintf("%s\n%s%s", mainView, statusView, helpStyle.Render(help))
}

// renderInputDialog renders the input dialog
func (m *FinderModel) renderInputDialog() string {
	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		Padding(1, 2).
		Width(60)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205"))

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	var title, help string
	var content strings.Builder

	switch m.inputMode {
	case ModeAddingTag:
		title = "Add Tags"
		help = "Enter comma-separated tags (e.g., dev,security). Press Enter to save, Esc to cancel."
		content.WriteString(fmt.Sprintf("Prompt: %s\n\n", m.currentPromptID))
		content.WriteString(m.textInput.View())

	case ModeAddingAlias:
		title = "Add Alias"
		help = "Enter an alias name for this prompt. Press Enter to save, Esc to cancel."
		content.WriteString(fmt.Sprintf("Prompt: %s\n\n", m.currentPromptID))
		content.WriteString(m.textInput.View())

	case ModeAddingNotes:
		title = "Add Notes"
		help = "Enter notes for this bookmark. Press Enter to save, Esc to cancel."
		content.WriteString(fmt.Sprintf("Prompt: %s\n\n", m.currentPromptID))
		content.WriteString(m.textInput.View())

	case ModeRemovingTag:
		title = "Remove Tags"
		help = "↑/↓: navigate | Space: remove selected tag | Esc: cancel"
		content.WriteString(fmt.Sprintf("Prompt: %s\n\n", m.currentPromptID))
		content.WriteString("Select tag to remove:\n\n")
		for i, tag := range m.promptTags {
			cursor := "  "
			if i == m.tagRemoveCursor {
				cursor = "→ "
			}
			content.WriteString(fmt.Sprintf("%s%s\n", cursor, tag))
		}
	}

	dialog := fmt.Sprintf("%s\n\n%s\n\n%s",
		titleStyle.Render(title),
		content.String(),
		helpStyle.Render(help))

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		dialogStyle.Render(dialog),
	)
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
