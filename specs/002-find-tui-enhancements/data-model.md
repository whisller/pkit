# Data Model: TUI State Extensions

**Feature**: Enhance pkit Find TUI
**Branch**: 002-find-tui-enhancements
**Date**: 2025-12-28

## Overview

This document defines the state model extensions needed for the 7 TUI enhancements. All changes are localized to `internal/tui/finder.go`. No database or persistent storage changes required - all state is ephemeral UI state.

## Current Model Structure

The existing `Model` struct in `internal/tui/finder.go` (approximately lines 50-120):

```go
type Model struct {
    // Data
    prompts          []index.SearchResult  // All prompts from index
    filteredPrompts  []index.SearchResult  // After applying filters
    sources          []string              // Available sources
    userTags         map[string][]string   // promptID -> tags
    bookmarks        map[string]bool       // promptID -> bookmarked

    // UI Components
    list             list.Model            // Bubbles list component
    filterList       list.Model            // Filter panel list
    width, height    int                   // Terminal dimensions

    // State
    activePanel      Panel                 // PanelFilters or PanelList
    selectedSource   string                // Current source filter
    selectedTags     map[string]bool       // tag -> selected
    bookmarkedOnly   bool                  // Bookmark filter active

    // Input modes
    inputMode        InputMode             // ModeNormal, ModeAddingTag, etc.
    input            textinput.Model       // For various input dialogs
    currentPrompt    *models.Prompt        // Currently selected/previewed

    // Status
    statusMsg        string
    statusExpiry     time.Time
}
```

## New State Extensions

### 1. Search Functionality (FR-001 to FR-005)

```go
// Add to Model struct:

// Search state
searchMode       bool              // Is search UI active?
searchQuery      string            // Current search text
searchInput      textinput.Model   // Dedicated search input component
searchedPrompts  []index.SearchResult // Search results cache

// Helper state
preSearchList    []index.SearchResult // Backup of filteredPrompts before search
```

**State Transitions**:
- `Normal → SearchMode`: User presses "/" key
  - Set `searchMode = true`
  - Focus `searchInput`
  - Store `preSearchList = filteredPrompts`
- `SearchMode → Normal` (cancel): User presses Esc
  - Set `searchMode = false`
  - Restore `filteredPrompts = preSearchList`
  - Clear `searchQuery`
- `SearchMode → Normal` (apply): User presses Enter
  - Set `searchMode = false`
  - Keep `filteredPrompts` as current search results
- `SearchMode + typing`: On each keystroke
  - Update `searchQuery`
  - Filter `preSearchList` by search query → `filteredPrompts`
  - Update list component with new filtered results

**Search Logic**:
```go
func (m *Model) applySearchFilter(prompts []index.SearchResult, query string) []index.SearchResult {
    if query == "" {
        return prompts
    }

    queryLower := strings.ToLower(query)
    var results []index.SearchResult

    for _, p := range prompts {
        // Search in ID, name, description
        if strings.Contains(strings.ToLower(p.Prompt.ID), queryLower) ||
           strings.Contains(strings.ToLower(p.Prompt.Name), queryLower) ||
           strings.Contains(strings.ToLower(p.Prompt.Description), queryLower) {
            results = append(results, p)
        }
    }

    return results
}
```

### 2. Navigation Help Text (FR-006 to FR-009)

```go
// Add to Model struct:

// Help text state
helpText    string    // Current help text (dynamically generated)
```

**Dynamic Help Generation**:
```go
func (m *Model) generateHelpText() string {
    switch {
    case m.inputMode == ModeViewingPrompt:
        // Preview mode help
        return "↑/↓: scroll | ctrl+b: bookmark | ctrl+t: tags | Esc: close"

    case m.inputMode == ModeAddingTag:
        // Tag dialog help
        return "Enter: save | Esc: cancel"

    case m.searchMode:
        // Search mode help
        return "Type to search | Enter: apply | Esc: cancel"

    case m.activePanel == PanelFilters:
        // Filters panel help
        return "↑/↓: navigate | Space: toggle | Tab: switch panel | /: search | q: quit"

    default:
        // Prompts list help
        hasMultiplePages := len(m.filteredPrompts) > m.list.Paginator.PerPage
        if hasMultiplePages {
            return "↑/↓: navigate | ←/→: pages | /: search | Enter: preview | ctrl+t: tags | ctrl+b: bookmark | q: quit"
        }
        return "↑/↓: navigate | /: search | Enter: preview | ctrl+t: tags | ctrl+b: bookmark | q: quit"
    }
}
```

### 3. Tag Display Truncation (FR-010 to FR-013)

```go
// Add to Model struct:

// Tag truncation
tagTruncateLength int               // = 25 (constant)
truncatedTags     map[string]string // fullTag -> displayTag (cached)
```

**Truncation Logic**:
```go
const MaxTagDisplayLength = 25

func truncateTag(tag string) string {
    visualWidth := lipgloss.Width(tag)
    if visualWidth <= MaxTagDisplayLength {
        return tag
    }

    // Truncate to 22 chars + "..." (3 chars) = 25 total
    runes := []rune(tag)
    truncated := ""
    currentWidth := 0

    for _, r := range runes {
        charWidth := runewidth.RuneWidth(r)
        if currentWidth + charWidth > (MaxTagDisplayLength - 3) {
            break
        }
        truncated += string(r)
        currentWidth += charWidth
    }

    return truncated + "..."
}

// Cache truncated tags on load
func (m *Model) cacheTruncatedTags(tags []string) {
    m.truncatedTags = make(map[string]string)
    for _, tag := range tags {
        m.truncatedTags[tag] = truncateTag(tag)
    }
}
```

### 4. Pagination Display (FR-014 to FR-017)

```go
// Add to Model struct:

// Pagination state
currentPage   int     // Current page number (1-indexed)
totalPages    int     // Total number of pages
pageSize      int     // Items per page
```

**Pagination Calculation**:
```go
func (m *Model) updatePagination() {
    totalItems := len(m.filteredPrompts)
    m.pageSize = m.list.Paginator.PerPage

    if m.pageSize == 0 {
        m.pageSize = 1 // Avoid division by zero
    }

    m.totalPages = (totalItems + m.pageSize - 1) / m.pageSize
    if m.totalPages == 0 {
        m.totalPages = 1
    }

    m.currentPage = m.list.Paginator.Page + 1 // Convert 0-indexed to 1-indexed
}

func (m *Model) getPaginationText() string {
    return fmt.Sprintf("%d/%d", m.currentPage, m.totalPages)
}
```

### 5. Preview Dialog Sizing (FR-018 to FR-022)

```go
// Add to Model struct:

// Preview sizing
previewMinHeight    int   // = 15 lines
previewMaxHeightPct float64 // = 0.5 (50% of terminal)
previewDynamicHeight int   // Calculated height for current preview
```

**Dynamic Height Calculation**:
```go
func (m *Model) calculatePreviewHeight(contentLines int) int {
    const (
        minHeight = 15
        maxHeightPercent = 0.5
        dialogOverhead = 6  // Border, title, padding
    )

    // Calculate max allowed height
    maxAllowedHeight := int(float64(m.height) * maxHeightPercent)

    // Desired height based on content
    desiredHeight := contentLines + dialogOverhead

    // Apply constraints
    height := desiredHeight
    if height < minHeight {
        height = minHeight
    }
    if height > maxAllowedHeight {
        height = maxAllowedHeight
    }

    return height
}
```

### 6. Bookmark Management in Preview (FR-023 to FR-027)

```go
// Add to Model struct:

// Bookmark management state
bookmarkChanged bool  // Flag to refresh list after bookmark change
```

**State Update Pattern**:
```go
// In preview mode, on ctrl+b or ctrl+x
func (m *Model) toggleBookmarkInPreview() tea.Cmd {
    if m.currentPrompt == nil {
        return nil
    }

    promptID := m.currentPrompt.ID

    // Check current bookmark status
    isBookmarked := m.bookmarks[promptID]

    var err error
    if isBookmarked {
        // Remove bookmark
        err = storage.RemoveBookmark(promptID)
        if err == nil {
            delete(m.bookmarks, promptID)
            m.setStatus("✓ Bookmark removed", 2*time.Second)
        }
    } else {
        // Add bookmark
        err = storage.AddBookmark(m.currentPrompt)
        if err == nil {
            m.bookmarks[promptID] = true
            m.setStatus("✓ Bookmarked", 2*time.Second)
        }
    }

    if err != nil {
        m.setStatus(fmt.Sprintf("Error: %v", err), 3*time.Second)
    }

    // Flag for list refresh (but stay in preview)
    m.bookmarkChanged = true

    return nil
}
```

### 7. Tag Dialog with Existing Tags (FR-028 to FR-032)

```go
// No new state needed - modify existing tag dialog rendering

// Modify ModeAddingTag rendering to show existing tags:
func (m *Model) renderTagDialog() string {
    if m.currentPrompt == nil {
        return ""
    }

    // Get existing tags
    existingTags := m.userTags[m.currentPrompt.ID]

    var existingTagsText string
    if len(existingTags) == 0 {
        existingTagsText = "No tags assigned"
    } else {
        existingTagsText = "Current tags: " + strings.Join(existingTags, ", ")
    }

    // Render dialog with existing tags display
    content := lipgloss.JoinVertical(
        lipgloss.Left,
        lipgloss.NewStyle().Bold(true).Render("Manage Tags"),
        "",
        lipgloss.NewStyle().Faint(true).Render(existingTagsText),
        "",
        "Add tags (comma-separated):",
        m.input.View(),
    )

    // ... rest of dialog rendering with overlay.Composite()
}
```

## State Machine Diagram

```
┌─────────────┐
│   Normal    │ ◄─────────────────┐
│   Mode      │                   │
└─────────────┘                   │
      │                           │
      │ "/" key                   │ Esc
      ▼                           │
┌─────────────┐                   │
│   Search    ├───────────────────┘
│   Mode      │ Enter (apply search)
└─────────────┘
      │
      │ Enter on prompt
      ▼
┌─────────────┐
│   Preview   │ ◄─────────────────┐
│   Mode      │                   │
└─────────────┘                   │
      │                           │
      │ ctrl+t                    │ Esc / Enter (save)
      ▼                           │
┌─────────────┐                   │
│   Tag       ├───────────────────┘
│   Dialog    │
└─────────────┘
```

**Note**: Bookmark management (ctrl+b, ctrl+x) can happen from Preview Mode without state transition.

## Data Flow

### Filter Chain (unchanged logic, search added to chain)

```
All Prompts (from index)
    │
    ├─ Source Filter (if selectedSource != "")
    │
    ├─ Tag Filter (if any selectedTags == true)
    │
    ├─ Bookmark Filter (if bookmarkedOnly == true)
    │
    └─ Search Filter (if searchQuery != "") ◄─── NEW
        │
        ▼
    Filtered Prompts → Display in List
```

### Search Flow

```
User presses "/"
    │
    ├─ Enable searchMode
    ├─ Backup current filteredPrompts → preSearchList
    └─ Focus searchInput
        │
        User types...
            │
            ├─ Update searchQuery
            └─ Apply search to preSearchList → filteredPrompts
                │
                Update list display (real-time)

        User presses Enter
            │
            └─ Disable searchMode, keep current filteredPrompts

        User presses Esc
            │
            └─ Disable searchMode, restore preSearchList → filteredPrompts
```

## Performance Considerations

### Search Performance
- **Target**: <100ms for search filter application
- **Optimization**: Simple string matching (no regex, no fuzzy)
- **Typical data**: 200-500 prompts after filters
- **Worst case**: 1000 prompts (still fast with simple string.Contains)

### Filter Application
- **Target**: <50ms UI update
- **Current**: Fast (already implemented)
- **Search adds**: Minimal overhead (one additional filter pass)

### Memory
- **Search state**: ~2KB (searchQuery string + searchedPrompts slice reference)
- **Tag truncation cache**: ~5KB (200 tags * 25 bytes avg)
- **Total new memory**: <10KB overhead

## Edge Cases

### Empty Results
- When search returns zero results:
  - Show "No prompts match your search" message
  - Keep search UI active (allow user to modify query)
  - Esc restores pre-search results

### Tag Truncation Boundaries
- Tag exactly 25 chars: Show full tag, no ellipsis
- Tag 26 chars: Truncate to "first-twenty-two-chars..."
- Tag with emoji/CJK: Use `runewidth` for accurate visual width
- Tag with ANSI codes: Strip ANSI before measuring width

### Preview Height Edge Cases
- Terminal < 30 lines high: Use minimum 15 lines (may overflow, viewport scrolls)
- Content < 15 lines: Use content height (smaller dialog)
- Terminal resized while preview open: Recalculate height on tea.WindowSizeMsg

### Pagination Single Page
- When filteredPrompts fit on one page:
  - Show "1/1" (keeps layout consistent)
  - Hide ←/→ navigation from help text
  - Page navigation keys do nothing (no visual change)

### Bookmark State Sync
- User removes last bookmark while "Bookmarked" filter active:
  - List becomes empty
  - Show "No bookmarked prompts" message
  - User can disable bookmark filter to see prompts again

## Testing State Transitions

### Critical Paths to Test

1. **Search → Cancel**:
   - Filter prompts → press "/" → type query → press Esc
   - Verify: Original filtered list restored

2. **Search → Apply**:
   - Filter prompts → press "/" → type query → press Enter
   - Verify: Search results persist, searchMode disabled

3. **Preview → Bookmark → List Refresh**:
   - Preview prompt → press ctrl+b → close preview
   - Verify: Bookmark icon appears in list

4. **Preview → Tag → List Refresh**:
   - Preview prompt → press ctrl+t → add tags → save
   - Verify: Tags appear in prompt metadata

5. **Filter + Search Combination**:
   - Select source filter → select tag filter → activate search
   - Verify: Search only searches within filtered results

## Summary

All state extensions are additive - no breaking changes to existing Model structure. The enhancements follow existing patterns (input modes, filter chains, overlay dialogs) and integrate cleanly with current architecture.

**Primary modification area**: `internal/tui/finder.go`
**Lines of new code**: ~300-400 lines
**State overhead**: <10KB memory
**Performance impact**: Negligible (<100ms for all operations)
