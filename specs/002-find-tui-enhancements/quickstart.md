# TUI Enhancement Quickstart Guide

**Feature**: Enhance pkit Find TUI
**Branch**: 002-find-tui-enhancements
**Target**: Developers implementing the 7 TUI enhancements
**Date**: 2025-12-28

## Quick Start

### Prerequisites
- Go 1.25.4 installed
- pkit repository cloned
- Basic understanding of Bubbletea framework

### Build and Test
```bash
# Build binary
go build -o bin/pkit ./cmd/pkit

# Run TUI (requires at least one subscribed source)
./bin/pkit find

# Run with specific filter
./bin/pkit find --source fabric
./bin/pkit find --bookmarked
```

## Architecture Overview

### File Structure
```
internal/tui/
└── finder.go              # ~2000 lines, all TUI logic
    ├── Model struct       # State management
    ├── Init() → Model     # Initialization
    ├── Update(msg) → Model, Cmd  # Event handling
    └── View() → string    # Rendering
```

### Bubbletea Elm Architecture

**Three core functions**:

1. **Init()**: Initialize model and start commands
   ```go
   func (m Model) Init() tea.Cmd {
       // Return initial commands (e.g., fetch data)
   }
   ```

2. **Update()**: Handle all events, update state
   ```go
   func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
       switch msg := msg.(type) {
       case tea.KeyMsg:
           // Handle keyboard input
       case tea.WindowSizeMsg:
           // Handle terminal resize
       }
       return m, cmd
   }
   ```

3. **View()**: Render current state to string
   ```go
   func (m Model) View() string {
       // Return terminal output as string
   }
   ```

**Key concept**: Model is immutable. Update() returns *new* state, never modifies in-place.

## Key Components

### Bubbles Components (Reusable TUI Widgets)

**list.Model** - Interactive list with selection
```go
import "github.com/charmbracelet/bubbles/list"

// Initialize
l := list.New(items, delegate, width, height)
l.SetShowTitle(false)
l.SetFilteringEnabled(false)

// In Update()
l, cmd = l.Update(msg)

// In View()
return l.View()
```

**textinput.Model** - Text input field
```go
import "github.com/charmbracelet/bubbles/textinput"

// Initialize
ti := textinput.New()
ti.Placeholder = "Enter search term..."
ti.Focus()

// In Update()
ti, cmd = ti.Update(msg)

// In View()
return ti.View()
```

**viewport.Model** - Scrollable content area
```go
import "github.com/charmbracelet/bubbles/viewport"

// Initialize
vp := viewport.New(width, height)
vp.SetContent(longText)

// In Update()
vp, cmd = vp.Update(msg)

// In View()
return vp.View()
```

### Lipgloss (Styling and Layout)

**Styling**:
```go
import "github.com/charmbracelet/lipgloss"

// Create style
titleStyle := lipgloss.NewStyle().
    Bold(true).
    Foreground(lipgloss.Color("205")).
    Background(lipgloss.Color("235")).
    Padding(0, 1)

// Apply style
rendered := titleStyle.Render("My Title")
```

**Layout**:
```go
// Vertical layout (stack)
content := lipgloss.JoinVertical(
    lipgloss.Left,
    "Line 1",
    "Line 2",
    "Line 3",
)

// Horizontal layout (side-by-side)
content := lipgloss.JoinHorizontal(
    lipgloss.Top,
    leftPanel,
    rightPanel,
)
```

**Width measurement** (ANSI-aware):
```go
// CORRECT: Visual width ignoring ANSI codes
visualWidth := lipgloss.Width(styledText)

// WRONG: Byte length including ANSI codes
wrongWidth := len(styledText)  // Don't use this for display!
```

### Overlay Library (Dialog System)

```go
import "github.com/rmhubbert/bubbletea-overlay"

// Render dialog over base view
func (m Model) View() string {
    baseView := m.renderBaseView()

    if m.showDialog {
        dialog := m.renderDialog()
        return overlay.Composite(
            dialog,           // Foreground (dialog)
            baseView,         // Background (base UI)
            overlay.Center,   // Horizontal position
            overlay.Center,   // Vertical position
            0,                // X offset
            0,                // Y offset
        )
    }

    return baseView
}
```

**Position options**: `overlay.Center`, `overlay.Top`, `overlay.Bottom`, `overlay.Left`, `overlay.Right`

## Implementation Patterns

### Pattern 1: Adding State

```go
// In Model struct (line ~50-120)
type Model struct {
    // Existing state
    prompts []index.SearchResult

    // NEW: Add your state here
    searchMode  bool
    searchQuery string
    searchInput textinput.Model
}
```

### Pattern 2: Handling Keyboard Input

```go
// In Update() method (line ~500-900)
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        // Add your key handler
        switch {
        case key.Matches(msg, m.keys.Search): // "/" key
            m.searchMode = true
            m.searchInput.Focus()
            return m, nil
        }
    }
}
```

### Pattern 3: Conditional Rendering

```go
// In View() method (line ~1200-1500)
func (m Model) View() string {
    // Render different UI based on mode
    if m.searchMode {
        return m.renderSearchView()
    }

    if m.inputMode == ModeViewingPrompt {
        return m.renderPreviewView()
    }

    return m.renderNormalView()
}
```

### Pattern 4: Real-time Filtering

```go
// In Update(), when search text changes
m.searchQuery = m.searchInput.Value()

// Filter prompts based on query
filtered := []index.SearchResult{}
queryLower := strings.ToLower(m.searchQuery)

for _, p := range m.preSearchList {
    if strings.Contains(strings.ToLower(p.Prompt.Name), queryLower) ||
       strings.Contains(strings.ToLower(p.Prompt.Description), queryLower) {
        filtered = append(filtered, p)
    }
}

m.filteredPrompts = filtered

// Update list component with new items
items := convertToListItems(filtered)
m.list.SetItems(items)
```

### Pattern 5: Showing Dialogs

```go
// In View()
func (m Model) View() string {
    baseView := m.renderBaseView()

    // Show dialog overlay if in dialog mode
    if m.inputMode == ModeAddingTag {
        dialog := m.renderTagDialog()
        return overlay.Composite(dialog, baseView, overlay.Center, overlay.Center, 0, 0)
    }

    return baseView
}

// Dialog rendering
func (m Model) renderTagDialog() string {
    // Get existing tags
    existingTags := m.userTags[m.currentPrompt.ID]
    existingText := "Current tags: " + strings.Join(existingTags, ", ")

    content := lipgloss.JoinVertical(
        lipgloss.Left,
        lipgloss.NewStyle().Bold(true).Render("Manage Tags"),
        "",
        lipgloss.NewStyle().Faint(true).Render(existingText),
        "",
        "Add tags (comma-separated):",
        m.input.View(),
        "",
        "Enter: save | Esc: cancel",
    )

    // Add border and padding
    dialog := lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        Padding(1, 2).
        Width(50).
        Render(content)

    return dialog
}
```

### Pattern 6: Status Messages

```go
// In Model struct
type Model struct {
    statusMsg    string
    statusExpiry time.Time
}

// Set status with auto-clear
func (m *Model) setStatus(msg string, duration time.Duration) {
    m.statusMsg = msg
    m.statusExpiry = time.Now().Add(duration)
}

// In Update(), clear expired status
if !m.statusExpiry.IsZero() && time.Now().After(m.statusExpiry) {
    m.statusMsg = ""
    m.statusExpiry = time.Time{}
}

// In View(), render status if active
func (m Model) renderStatusBar() string {
    if m.statusMsg != "" {
        return lipgloss.NewStyle().
            Foreground(lipgloss.Color("42")).
            Render(m.statusMsg)
    }
    return m.generateHelpText()
}
```

### Pattern 7: Dynamic Text with ANSI Codes

```go
// CORRECT: Measure visual width
text := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render("Hello")
visualWidth := lipgloss.Width(text)  // Returns 5 (ignores ANSI codes)

// Truncate safely (character-aware)
func truncateText(text string, maxWidth int) string {
    if lipgloss.Width(text) <= maxWidth {
        return text
    }

    runes := []rune(text)
    truncated := ""
    currentWidth := 0

    for _, r := range runes {
        charWidth := runewidth.RuneWidth(r)
        if currentWidth + charWidth > (maxWidth - 3) {
            break
        }
        truncated += string(r)
        currentWidth += charWidth
    }

    return truncated + "..."
}
```

## Feature Implementation Guide

### Feature 1: Search Within Filtered Results

**Files to modify**: `internal/tui/finder.go`

**Steps**:
1. Add state to Model:
   ```go
   searchMode       bool
   searchQuery      string
   searchInput      textinput.Model
   preSearchList    []index.SearchResult
   ```

2. Initialize search input in model creation:
   ```go
   si := textinput.New()
   si.Placeholder = "Search prompts..."
   si.Width = 40
   m.searchInput = si
   ```

3. Handle "/" key in Update():
   ```go
   case key.Matches(msg, m.keys.Search):
       m.searchMode = true
       m.searchInput.Focus()
       m.preSearchList = m.filteredPrompts
       return m, nil
   ```

4. Handle search input updates:
   ```go
   if m.searchMode {
       m.searchInput, cmd = m.searchInput.Update(msg)
       m.searchQuery = m.searchInput.Value()
       m.filteredPrompts = m.applySearchFilter(m.preSearchList, m.searchQuery)
       m.list.SetItems(convertToListItems(m.filteredPrompts))
       return m, cmd
   }
   ```

5. Render search UI in View():
   ```go
   if m.searchMode {
       searchBar := m.renderSearchBar()
       promptsList := m.list.View()
       return lipgloss.JoinVertical(lipgloss.Left, searchBar, promptsList)
   }
   ```

**Testing**:
- Filter by source → press "/" → type "code" → verify only matching prompts shown
- Press Esc → verify original filtered list restored
- Press Enter → verify search results persist

### Feature 2: Tag Display Truncation

**Files to modify**: `internal/tui/finder.go`

**Steps**:
1. Add truncation cache to Model:
   ```go
   truncatedTags map[string]string
   ```

2. Implement truncation function:
   ```go
   const MaxTagDisplayLength = 25

   func truncateTag(tag string) string {
       visualWidth := lipgloss.Width(tag)
       if visualWidth <= MaxTagDisplayLength {
           return tag
       }
       // Truncate logic using runewidth
       // ... (see Pattern 7 above)
   }
   ```

3. Cache truncated tags on load:
   ```go
   func (m *Model) loadTags() {
       tags := storage.GetAllTags()
       m.truncatedTags = make(map[string]string)
       for _, tag := range tags {
           m.truncatedTags[tag] = truncateTag(tag)
       }
   }
   ```

4. Use truncated version in filter panel rendering:
   ```go
   for _, tag := range tags {
       displayTag := m.truncatedTags[tag]
       checkbox := "[ ]"
       if m.selectedTags[tag] {
           checkbox = "[x]"
       }
       line := fmt.Sprintf("%s %s", checkbox, displayTag)
       filterLines = append(filterLines, line)
   }
   ```

**Testing**:
- Create tag with 30 characters → verify truncated to 25 with "..."
- Create tag with 25 characters → verify shown in full
- Verify checkbox alignment is consistent for all tag lengths

### Feature 3: Numeric Pagination

**Files to modify**: `internal/tui/finder.go`

**Steps**:
1. Disable list's built-in pagination:
   ```go
   l.Paginator.Type = paginator.Arabic  // Use numeric
   l.SetShowHelp(false)  // Hide built-in help
   ```

2. Add pagination calculation:
   ```go
   func (m *Model) updatePagination() {
       totalItems := len(m.filteredPrompts)
       pageSize := m.list.Paginator.PerPage
       m.totalPages = (totalItems + pageSize - 1) / pageSize
       m.currentPage = m.list.Paginator.Page + 1
   }
   ```

3. Render pagination in panel border:
   ```go
   paginationText := fmt.Sprintf("%d/%d", m.currentPage, m.totalPages)
   // Render in bottom-right of prompts panel border
   ```

**Testing**:
- Navigate to page 3 of 5 → verify "3/5" displayed
- Filter to single page → verify "1/1" displayed
- Press → → verify pagination updates to "4/5"

### Feature 4: Preview Dialog Height

**Files to modify**: `internal/tui/finder.go`

**Steps**:
1. Implement dynamic height calculation:
   ```go
   func (m *Model) calculatePreviewHeight(contentLines int) int {
       const (
           MinHeight = 15
           MaxHeightPercent = 0.5
           DialogOverhead = 6
       )
       maxAllowed := int(float64(m.height) * MaxHeightPercent)
       desired := contentLines + DialogOverhead
       // Apply min/max constraints
       // ... (see data-model.md)
   }
   ```

2. Use when creating preview viewport:
   ```go
   contentLines := strings.Count(m.currentPrompt.Content, "\n")
   height := m.calculatePreviewHeight(contentLines)
   m.previewViewport = viewport.New(width, height)
   ```

**Testing**:
- Preview 10-line prompt → verify small dialog
- Preview 150-line prompt → verify max 50% height with scrolling
- Resize terminal → verify dialog recalculates height

### Feature 5: Bookmark Management in Preview

**Files to modify**: `internal/tui/finder.go`

**Steps**:
1. Add keyboard handlers in preview mode:
   ```go
   if m.inputMode == ModeViewingPrompt {
       case key.Matches(msg, m.keys.ToggleBookmark):  // ctrl+b
           return m, m.toggleBookmarkInPreview()
       case key.Matches(msg, m.keys.RemoveBookmark):  // ctrl+x
           return m, m.removeBookmarkInPreview()
   }
   ```

2. Implement bookmark toggle:
   ```go
   func (m *Model) toggleBookmarkInPreview() tea.Cmd {
       promptID := m.currentPrompt.ID
       isBookmarked := m.bookmarks[promptID]

       if isBookmarked {
           storage.RemoveBookmark(promptID)
           delete(m.bookmarks, promptID)
           m.setStatus("✓ Bookmark removed", 2*time.Second)
       } else {
           storage.AddBookmark(m.currentPrompt)
           m.bookmarks[promptID] = true
           m.setStatus("✓ Bookmarked", 2*time.Second)
       }

       return nil  // Stay in preview
   }
   ```

3. Update help text in preview mode:
   ```go
   "↑/↓: scroll | ctrl+b: bookmark | ctrl+t: tags | Esc: close"
   ```

**Testing**:
- Preview unbookmarked prompt → ctrl+b → verify "✓ Bookmarked" status
- Preview bookmarked prompt → ctrl+x → verify "✓ Bookmark removed" status
- Close preview → verify bookmark icon updated in list

### Feature 6: Tag Dialog with Existing Tags

**Files to modify**: `internal/tui/finder.go`

**Steps**:
1. Modify tag dialog rendering:
   ```go
   func (m Model) renderTagDialog() string {
       existingTags := m.userTags[m.currentPrompt.ID]

       var existingText string
       if len(existingTags) == 0 {
           existingText = "No tags assigned"
       } else {
           existingText = "Current tags: " + strings.Join(existingTags, ", ")
       }

       content := lipgloss.JoinVertical(
           lipgloss.Left,
           lipgloss.NewStyle().Bold(true).Render("Manage Tags"),
           "",
           lipgloss.NewStyle().Faint(true).Render(existingText),
           "",
           "Add tags (comma-separated):",
           m.input.View(),
       )
       // ... render with overlay
   }
   ```

2. Handle tag addition (preserve existing):
   ```go
   newTagsStr := m.input.Value()
   newTags := strings.Split(newTagsStr, ",")
   existingTags := m.userTags[m.currentPrompt.ID]

   // Merge and deduplicate
   allTags := append(existingTags, newTags...)
   uniqueTags := removeDuplicates(trimAll(allTags))

   storage.SetTags(m.currentPrompt.ID, uniqueTags)
   ```

**Testing**:
- Open tag dialog on untagged prompt → verify "No tags assigned"
- Open tag dialog on tagged prompt → verify "Current tags: code, review"
- Add new tags → verify merged with existing (no duplicates)

### Feature 7: Navigation Help Text

**Files to modify**: `internal/tui/finder.go`

**Steps**:
1. Implement dynamic help generation:
   ```go
   func (m *Model) generateHelpText() string {
       switch {
       case m.inputMode == ModeViewingPrompt:
           return "↑/↓: scroll | ctrl+b: bookmark | ctrl+t: tags | Esc: close"
       case m.searchMode:
           return "Type to search | Enter: apply | Esc: cancel"
       case m.activePanel == PanelFilters:
           return "↑/↓: navigate | Space: toggle | Tab: switch | /: search | q: quit"
       default:
           hasPages := m.totalPages > 1
           if hasPages {
               return "↑/↓: navigate | ←/→: pages | /: search | Enter: preview | ctrl+t: tags | ctrl+b: bookmark | q: quit"
           }
           return "↑/↓: navigate | /: search | Enter: preview | ctrl+t: tags | ctrl+b: bookmark | q: quit"
       }
   }
   ```

2. Render in status bar area:
   ```go
   func (m Model) View() string {
       // ... main content
       helpText := m.generateHelpText()
       helpBar := lipgloss.NewStyle().Faint(true).Render(helpText)
       return lipgloss.JoinVertical(lipgloss.Left, mainContent, helpBar)
   }
   ```

**Testing**:
- Normal mode with multiple pages → verify page navigation shown
- Normal mode with single page → verify page navigation hidden
- Search mode → verify search-specific help
- Preview mode → verify preview-specific help with bookmark/tag shortcuts

## Debugging Tips

### Debug Output (stderr)
```go
// Print debug info without corrupting TUI
fmt.Fprintf(os.Stderr, "DEBUG: searchMode=%v, query=%s\n", m.searchMode, m.searchQuery)
```

### ANSI Width Issues
```go
// Check visual width vs byte length
text := styledText
fmt.Fprintf(os.Stderr, "Visual: %d, Bytes: %d\n", lipgloss.Width(text), len(text))

// If they differ significantly, ANSI codes are present
```

### State Transitions
```go
// Log mode changes
fmt.Fprintf(os.Stderr, "Mode transition: %v → %v\n", oldMode, newMode)
```

### Terminal Resize
```go
// In Update()
case tea.WindowSizeMsg:
    fmt.Fprintf(os.Stderr, "Terminal size: %dx%d\n", msg.Width, msg.Height)
```

## Common Pitfalls

### ❌ Modifying Model In-Place
```go
// WRONG
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    m.searchMode = true  // This creates a copy!
    return m, nil
}

// CORRECT (explicit)
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    newModel := m
    newModel.searchMode = true
    return newModel, nil
}

// BEST (idiomatic)
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    m.searchMode = true  // This works because we return m
    return m, nil
}
```

### ❌ Using len() for Display Width
```go
// WRONG
text := lipgloss.NewStyle().Bold(true).Render("Hello")
width := len(text)  // Returns ~15 (includes ANSI codes)

// CORRECT
width := lipgloss.Width(text)  // Returns 5 (visual width)
```

### ❌ Forgetting to Update Filtered List
```go
// WRONG
m.filteredPrompts = newResults
// List still shows old items!

// CORRECT
m.filteredPrompts = newResults
items := convertToListItems(newResults)
m.list.SetItems(items)  // Update list component
```

### ❌ Blocking Operations in Update()
```go
// WRONG (blocks TUI)
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    time.Sleep(2 * time.Second)  // Never block!
    return m, nil
}

// CORRECT (use commands for async work)
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    return m, m.fetchDataAsync()
}
```

## Resources

### Official Documentation
- Bubbletea: https://github.com/charmbracelet/bubbletea
- Bubbles: https://github.com/charmbracelet/bubbles
- Lipgloss: https://github.com/charmbracelet/lipgloss
- Overlay: https://github.com/rmhubbert/bubbletea-overlay

### Project Documentation
- Feature Spec: `specs/002-find-tui-enhancements/spec.md`
- Implementation Plan: `specs/002-find-tui-enhancements/plan.md`
- Research: `specs/002-find-tui-enhancements/research.md`
- Data Model: `specs/002-find-tui-enhancements/data-model.md`
- Keyboard Shortcuts: `specs/002-find-tui-enhancements/contracts/keyboard-shortcuts.md`
- Display Rules: `specs/002-find-tui-enhancements/contracts/display-rules.md`

### Code Reference
- Main TUI: `internal/tui/finder.go`
- Storage: `internal/storage/bookmarks.go`, `internal/storage/tags.go`
- Index: `internal/index/search.go`
- Models: `pkg/models/prompt.go`

## Summary

**Key takeaways**:
1. All changes in `internal/tui/finder.go`
2. Use Bubbletea Elm architecture (immutable updates)
3. Use lipgloss.Width() for ANSI-aware measurements
4. Test interactively after each change
5. Debug with stderr (doesn't corrupt TUI)
6. Follow existing patterns (modes, filters, overlays)

**Next steps**:
1. Read feature spec and contracts
2. Choose a feature to implement
3. Add state to Model struct
4. Implement Update() handler
5. Implement View() rendering
6. Test manually
7. Repeat for next feature

Good luck! The TUI enhancements are well-scoped and follow established patterns.
