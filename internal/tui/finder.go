package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/whisller/pkit/pkg/models"
)

// KeyMap defines keyboard shortcuts
type KeyMap struct {
	Quit     key.Binding
	Select   key.Binding
	Get      key.Binding
	Bookmark key.Binding
	Tag      key.Binding
	Up       key.Binding
	Down     key.Binding
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
	list           list.Model
	search         textinput.Model
	keys           KeyMap
	prompts        []models.Prompt
	width          int
	height         int
	selectedID     string
	actionGet      bool
	actionBookmark bool
	actionTag      bool
	quitting       bool
}

// NewFinderModel creates a new finder model
func NewFinderModel(prompts []models.Prompt) FinderModel {
	// Convert prompts to list items
	items := make([]list.Item, len(prompts))
	for i, p := range prompts {
		items[i] = PromptItem{Prompt: p}
	}

	// Create list
	delegate := list.NewDefaultDelegate()
	l := list.New(items, delegate, 0, 0)
	l.Title = "Find Prompt"
	l.SetShowHelp(true)
	l.SetFilteringEnabled(true)
	l.SetShowStatusBar(true)

	// Create search input
	ti := textinput.New()
	ti.Placeholder = "Search prompts..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 50

	return FinderModel{
		list:    l,
		search:  ti,
		keys:    DefaultKeyMap(),
		prompts: prompts,
	}
}

// Init initializes the model
func (m FinderModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages
func (m FinderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 4)
		return m, nil

	case tea.KeyMsg:
		// Don't match any key bindings while filtering
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit

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
	}

	// Update list
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the UI
func (m FinderModel) View() string {
	if m.quitting {
		return ""
	}

	// Styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Padding(0, 1)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Padding(1, 0)

	// Build view
	var s strings.Builder

	s.WriteString(titleStyle.Render("Interactive Prompt Finder"))
	s.WriteString("\n\n")
	s.WriteString(m.list.View())
	s.WriteString("\n")
	s.WriteString(helpStyle.Render(
		"enter: select | ctrl+g: get | ctrl+s: bookmark | ctrl+t: tag | q: quit",
	))

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
	p := tea.NewProgram(NewFinderModel(prompts))

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
