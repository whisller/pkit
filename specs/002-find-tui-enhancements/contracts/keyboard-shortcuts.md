# Keyboard Shortcuts Contract

**Feature**: Enhance pkit Find TUI
**Branch**: 002-find-tui-enhancements
**Date**: 2025-12-28

## Overview

This contract defines all keyboard shortcuts and their behavior across different modes in the `pkit find` TUI. Modifications extend existing shortcuts without breaking changes.

## Shortcut Reference Table

| Key | Mode | Action | Requirements |
|-----|------|--------|--------------|
| `/` | Normal | Activate search mode | FR-005 |
| `↑` / `↓` | Normal | Navigate prompts | FR-006 |
| `←` / `→` | Normal | Navigate pages (if multiple pages) | FR-007 |
| `Enter` | Normal | Preview selected prompt | Existing |
| `ctrl+t` | Normal | Open tag dialog for selected prompt | Existing (enhanced FR-028) |
| `ctrl+b` | Normal | Toggle bookmark for selected prompt | Existing |
| `Tab` | Normal | Switch between filters/prompts panel | Existing |
| `q` | Normal | Quit application | Existing |
| | | | |
| `Esc` | Search | Exit search, restore filtered list | FR-004 |
| `Enter` | Search | Apply search, exit search mode | FR-004 |
| (typing) | Search | Real-time filter prompts | FR-001, FR-002 |
| | | | |
| `↑` / `↓` | Preview | Scroll preview content | Existing |
| `j` / `k` | Preview | Scroll preview content (vim-style) | Existing |
| `ctrl+x` | Preview | Remove bookmark (if bookmarked) | FR-023, FR-024 |
| `ctrl+b` | Preview | Toggle bookmark (add/remove) | FR-026 |
| `ctrl+t` | Preview | Open tag dialog | FR-028 (new in preview) |
| `Esc` | Preview | Close preview | Existing |
| | | | |
| `Enter` | Tag Dialog | Save tags | FR-031 |
| `Esc` | Tag Dialog | Cancel tag editing | Existing |
| (typing) | Tag Dialog | Enter comma-separated tags | FR-031 |

## Mode-Specific Behavior

### Normal Mode (Main List View)

**Context**: User is navigating the prompts list or filters panel.

**Primary Actions**:
```
/           → Activate search mode
↑/↓         → Navigate through prompts
←/→         → Navigate between pages (when multiple pages exist)
Enter       → Preview selected prompt
ctrl+t      → Open tag dialog for selected prompt
ctrl+b      → Toggle bookmark status for selected prompt
Tab         → Switch focus between filters panel and prompts list
q           → Quit application
```

**Conditional Visibility**:
- `←/→` page navigation: Only shown in help text when `totalPages > 1`
- Tag filter shortcuts: Only effective when in filters panel

**Implementation Reference**: `internal/tui/finder.go`, `Update()` method, `tea.KeyMsg` handling

### Search Mode

**Context**: User has pressed `/` and search input is active.

**Primary Actions**:
```
(typing)    → Real-time filter currently displayed prompts
Enter       → Apply search, return to Normal mode (keep filtered results)
Esc         → Cancel search, restore pre-search filtered list, return to Normal mode
```

**Search Behavior**:
- **Input**: Text entered into search field
- **Scope**: Searches within current `filteredPrompts` (already filtered by source/tags/bookmarks)
- **Fields**: Matches against prompt ID, name, and description
- **Case sensitivity**: Case-insensitive matching
- **Update frequency**: Real-time (every keystroke)
- **Empty results**: Shows "No prompts match your search" message

**Visual Feedback**:
- Search input field visible at top or bottom of prompts panel
- Help text changes to: `"Type to search | Enter: apply | Esc: cancel"`
- Prompt count updates in real-time: `"PROMPTS (X items)"`

**State Persistence**:
- Enter: Search results persist after exiting search mode
- Esc: Restores pre-search filtered list

**Implementation Reference**: New in `Update()` method, `searchMode` flag

### Preview Mode

**Context**: User has pressed Enter on a prompt to view full content.

**Primary Actions**:
```
↑/↓ or j/k  → Scroll preview content up/down
ctrl+x      → Remove bookmark (if prompt is currently bookmarked) [NEW]
ctrl+b      → Toggle bookmark (add if not bookmarked, remove if bookmarked) [NEW]
ctrl+t      → Open tag dialog (shows existing tags, allows adding more) [NEW]
Esc         → Close preview, return to Normal mode
```

**Bookmark Management (FR-023 to FR-027)**:
- **ctrl+x**: Quick removal shortcut
  - Only effective if prompt is currently bookmarked
  - Shows status message: `"✓ Bookmark removed"`
  - Updates bookmark status immediately (no need to close preview)
  - Prompt remains in list unless bookmark filter is active

- **ctrl+b**: Toggle shortcut
  - Adds bookmark if not bookmarked: `"✓ Bookmarked"`
  - Removes bookmark if bookmarked: `"✓ Bookmark removed"`
  - Same behavior as ctrl+x when removing

**Tag Management (FR-028 to FR-032)**:
- **ctrl+t**: Opens tag dialog overlay
  - Shows existing tags: `"Current tags: code, review"` or `"No tags assigned"`
  - Input field for adding new tags (comma-separated)
  - Enter saves, Esc cancels
  - Preview remains open after tag dialog closes

**Visual Feedback**:
- Dialog overlays on preview (using `bubbletea-overlay`)
- Status messages appear at bottom of preview
- Bookmark icon (★) updates in list when preview closes
- Help text: `"↑/↓: scroll | ctrl+b: bookmark | ctrl+t: tags | Esc: close"` (FR-027)

**Implementation Reference**: New keyboard handling in `inputMode == ModeViewingPrompt` section

### Tag Dialog Mode

**Context**: User has pressed ctrl+t to manage tags (from Normal or Preview mode).

**Primary Actions**:
```
(typing)    → Enter comma-separated tag names
Enter       → Save tags, close dialog
Esc         → Cancel, close dialog without saving
```

**Display (FR-028 to FR-032)**:
- **Existing tags section**:
  - If tags exist: `"Current tags: tag1, tag2, tag3"`
  - If no tags: `"No tags assigned"`
  - Long tag lists wrap for readability

- **Input field**: Text input for adding tags
  - Placeholder: `"e.g., code, review, python"`
  - Format: Comma-separated list

**Tag Addition Behavior (FR-031)**:
- New tags are ADDED to existing tags (not replaced)
- Duplicate tags are automatically filtered
- Empty/whitespace-only tags are ignored
- Tags are trimmed of leading/trailing whitespace

**Example Flow**:
```
Existing tags: ["code", "go"]
User enters: "review, security"
Result: ["code", "go", "review", "security"]
```

**Visual Layout**:
```
┌─ Manage Tags ────────────────┐
│                              │
│ Current tags: code, go       │
│                              │
│ Add tags (comma-separated):  │
│ ▸ review, security_________  │
│                              │
│ Enter: save | Esc: cancel    │
└──────────────────────────────┘
```

**Implementation Reference**: `inputMode == ModeAddingTag`, enhanced to show existing tags

## Help Text Context Switching (FR-008)

The help text dynamically changes based on current mode and context:

### Filters Panel Active
```
↑/↓: navigate | Space: toggle | Tab: switch panel | /: search | q: quit
```

### Prompts List Active (Multiple Pages)
```
↑/↓: navigate | ←/→: pages | /: search | Enter: preview | ctrl+t: tags | ctrl+b: bookmark | q: quit
```

### Prompts List Active (Single Page)
```
↑/↓: navigate | /: search | Enter: preview | ctrl+t: tags | ctrl+b: bookmark | q: quit
```

### Search Mode
```
Type to search | Enter: apply | Esc: cancel
```

### Preview Mode
```
↑/↓: scroll | ctrl+b: bookmark | ctrl+t: tags | Esc: close
```

### Tag Dialog Mode
```
Enter: save | Esc: cancel
```

**Implementation**: `generateHelpText()` method with mode/panel switching logic

## Conflict Resolution

### No Conflicting Shortcuts
All new shortcuts (ctrl+x, ctrl+t in preview, / for search) do not conflict with existing shortcuts in their respective modes.

### Shortcut Priority
When multiple handlers could match:
1. Input dialog handlers (highest priority)
2. Mode-specific handlers (search, preview)
3. Global handlers (q for quit)

**Example**: In search mode, typing 'q' enters 'q' into search field (doesn't quit app).

## Accessibility Considerations

### Keyboard-Only Navigation
- All functionality accessible via keyboard (no mouse required)
- Consistent navigation: ↑/↓ for vertical, ←/→ for horizontal
- Esc always cancels/closes current mode or dialog

### Visual Feedback
- Help text always visible showing current mode's shortcuts
- Status messages confirm actions (bookmark added/removed, tags saved)
- Mode transitions are instant with clear visual changes

### Screen Reader Support
- All shortcuts operate on visible UI elements
- Status messages are text-based (compatible with terminal screen readers)
- No hidden or ambiguous shortcuts

## Testing Matrix

### Critical Shortcut Combinations

| Test Case | Steps | Expected Result |
|-----------|-------|-----------------|
| Search activation | Press `/` | Search mode activates, input focused |
| Search real-time | Type "code" | List filters to matching prompts |
| Search apply | Type "review" + Enter | Results persist, normal mode |
| Search cancel | Type "test" + Esc | Original list restored |
| | | |
| Preview bookmark toggle | Enter on prompt → ctrl+b | Status: "✓ Bookmarked" |
| Preview bookmark remove | Enter on bookmarked → ctrl+x | Status: "✓ Bookmark removed" |
| Preview tag dialog | Enter on prompt → ctrl+t | Tag dialog shows existing tags |
| Preview tag add | ctrl+t → type "new" → Enter | Tags saved, preview remains open |
| | | |
| Page navigation | Filter to 50+ prompts → press → | Next page shown, pagination updates |
| Tag dialog from list | Select prompt → ctrl+t | Tag dialog opens, shows existing tags |
| Bookmark from list | Select prompt → ctrl+b | Bookmark toggled, icon updates |

### Edge Cases

| Test Case | Steps | Expected Result |
|-----------|-------|-----------------|
| Search with no results | Type non-existent term | "No prompts match" message |
| Remove last bookmark in bookmark view | Bookmark filter on → ctrl+x last item | Empty list, filter still active |
| Tag dialog on untagged prompt | ctrl+t on new prompt | "No tags assigned" shown |
| Page navigation on single page | Press → with <20 prompts | No action (shortcut not shown in help) |
| Multiple filters + search | Source + tag filter → search | Search limited to filtered subset |

## Implementation Notes

### Key Binding Registration

```go
// In finder.go keyMap struct
type keyMap struct {
    // Existing keys
    Up    key.Binding
    Down  key.Binding
    Enter key.Binding
    // ... more existing keys

    // New keys for enhancements
    Search        key.Binding  // "/"
    RemoveBookmark key.Binding // "ctrl+x"
    // ctrl+b and ctrl+t already exist, just enhanced in preview mode
}

// Initialize new bindings
func newKeyMap() keyMap {
    return keyMap{
        // ... existing bindings
        Search: key.NewBinding(
            key.WithKeys("/"),
            key.WithHelp("/", "search"),
        ),
        RemoveBookmark: key.NewBinding(
            key.WithKeys("ctrl+x"),
            key.WithHelp("ctrl+x", "remove bookmark"),
        ),
    }
}
```

### Mode-Specific Handling

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        // Priority 1: Input dialog handlers
        if m.inputMode == ModeAddingTag || m.inputMode == ModeAddingAlias {
            // Handle input dialog keys
            // ...
        }

        // Priority 2: Search mode handlers
        if m.searchMode {
            switch {
            case key.Matches(msg, m.keys.Enter):
                // Apply search
            case key.Matches(msg, m.keys.Escape):
                // Cancel search
            default:
                // Pass to search input
            }
        }

        // Priority 3: Preview mode handlers
        if m.inputMode == ModeViewingPrompt {
            switch {
            case key.Matches(msg, m.keys.ToggleBookmark):
                // ctrl+b: Toggle bookmark
            case key.Matches(msg, m.keys.RemoveBookmark):
                // ctrl+x: Remove bookmark
            case key.Matches(msg, m.keys.AddTag):
                // ctrl+t: Open tag dialog
            }
        }

        // Priority 4: Normal mode handlers
        switch {
        case key.Matches(msg, m.keys.Search):
            // "/" key activates search
        case key.Matches(msg, m.keys.Left), key.Matches(msg, m.keys.Right):
            // Page navigation
        // ... rest of normal mode keys
        }
    }
}
```

## References

- **Functional Requirements**: FR-001 to FR-032 in spec.md
- **Success Criteria**: SC-001, SC-002, SC-006, SC-007, SC-008 in spec.md
- **User Stories**: All 7 user stories in spec.md
- **Implementation**: `internal/tui/finder.go`
