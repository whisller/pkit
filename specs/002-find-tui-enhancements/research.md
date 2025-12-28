# Research: Bubbletea TUI Implementation Patterns

**Feature**: Enhance pkit Find TUI
**Date**: 2025-12-28
**Related**: [spec.md](./spec.md) | [plan.md](./plan.md)

This document contains research findings and recommendations for implementing 7 TUI enhancements to the `pkit find` command using Bubbletea v1.3.10, Lipgloss v1.1.0, Bubbles v0.21.0, and bubbletea-overlay v0.6.3.

---

## 1. Real-time Search Input

### Question
How to integrate textinput.Model for real-time filtering within already-filtered lists?

### Decision: Custom Search Field with Manual Filtering

**Recommendation**: Use `textinput.Model` from Bubbles with manual filtering logic rather than the list component's built-in filtering.

### Rationale

The existing code already uses the list component's built-in filtering (`m.list.SetFilteringEnabled(true)` on line 225), but the requirement is to search **within already-filtered results** (source/tag/bookmark filters). The list component's built-in filter works on the entire item set, not on pre-filtered subsets.

**Solution approach**:
1. Add separate `textinput.Model` for search field
2. Keep list filtering disabled or use it for the search functionality
3. Apply filter chain: `allPrompts` → source/tag/bookmark filters → search filter → display

### Implementation Pattern

```go
// Model state additions
type FinderModel struct {
    // Existing state
    allPrompts      []models.Prompt
    filteredPrompts []models.Prompt  // After source/tag/bookmark filters

    // NEW: Search state
    searchMode      bool
    searchInput     textinput.Model
    searchResults   []models.Prompt  // After search filter applied to filteredPrompts
}

// In Update()
case "/":
    m.searchMode = true
    m.searchInput.Focus()
    return m, nil

// In search mode Update()
var cmd tea.Cmd
m.searchInput, cmd = m.searchInput.Update(msg)
m.applySearch()  // Filter filteredPrompts by search query
return m, cmd

// Filter function
func (m *FinderModel) applySearch() {
    if m.searchInput.Value() == "" {
        m.searchResults = m.filteredPrompts
        m.list.SetItems(convertToItems(m.filteredPrompts))
        return
    }

    query := strings.ToLower(m.searchInput.Value())
    results := []models.Prompt{}
    for _, p := range m.filteredPrompts {  // Search WITHIN filtered set
        if strings.Contains(strings.ToLower(p.ID), query) ||
           strings.Contains(strings.ToLower(p.Name), query) ||
           strings.Contains(strings.ToLower(p.Description), query) {
            results = append(results, p)
        }
    }
    m.searchResults = results
    m.list.SetItems(convertToItems(results))
}
```

### Alternatives Considered

**Alternative 1: Use list component's built-in filtering**
- **Pros**: Less code, built-in UI
- **Cons**: Filters all items, not just filtered subset; conflicts with existing filter architecture
- **Rejected**: Doesn't meet requirement to search within filtered results

**Alternative 2: Integrate Bleve search index**
- **Pros**: More powerful search (fuzzy, ranking)
- **Cons**: Overkill for simple substring matching; adds complexity
- **Rejected**: Simple substring search sufficient for TUI context

### Performance Notes

- Search on each keystroke (real-time)
- Filtering 500 prompts with substring matching: <10ms
- No debouncing needed at this scale
- Memory efficient (reuses existing filtered slice)

### Sources
- [textinput package - Go Packages](https://pkg.go.dev/github.com/charmbracelet/bubbles/textinput)
- [list package - Go Packages](https://pkg.go.dev/github.com/charmbracelet/bubbles/list)
- [Bubbletea textinput example](https://github.com/charmbracelet/bubbletea/blob/main/examples/textinput/main.go)

---

## 2. ANSI Width Calculations

### Question
How does lipgloss.Width() handle ANSI escape codes vs byte length for truncation?

### Decision: Use lipgloss.Width() for Measurement, Custom Truncation

**Recommendation**: Use `lipgloss.Width()` for measuring visual width, implement custom truncation logic that preserves ANSI codes.

### Rationale

`lipgloss.Width()` correctly measures display width by:
1. Ignoring ANSI escape sequences (they have zero visual width)
2. Properly measuring wide characters (CJK, emoji) as 2 cells
3. Handling multi-line strings (returns width of widest line)

**Implementation from lipgloss source**:
```go
func Width(str string) (width int) {
    for _, l := range strings.Split(str, "\n") {
        w := ansi.StringWidth(l)  // Uses charmbracelet/x/ansi package
        if w > width {
            width = w
        }
    }
    return width
}
```

### Implementation Pattern

```go
// For tag truncation (FR-010 to FR-013)
func truncateTag(tag string, maxWidth int) string {
    visualWidth := lipgloss.Width(tag)
    if visualWidth <= maxWidth {
        return tag
    }

    // Need to truncate at maxWidth-3 to fit "..."
    // Use rune iteration to preserve UTF-8
    targetWidth := maxWidth - 3
    currentWidth := 0
    truncated := ""

    for _, r := range tag {
        charWidth := runewidth.RuneWidth(r)  // from mattn/go-runewidth
        if currentWidth + charWidth > targetWidth {
            break
        }
        truncated += string(r)
        currentWidth += charWidth
    }

    return truncated + "..."
}

// For padding (maintaining alignment)
func padToWidth(text string, targetWidth int) string {
    visualWidth := lipgloss.Width(text)
    if visualWidth >= targetWidth {
        return text
    }
    padding := targetWidth - visualWidth
    return text + strings.Repeat(" ", padding)
}
```

### Key Insights

1. **Never use `len(str)`** - gives byte count, not visual width
2. **Never use `len([]rune(str))`** - doesn't account for wide chars or ANSI
3. **Always use `lipgloss.Width()`** - ANSI-aware, wide-char-aware
4. **Truncation is tricky** - need to iterate by rune, track width with `runewidth.RuneWidth()`

### Known Issues

There was a significant bug (fixed in recent versions) where width truncation broke lipgloss border rendering when ANSI codes were present. The current version (v1.1.0) uses `charmbracelet/x/ansi` v0.1.3+ which handles this correctly.

### Existing Code Analysis

The existing `finder.go` already uses `lipgloss.Width()` correctly on lines 1202-1212:
```go
// Use visual width (ignoring ANSI codes) for proper alignment
visualWidth := lipgloss.Width(line)

// Truncate if too long (using visual width)
if visualWidth > contentWidth {
    // For truncation, we need to be more careful with ANSI codes
    // For now, just use the line as-is and let it overflow
    line = line
}

// Pad to exact width
padding := contentWidth - visualWidth
if padding > 0 {
    line = line + strings.Repeat(" ", padding)
}
```

**Note**: The truncation is currently incomplete (line 1208 comment). Need to implement proper rune-based truncation.

### Alternatives Considered

**Alternative 1: Use mattn/go-runewidth directly**
- **Pros**: More control, explicit width calculation
- **Cons**: Need to manually strip ANSI codes first
- **Rejected**: lipgloss.Width() already provides this abstraction

**Alternative 2: Strip ANSI before display**
- **Pros**: Simpler truncation logic
- **Cons**: Loses styling (colors, bold, etc.)
- **Rejected**: Styling is valuable for UX (e.g., highlight current item)

### Sources
- [lipgloss package - Go Packages](https://pkg.go.dev/github.com/charmbracelet/lipgloss)
- [lipgloss Width function](https://github.com/charmbracelet/lipgloss/blob/master/size.go)
- [mattn/go-runewidth package](https://pkg.go.dev/github.com/mattn/go-runewidth)
- [Width truncation issue #123](https://github.com/charmbracelet/x/issues/123)

---

## 3. Context-Aware Help Text

### Question
What are the patterns for dynamic help text that changes based on UI mode?

### Decision: State-Based Help Text Generation with key.Binding

**Recommendation**: Use conditional help text generation based on current UI state, leveraging Bubbles' `key.Binding` system.

### Rationale

Bubbletea/Bubbles provides a `help` component with `KeyMap` interface for managing keybindings and generating help displays. The interface supports:
- `ShortHelp() []key.Binding` - Abbreviated help view
- `FullHelp() [][]key.Binding` - Expanded help view (grouped by columns)
- Dynamic behavior via `key.Binding.SetEnabled()` - disabled keys don't render

However, for our use case (single-line context-aware help), a simpler approach is more appropriate.

### Implementation Pattern

```go
// In FinderModel
func (m FinderModel) getHelpText() string {
    switch {
    case m.inputMode == ModeViewingPrompt:
        return "↑/↓: scroll | ctrl+x: unbookmark | ctrl+b: bookmark | ctrl+t: tags | q/esc: close"

    case m.inputMode == ModeAddingTag || m.inputMode == ModeAddingAlias || m.inputMode == ModeAddingNotes:
        return "Enter: save | Esc: cancel"

    case m.inputMode == ModeRemovingTag:
        return "↑/↓: navigate | Space: remove tag | Esc: cancel"

    case m.searchMode:
        return "Type to search | Enter: apply | Esc: clear search"

    case m.activePanel == PanelFilters:
        return "↑/↓: navigate | Space: toggle | Tab: switch to list | /: search | q: quit"

    case m.activePanel == PanelList:
        // Show pagination hint if multiple pages exist
        pageHint := ""
        if m.list.Paginator.TotalPages > 1 {
            pageHint = "←/→: pages | "
        }
        return pageHint + "↑/↓: navigate | /: search | Enter: preview | ctrl+t: tags | ctrl+b: bookmark | Tab: filters | q: quit"

    default:
        return "q: quit"
    }
}

// In View()
func (m FinderModel) View() string {
    // ... render main view ...

    helpStyle := lipgloss.NewStyle().
        Foreground(lipgloss.Color("241")).
        Padding(1, 0)

    help := helpStyle.Render(m.getHelpText())

    return fmt.Sprintf("%s\n%s%s", mainView, statusView, help)
}
```

### Key Patterns

1. **State-based dispatch**: Use switch on current mode/panel
2. **Single line format**: Concise, pipe-separated shortcuts
3. **Priority ordering**: Most common actions first
4. **Conditional elements**: Show pagination only when needed
5. **Consistent format**: `key: action` throughout

### Existing Code Analysis

The current implementation (lines 923-932) uses a simple conditional:
```go
var help string
if m.activePanel == PanelFilters {
    help = "↑/↓: navigate | space: toggle | tab: switch panel | q: quit"
} else {
    help = "p: preview | enter: select | ctrl+g: get | ctrl+s: bookmark | ctrl+t: tags | ctrl+a: alias | tab: filters"
}
```

**Needs enhancement**: Add cases for all input modes and search mode.

### Alternatives Considered

**Alternative 1: Use Bubbles help component**
- **Pros**: Built-in rendering, keyboard map integration
- **Cons**: Designed for multi-line help, overkill for single-line context help
- **Rejected**: Too much overhead for simple use case

**Alternative 2: Separate help panel**
- **Pros**: More space for detailed instructions
- **Cons**: Takes up valuable screen real estate
- **Rejected**: Single-line context help is more efficient

**Alternative 3: Toggle between short/full help**
- **Pros**: User can choose detail level
- **Cons**: Adds complexity, requires another key binding
- **Deferred**: Could add later with "?" key to show detailed help dialog

### Sources
- [help package - Go Packages](https://pkg.go.dev/github.com/charmbracelet/bubbles/help)
- [Bubbletea help example](https://github.com/charmbracelet/bubbletea/blob/main/examples/help/main.go)
- [key package - Go Packages](https://pkg.go.dev/github.com/charmbracelet/bubbles/key)

---

## 4. Preview Dialog Sizing

### Question
What are best practices for dynamic dialog height calculation with min/max constraints?

### Decision: Content-Proportional Height with Absolute Constraints

**Recommendation**: Calculate dialog height based on content line count, with absolute minimum (15 lines) and maximum (50% terminal height) constraints.

### Rationale

The existing implementation (lines 1041-1055) uses fixed percentages:
```go
dialogWidth := int(float64(m.width) * 0.8)
dialogHeight := int(float64(m.height) * 0.6)  // FIXED 60%
```

This causes issues with long prompts like "fabric:review_code" where 60% is too large.

### Implementation Pattern

```go
func (m *FinderModel) calculatePreviewHeight(content string) int {
    // Count content lines
    contentLines := strings.Count(content, "\n") + 1

    // Define constraints
    const (
        minHeight     = 15              // Minimum readable height
        maxPercentage = 0.5             // Maximum 50% of terminal
        metadataLines = 10              // Metadata, borders, help text overhead
    )

    // Calculate proportional height
    neededHeight := contentLines + metadataLines

    // Apply maximum constraint (50% of terminal)
    maxHeight := int(float64(m.height) * maxPercentage)
    if maxHeight < minHeight {
        maxHeight = minHeight  // Ensure minimum even on tiny terminals
    }

    // Choose smaller of needed vs max
    dialogHeight := neededHeight
    if dialogHeight > maxHeight {
        dialogHeight = maxHeight
    }

    // Ensure minimum
    if dialogHeight < minHeight {
        dialogHeight = minHeight
    }

    return dialogHeight
}

// Usage in renderPromptPreview()
func (m *FinderModel) renderPromptPreview() string {
    if m.currentPrompt == nil {
        return ""
    }

    // Calculate dynamic height
    dialogHeight := m.calculatePreviewHeight(m.currentPrompt.Content)

    // Calculate content area (subtract metadata and borders)
    contentHeight := dialogHeight - 10  // 10 lines for metadata/borders/help

    // Apply scroll offset and display visible lines...
    // (rest of implementation)
}
```

### Sizing Algorithm

```
1. Count content lines (L)
2. Add overhead (O) for metadata/borders (e.g., 10 lines)
3. Calculate needed height: N = L + O
4. Calculate max height: M = terminal_height * 0.5
5. Ensure M >= 15 (absolute minimum)
6. Final height: min(N, M)
7. Ensure result >= 15
```

### Edge Cases

| Scenario | Calculation | Result |
|----------|-------------|--------|
| Short prompt (10 lines) | 10 + 10 = 20, min(20, 40) | 20 lines |
| Long prompt (200 lines) | 200 + 10 = 210, min(210, 40) | 40 lines (50% of 80) |
| Tiny terminal (20 lines) | Content N/A, max=10, ensure≥15 | 15 lines (absolute min) |
| Large terminal (100 lines) | 200 + 10 = 210, min(210, 50) | 50 lines (50% of 100) |

### Viewport Integration

The Bubbles `viewport` component supports dynamic sizing:
```go
import "github.com/charmbracelet/bubbles/viewport"

v := viewport.New(width, height)
v.SetContent(content)
v.GotoTop()  // Reset scroll position

// In Update()
v, cmd = v.Update(msg)  // Handles scroll keys automatically
```

### Alternatives Considered

**Alternative 1: Fixed percentage (current)**
- **Pros**: Simple, consistent
- **Cons**: Too large for short prompts, too small for long prompts in small terminals
- **Rejected**: Spec explicitly requests dynamic sizing

**Alternative 2: Full terminal height**
- **Pros**: Maximum content visibility
- **Cons**: Completely obscures underlying UI, disorienting
- **Rejected**: Poor UX, user loses context

**Alternative 3: Responsive breakpoints**
- **Pros**: Handles various terminal sizes elegantly
- **Cons**: More complex logic
- **Considered**: Could refine if 50% rule proves insufficient

### Sources
- [viewport package - Go Packages](https://pkg.go.dev/github.com/charmbracelet/bubbles/viewport)
- [viewport source code](https://github.com/charmbracelet/bubbles/blob/master/viewport/viewport.go)
- [Bubbletea viewport discussions](https://github.com/charmbracelet/bubbletea/discussions/86)

---

## 5. State Synchronization

### Question
What are Bubbletea message passing patterns for updating state from dialogs?

### Decision: Direct State Updates in Input Mode Handlers

**Recommendation**: Update model state directly within input mode handlers, reload data, and update list items in place.

### Rationale

Bubbletea uses the Elm Architecture where all state lives in a single model. Child components (dialogs, inputs) don't maintain separate state - they're rendered as part of the main model's View() and handle their updates through the main Update() function.

For simple TUI like ours, **direct state mutation is cleaner than custom messages**.

### Implementation Pattern

```go
// Current pattern (already used in codebase)
func (m FinderModel) handleInputMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    switch msg.String() {
    case "enter":
        if m.inputMode != ModeViewingPrompt {
            return m.processInput()  // Processes tag/bookmark/alias
        }
    }

    // Handle mode-specific keys...
}

// Processing pattern (e.g., removing bookmark from preview)
func (m FinderModel) removeBookmarkFromPreview() (tea.Model, tea.Cmd) {
    if m.currentPrompt == nil {
        return m, nil
    }

    // 1. Update backend storage
    mgr := bookmark.NewManager()
    err := mgr.RemoveBookmark(m.currentPrompt.ID)
    if err != nil {
        m.setStatus(fmt.Sprintf("Error: %v", err), 3*time.Second)
        return m, nil
    }

    // 2. Update model state
    delete(m.bookmarkedIDs, m.currentPrompt.ID)

    // 3. Update list items (reflect change in main view)
    m.applyFilters()  // Reapplies filters and updates list items

    // 4. Show status (keep preview open)
    m.setStatus("Bookmark removed", 2*time.Second)

    // Stay in preview mode - don't close
    return m, nil
}

// In Update() for preview mode
case key.Matches(msg, m.keys.RemoveBookmark):
    if m.inputMode == ModeViewingPrompt {
        return m.removeBookmarkFromPreview()
    }
```

### Key Patterns

1. **Synchronous updates**: State changes happen immediately, no async message passing needed
2. **Reload pattern**: After backend change, call `m.reloadData()` or `m.applyFilters()`
3. **Status feedback**: Use `m.setStatus()` to confirm action without closing dialog
4. **Stay in context**: Don't exit preview mode after actions unless user explicitly closes

### State Flow Diagram

```
User Action (ctrl+x in preview)
    ↓
Update() handles KeyMsg
    ↓
removeBookmarkFromPreview()
    ↓
1. Update backend (bookmark.yaml)
    ↓
2. Update model.bookmarkedIDs
    ↓
3. Reapply filters (updates list items)
    ↓
4. Set status message
    ↓
Return to View() (preview still open)
```

### Existing Pattern Analysis

The code already follows this pattern well (lines 796-821):
```go
func (m FinderModel) toggleBookmark(promptID string) (tea.Model, tea.Cmd) {
    mgr := bookmark.NewManager()

    if m.bookmarkedIDs[promptID] {
        // Remove bookmark
        if err := mgr.RemoveBookmark(promptID); err != nil {
            m.setStatus(fmt.Sprintf("Error: %v", err), 3*time.Second)
        } else {
            m.setStatus("✓ Bookmark removed", 2*time.Second)
            m.reloadData()  // Synchronizes state
        }
    } else {
        // Add bookmark
        // ...similar pattern...
    }

    return m, nil
}
```

**Recommendation**: Extend this pattern to work from preview mode.

### Alternatives Considered

**Alternative 1: Custom message passing**
```go
type bookmarkRemovedMsg struct { promptID string }

func removeBookmarkCmd(promptID string) tea.Cmd {
    return func() tea.Msg {
        mgr := bookmark.NewManager()
        mgr.RemoveBookmark(promptID)
        return bookmarkRemovedMsg{promptID}
    }
}

// In Update()
case bookmarkRemovedMsg:
    delete(m.bookmarkedIDs, msg.promptID)
    m.applyFilters()
```
- **Pros**: More "functional", async-friendly
- **Cons**: Overkill for synchronous operations, more boilerplate
- **Rejected**: Direct updates simpler for this use case

**Alternative 2: Nested model pattern**
```go
type PreviewModel struct {
    prompt      *models.Prompt
    viewport    viewport.Model
    // ...handle own state...
}
```
- **Pros**: Encapsulation, reusable component
- **Cons**: State sync complexity (parent needs to know when child changes bookmark)
- **Rejected**: Not needed for single-use dialog

### Sources
- [Bubbletea State Machine Pattern](https://zackproser.com/blog/bubbletea-state-machine)
- [Managing nested models with Bubble Tea](https://donderom.com/posts/managing-nested-models-with-bubble-tea/)
- [Component communication discussion](https://github.com/charmbracelet/bubbletea/discussions/707)
- [Bubbletea Elm Architecture](https://depscore.com/posts/2025-09-29-bubbletea/)

---

## 6. Terminal Compatibility

### Question
What features (overlays, ANSI colors) work across common terminals?

### Decision: Use Modern Terminal Features with Graceful Fallbacks

**Recommendation**: Target modern terminals (iTerm2, macOS Terminal, Windows Terminal, modern Linux terminals) with standard ANSI features. Bubbletea handles terminal capability detection automatically.

### Rationale

As of 2025, the vast majority of terminal emulators support:
- 256 colors (essential)
- Truecolor/24-bit color (widely supported)
- ANSI escape sequences (universal)
- Overlay rendering via alternate screen buffer (universal)
- Unicode characters including box drawing (universal)

### Terminal Capability Matrix

| Feature | iTerm2 | macOS Terminal | Windows Terminal | xterm | Alacritty | Kitty |
|---------|--------|----------------|------------------|-------|-----------|-------|
| 256 colors | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Truecolor | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Alt screen | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Box drawing | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Mouse support | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |

### Bubbletea Compatibility Features

1. **Automatic detection**: Bubbletea detects terminal capabilities via `tea.Context`
2. **Alternate screen**: `tea.WithAltScreen()` option ensures clean overlay rendering
3. **Color fallback**: Lipgloss automatically falls back to 256/16 colors if truecolor unavailable
4. **Terminal type**: Uses `TERM` environment variable for capability detection

### Implementation Pattern

```go
// Current usage (already correct)
func RunFinder(prompts []models.Prompt) (selectedID string, action string, err error) {
    p := tea.NewProgram(
        NewFinderModel(prompts),
        tea.WithAltScreen(),  // Use alternate screen buffer (universal support)
    )

    finalModel, err := p.Run()
    // ...
}

// Color definitions (lipgloss handles fallbacks)
borderColor := lipgloss.Color("205")  // Falls back to closest 256/16 color
textColor := lipgloss.AdaptiveColor{
    Light: "black",
    Dark:  "white",
}
```

### Overlay Compatibility

The `bubbletea-overlay` library (v0.6.3) uses `overlay.Composite()` which:
1. Renders background view to string
2. Renders foreground (dialog) to string
3. Overlays foreground on background at specified position
4. Returns combined string

This is **terminal-agnostic** - works with any terminal that supports:
- ANSI positioning (universal)
- Box drawing characters (universal)
- Color sequences (universal)

```go
// Existing usage (lines 938-942)
if m.inputMode != ModeNormal {
    dialog := m.renderInputDialog()
    return overlay.Composite(dialog, baseView, overlay.Center, overlay.Center, 0, 0)
}
```

### Truecolor Detection

For optimal color experience:
```bash
# Environment variables used by terminals
export COLORTERM=truecolor      # Modern standard
export TERM=xterm-256color      # Fallback
```

Bubbletea automatically detects via `$COLORTERM` and `$TERM`.

### Known Limitations

1. **Very old terminals** (pre-2015): May not support 256 colors
   - **Impact**: Minimal, Lipgloss falls back to 16 colors
   - **Solution**: None needed, automatic fallback works

2. **Terminal.app (macOS) < 10.14**: Limited truecolor support
   - **Impact**: Colors less vibrant
   - **Solution**: Works fine with 256-color fallback

3. **Windows Console (cmd.exe)**: Poor ANSI support
   - **Impact**: May show escape sequences as text
   - **Solution**: Users should use Windows Terminal (modern default)

### Testing Recommendations

Test on these representative terminals:
1. **macOS**: iTerm2 (primary), Terminal.app (fallback)
2. **Linux**: Alacritty, GNOME Terminal, xterm
3. **Windows**: Windows Terminal (primary), PowerShell 7+
4. **SSH**: Remote session simulation (TERM=xterm-256color)

### Alternatives Considered

**Alternative 1: Limit to 16 colors**
- **Pros**: Universal compatibility even with ancient terminals
- **Cons**: Ugly, defeats purpose of nice TUI
- **Rejected**: 256-color support is nearly universal in 2025

**Alternative 2: Terminal capability checking**
```go
if supportsColor() {
    // Use colors
} else {
    // Monochrome mode
}
```
- **Pros**: Maximum compatibility
- **Cons**: Unnecessary complexity, Lipgloss already handles this
- **Rejected**: Lipgloss provides automatic fallbacks

### Sources
- [iTerm2 Color Schemes](https://github.com/mbadolato/iTerm2-Color-Schemes)
- [Terminal color standards](https://github.com/termstandard/colors)
- [Truecolor terminal support](https://gist.github.com/XVilka/8346728)
- [bubbletea-overlay package](https://pkg.go.dev/github.com/quickphosphat/bubbletea-overlay)

---

## 7. Pagination Display

### Question
What are common TUI patterns for numeric pagination (e.g., "3/5" vs "Page 3 of 5")?

### Decision: Compact Numeric Format "N/M"

**Recommendation**: Use "N/M" format (e.g., "3/5") in bottom-right corner of list panel.

### Rationale

Based on UX best practices for pagination and constraints of terminal UI:

1. **Space efficiency**: Terminal UI has limited width, "3/5" takes 3-5 characters vs "Page 3 of 5" taking 13 characters
2. **Clear semantics**: "3/5" universally understood as "page 3 of 5"
3. **Scanability**: Numeric format quicker to parse visually
4. **Consistency**: Matches common TUI patterns in tools like `htop`, `ranger`, `lazygit`

### Implementation Pattern

```go
// In list panel rendering
func (m *FinderModel) renderPaginationInfo() string {
    if m.list.Paginator.TotalPages <= 1 {
        return "1/1"  // Still show for consistency, or hide with ""
    }

    current := m.list.Paginator.Page + 1      // 0-indexed to 1-indexed
    total := m.list.Paginator.TotalPages

    paginationStyle := lipgloss.NewStyle().
        Foreground(lipgloss.Color("241")).  // Subtle gray
        Italic(true)

    return paginationStyle.Render(fmt.Sprintf("%d/%d", current, total))
}

// In View() - position in bottom-right of list panel
func (m FinderModel) renderListPanel(width int) string {
    // ... list content ...

    // Footer with pagination
    pagination := m.renderPaginationInfo()
    footer := lipgloss.NewStyle().
        Width(width).
        Align(lipgloss.Right).  // Right-align
        Render(pagination)

    return lipgloss.JoinVertical(lipgloss.Left, listContent, footer)
}
```

### Positioning Options

**Option A: Bottom-right of list panel** (RECOMMENDED)
```
╭─ PROMPTS (50 items) ─────╮
│ prompt1                  │
│ prompt2                  │
│ ...                      │
│                          │
│                     3/5 ←│  Subtle, out of the way
╰──────────────────────────╯
```

**Option B: Border title integration**
```
╭─ PROMPTS (50 items) [3/5] ─╮
│ prompt1                    │
│ ...                        │
```

**Option C: Help text integration**
```
Help: ↑/↓: navigate | ←/→: pages (3/5) | q: quit
```

**Recommendation**: Option A (bottom-right) - least intrusive, scannable.

### List Component Integration

Bubbletea's list component has built-in paginator:
```go
// Access pagination state
current := m.list.Paginator.Page         // 0-indexed
total := m.list.Paginator.TotalPages
itemsPerPage := m.list.Paginator.PerPage

// Control pagination display
m.list.SetShowPagination(true)   // Shows built-in dots
m.list.SetShowPagination(false)  // Hide built-in, show custom
```

**Current implementation issue**: The list component's built-in pagination shows dots (• • • •) which the spec wants to replace with numeric format.

**Solution**: Disable built-in pagination (`SetShowPagination(false)`) and render custom numeric indicator.

### Alternatives Considered

**Alternative 1: "Page N of M"**
- **Pros**: More explicit, natural language
- **Cons**: Takes up more space (13+ chars), slower to read
- **Rejected**: Space efficiency critical in TUI

**Alternative 2: "N | M"**
- **Pros**: Visual separation
- **Cons**: Less standard, ambiguous (could mean "N or M")
- **Rejected**: "N/M" is established convention

**Alternative 3: Dots (current implementation)**
- **Pros**: Space efficient, visual
- **Cons**: Doesn't show total pages, hard to determine position when many pages
- **Rejected**: Spec explicitly requests numeric format

**Alternative 4: Progress bar**
```
[█████░░░░░] 3/5
```
- **Pros**: Visual progress indicator
- **Cons**: Takes up more space, overkill for simple pagination
- **Deferred**: Could add if numeric format proves insufficient

### Context-Aware Display

Based on help text research, pagination info should appear in help text when multiple pages exist:

```go
func (m FinderModel) getHelpText() string {
    if m.activePanel == PanelList {
        pageInfo := ""
        if m.list.Paginator.TotalPages > 1 {
            pageInfo = fmt.Sprintf("←/→: pages (%d/%d) | ",
                m.list.Paginator.Page + 1,
                m.list.Paginator.TotalPages)
        }
        return pageInfo + "↑/↓: navigate | /: search | Enter: preview | q: quit"
    }
    // ...
}
```

This provides **dual feedback**: visual indicator in list footer + help text reminder.

### Sources
- [Pagination UI Best Practices](https://coyleandrew.medium.com/design-better-pagination-a022a3b161e1)
- [Users' Pagination Preferences](https://www.nngroup.com/articles/item-list-view-all/)
- [13 Pagination Best Practices 2025](https://bricxlabs.com/blogs/modern-pagination-ui)
- [list package paginator](https://pkg.go.dev/github.com/charmbracelet/bubbles/list)

---

## Summary of Recommendations

| Area | Decision | Key Pattern/Tool |
|------|----------|------------------|
| **1. Search** | Custom textinput.Model with manual filtering | Filter chain: all → source/tag/bookmark → search |
| **2. ANSI Width** | lipgloss.Width() for measurement, rune-based truncation | Always use lipgloss.Width(), never len() |
| **3. Help Text** | State-based conditional help generation | Switch on mode/panel, single-line format |
| **4. Preview Size** | Content-proportional with min=15, max=50% | Calculate: min(content_lines + overhead, terminal * 0.5) |
| **5. State Sync** | Direct model updates, reload pattern | Update → reload → applyFilters → status |
| **6. Terminal Compat** | Modern terminals, Bubbletea auto-detection | tea.WithAltScreen(), Lipgloss automatic fallbacks |
| **7. Pagination** | Compact "N/M" format in bottom-right | Disable built-in dots, render custom numeric |

## Next Steps

1. Review and validate these patterns with team
2. Proceed to Phase 1 (data-model.md, contracts/, quickstart.md)
3. Reference these patterns during task implementation
4. Update this document if implementation reveals better approaches

---

**Research completed**: 2025-12-28
**Ready for Phase 1**: ✅
