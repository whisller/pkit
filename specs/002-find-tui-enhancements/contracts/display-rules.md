# Display Rules Contract

**Feature**: Enhance pkit Find TUI
**Branch**: 002-find-tui-enhancements
**Date**: 2025-12-28

## Overview

This contract defines precise visual display rules for all TUI enhancements. These rules ensure consistent layout, proper text handling, and clear visual feedback across different terminal sizes and content variations.

## Tag Truncation Rules (FR-010 to FR-013)

### Truncation Constraints

| Property | Value | Rationale |
|----------|-------|-----------|
| Maximum displayed length | 25 characters (visual width) | Prevents filter panel overflow |
| Truncation indicator | "..." (3 chars) | Universal truncation convention |
| Truncation point | 22 chars content + "..." | Maintains 25 char total |
| Measurement method | `lipgloss.Width()` | ANSI-aware visual width |
| Character handling | `runewidth.RuneWidth()` | Supports CJK and emoji |

### Truncation Algorithm

```go
const MaxTagDisplayLength = 25

func truncateTag(tag string) string {
    // Measure visual width (ignoring ANSI codes)
    visualWidth := lipgloss.Width(tag)

    // No truncation needed
    if visualWidth <= MaxTagDisplayLength {
        return tag
    }

    // Truncate at character boundaries (not byte boundaries)
    runes := []rune(tag)
    truncated := ""
    currentWidth := 0

    for _, r := range runes {
        // Get character width (1 for ASCII, 2 for CJK, varies for emoji)
        charWidth := runewidth.RuneWidth(r)

        // Stop before exceeding limit
        if currentWidth + charWidth > (MaxTagDisplayLength - 3) {
            break
        }

        truncated += string(r)
        currentWidth += charWidth
    }

    return truncated + "..."
}
```

### Visual Examples

| Original Tag | Visual Width | Displayed As | Notes |
|--------------|--------------|--------------|-------|
| `code` | 4 | `code` | No truncation |
| `this-is-exactly-25-chars` | 25 | `this-is-exactly-25-chars` | Exactly at limit |
| `this-is-a-very-long-tag-name` | 30 | `this-is-a-very-long-...` | Truncated at 22+3 |
| `日本語タグ名前長い` | 16 | `日本語タグ名前長い` | CJK counted correctly |
| `emoji-🎉-test-long-name` | 26 | `emoji-🎉-test-lon...` | Emoji width handled |

### Layout Consistency (FR-012)

**Requirement**: Checkbox alignment must not shift regardless of tag name length.

**Implementation**:
```
Filter panel width: Fixed at 30% of terminal width
Checkbox column: Always at position 2 (left margin)
Tag text column: Always starts at position 6 (after "[ ] ")
Tag max width: Filter panel width - 10 (margins + checkbox)

Example layout:
┌─ FILTERS ──────────────┐
│                        │
│ [ ] short              │
│ [x] this-is-a-very-...  │  ← Checkbox at same position
│ [ ] emoji-🎉-test-l...  │  ← Text aligned consistently
│                        │
└────────────────────────┘
```

### Full Tag Name Access (FR-013)

**Methods for accessing full tag name**:

1. **Status line display** (recommended):
   - When tag is selected/hovered: Show full name in status bar
   - Status format: `"Tag: {full_tag_name}"`
   - Duration: While tag is highlighted

2. **Tag list command** (existing):
   - User can run `pkit tag list` to see all full tag names
   - CLI output shows complete tags without truncation

**Example**:
```
User navigates to truncated tag "very-long-tag-name-..."
Status bar shows: "Tag: very-long-tag-name-that-exceeds-display-limit"
```

## Pagination Display Rules (FR-014 to FR-017)

### Pagination Format

| Property | Value | Rationale |
|----------|-------|-----------|
| Format | "N/M" | Compact, universally understood |
| Position | Bottom-right of prompts panel | Standard pagination placement |
| Example | "3/5" | Current page 3 of 5 total |
| Single page | "1/1" | Keeps layout consistent, clear state |
| Update frequency | Real-time on navigation | Immediate visual feedback |

### Display Logic

```go
func (m *Model) getPaginationText() string {
    // Always use 1-indexed page numbers (more intuitive)
    current := m.list.Paginator.Page + 1
    total := m.totalPages

    return fmt.Sprintf("%d/%d", current, total)
}
```

### Visual Positioning

```
Prompts panel layout:
┌─ PROMPTS (232 items) ──────────────────┐
│                                        │
│  1. fabric:summarize                   │
│  2. fabric:review_code                 │
│  ...                                   │
│ 20. fabric:explain                     │
│                                        │
└──────────────────────────────── 3/5 ───┘
                                   └─ Pagination here
```

**Implementation**:
- Render in bottom border line
- Right-aligned within panel width
- Use subtle color (faint or dim) to not distract
- Update on every page change (← → keys)

### Edge Cases

| Scenario | Display | Behavior |
|----------|---------|----------|
| Zero prompts | "0/0" or hide | No pagination needed |
| Exactly 20 items (1 page) | "1/1" | Consistent with spec FR-017 |
| Large page count (e.g., 23/47) | "23/47" | No abbreviation, full numbers |
| After filtering to 1 page | "1/1" | Updates dynamically |

### Replacement of Dots

**Old behavior (to remove)**:
```
Prompts panel with dots:
• • • ○ •  ← Remove this dot pagination
```

**New behavior**:
```
Replace list's built-in pagination dots with numeric format:
- Set `list.Paginator.ShowHelp = false`
- Set custom pagination rendering in panel border
```

## Preview Dialog Sizing Rules (FR-018 to FR-022)

### Sizing Constraints

| Property | Value | Rationale |
|----------|-------|-----------|
| Width | 80% of terminal width | Unchanged from current |
| Height (dynamic) | Proportional to content | Adaptive sizing |
| Maximum height | 50% of terminal height | Prevents overwhelming UI |
| Minimum height | 15 lines | Ensures readability |
| Maximum absolute | Lesser of content or max% | Respects both constraints |

### Height Calculation Algorithm

```go
func (m *Model) calculatePreviewHeight(contentLines int) int {
    const (
        MinHeight = 15
        MaxHeightPercent = 0.5
        DialogOverhead = 6  // Title (1) + top border (1) + padding (2) + bottom border (1) + status (1)
    )

    // Calculate available space
    maxAllowedHeight := int(float64(m.height) * MaxHeightPercent)

    // Desired height based on actual content
    desiredHeight := contentLines + DialogOverhead

    // Apply constraints
    height := desiredHeight

    // Ensure minimum readability
    if height < MinHeight {
        height = MinHeight
    }

    // Don't overwhelm screen
    if height > maxAllowedHeight {
        height = maxAllowedHeight
    }

    return height
}
```

### Visual Examples

**Scenario 1: Short prompt (10 lines)**
```
Terminal height: 40 lines
Content: 10 lines
Max allowed (50%): 20 lines
Calculation: 10 + 6 = 16 lines
Applied min: 16 lines (> 15 ✓)
Applied max: 16 lines (< 20 ✓)
Result: 16 lines tall dialog
```

**Scenario 2: Medium prompt (30 lines)**
```
Terminal height: 60 lines
Content: 30 lines
Max allowed (50%): 30 lines
Calculation: 30 + 6 = 36 lines
Applied max: 30 lines (< 30% limit)
Result: 30 lines tall dialog (viewport scrolls through 30 lines of content)
```

**Scenario 3: Very long prompt (150 lines)**
```
Terminal height: 50 lines
Content: 150 lines
Max allowed (50%): 25 lines
Calculation: 150 + 6 = 156 lines
Applied max: 25 lines (50% limit)
Result: 25 lines tall dialog (viewport scrolls through 150 lines)
```

**Scenario 4: Tiny terminal**
```
Terminal height: 24 lines
Content: 50 lines
Max allowed (50%): 12 lines
Calculation: 50 + 6 = 56 lines
Applied min: 15 lines (forced minimum)
Result: 15 lines tall dialog (may overflow small terminal)
```

### Scroll Indicators (FR-022)

**Requirements**:
- Must clearly show when content extends beyond visible area
- Must show at top when content above viewport
- Must show at bottom when content below viewport

**Visual Indicators**:
```
Top scroll indicator (when scrolled down):
┌─ Preview: fabric:review_code ─────┐
│ ▲ More content above ▲            │ ← Scroll indicator
│ ─────────────────────────────────│
│                                   │
│ [visible content...]              │
│                                   │

Bottom scroll indicator (when more content below):
│                                   │
│ [visible content...]              │
│                                   │
│ ─────────────────────────────────│
│ ▼ More content below ▼            │ ← Scroll indicator
└───────────────────────────────────┘
```

**Implementation**:
- Use viewport component's built-in scroll percentage
- Show indicators when `viewport.ScrollPercent() > 0` (top) or `< 1.0` (bottom)
- Include scroll hints in help text: "↑/↓: scroll"

### Width Handling (Unchanged)

**Current behavior (maintain)**:
```
Width: 80% of terminal width
Maximum: 100 characters
Minimum: 60 characters

Example (100 char terminal):
Dialog width: 80 chars (80%)

Example (50 char terminal):
Dialog width: 60 chars (minimum enforced)
```

## Help Text Display Rules (FR-006 to FR-009)

### Location and Format

| Property | Value | Rationale |
|----------|-------|-----------|
| Location | Bottom line of TUI | Standard help bar position |
| Format | Single line | Must fit terminal width |
| Separator | " \| " (space-pipe-space) | Clear visual separation |
| Style | Faint/dim or subtle color | Non-distracting |
| Content | Context-aware | Shows relevant shortcuts only |

### Context-Aware Content

**Filters Panel Active**:
```
↑/↓: navigate | Space: toggle | Tab: switch panel | /: search | q: quit
```

**Prompts List Active (Multiple Pages)**:
```
↑/↓: navigate | ←/→: pages | /: search | Enter: preview | ctrl+t: tags | ctrl+b: bookmark | q: quit
```

**Prompts List Active (Single Page)**:
```
↑/↓: navigate | /: search | Enter: preview | ctrl+t: tags | ctrl+b: bookmark | q: quit
```
*Note: ←/→ page navigation removed when not applicable*

**Search Mode**:
```
Type to search | Enter: apply | Esc: cancel
```

**Preview Mode**:
```
↑/↓: scroll | ctrl+b: bookmark | ctrl+t: tags | Esc: close
```

**Tag Dialog Mode**:
```
Enter: save | Esc: cancel
```

### Truncation for Narrow Terminals

**Priority order** (if terminal too narrow to fit all shortcuts):

1. Keep mode-specific primary actions (navigation, enter, esc)
2. Drop secondary actions (bookmark, tags)
3. Drop quit action (always available via ctrl+c)

**Example (80 char terminal)**:
```
Full: ↑/↓: navigate | ←/→: pages | /: search | Enter: preview | ctrl+t: tags | ctrl+b: bookmark | q: quit
Truncated (65 chars): ↑/↓: navigate | ←/→: pages | /: search | Enter: preview | q: quit
```

**Implementation**:
```go
func (m *Model) generateHelpText() string {
    fullHelp := m.getFullHelpForMode()

    // If fits terminal width, use full help
    if lipgloss.Width(fullHelp) <= m.width {
        return fullHelp
    }

    // Otherwise, use truncated version
    return m.getTruncatedHelpForMode()
}
```

### Visual Styling

**Recommended style**:
```go
helpStyle := lipgloss.NewStyle().
    Foreground(lipgloss.Color("240")).  // Subtle gray
    Background(lipgloss.Color("235")).  // Slightly darker background
    Padding(0, 1)                       // Horizontal padding

renderedHelp := helpStyle.Render(m.generateHelpText())
```

**Position in layout**:
```
┌─ Full TUI Layout ─────────────────┐
│                                   │
│  [Filters]     [Prompts]          │
│                                   │
│                                   │
├───────────────────────────────────┤
│ ↑/↓: navigate | Enter: preview... │ ← Help text here (bottom line)
└───────────────────────────────────┘
```

## Status Messages

### Status Message Display

**Location**: Bottom of TUI, above help text (or replace help text temporarily)

**Format**:
```
Success: "✓ {action}"
Error: "✗ {error message}"
Info: "ℹ {information}"
```

**Duration**: 2-3 seconds (auto-clear)

**Examples**:
```
✓ Bookmarked
✓ Bookmark removed
✓ Tags saved: code, review
✗ Error: Failed to save bookmark
ℹ Searching... (42 results)
```

### Visual Priority

**Layering** (bottom to top):
1. Help text (always visible)
2. Status messages (temporary overlay)
3. Dialogs (full overlay)

**Implementation**:
```
When status message active:
  Hide help text, show status
  After expiry, restore help text

When dialog active:
  Overlay entire TUI
  No help or status visible
```

## Accessibility

### Screen Reader Compatibility

**Text-based elements**:
- All display elements are plain text (no graphics)
- Status messages are text strings (screen reader compatible)
- Help text provides keyboard navigation context

**Visual width handling**:
- Use `lipgloss.Width()` for correct visual measurement
- Ensures proper layout even with ANSI styling
- Avoids misalignment that could confuse screen reader output

### High Contrast

**Recommended color scheme**:
- High contrast between text and background
- Use lipgloss color names (not raw ANSI codes)
- Test with `$NO_COLOR` environment variable (should degrade gracefully)

### Keyboard-Only Navigation

**All functionality keyboard-accessible**:
- No mouse-only features
- Clear keyboard shortcuts displayed in help text
- Consistent navigation patterns (↑/↓ for vertical, ←/→ for horizontal)

## Testing Checklist

### Visual Regression Tests

- [ ] Tag truncation at 25 chars exactly
- [ ] Tag truncation at 26 chars (shows ellipsis)
- [ ] Checkbox alignment with short tags
- [ ] Checkbox alignment with long tags
- [ ] CJK tag truncation (日本語)
- [ ] Emoji tag truncation (🎉)
- [ ] Pagination "1/1" on single page
- [ ] Pagination "3/5" on page 3 of 5
- [ ] Pagination updates on page navigation
- [ ] Preview dialog 15 lines (minimum)
- [ ] Preview dialog 50% height (maximum)
- [ ] Preview dialog proportional (medium content)
- [ ] Scroll indicators when content overflows
- [ ] Help text changes per mode
- [ ] Help text truncation on narrow terminal (40 chars)
- [ ] Status message display and auto-clear
- [ ] Layout with terminal resize

### Terminal Size Tests

| Terminal Size | Expected Behavior |
|---------------|-------------------|
| 80x24 (minimum) | All UI elements fit, minimum sizes enforced |
| 120x40 (typical) | Optimal layout, all features visible |
| 200x60 (large) | Scaling handled, no excessive whitespace |
| 40x20 (tiny) | Truncation/minimum sizes prevent corruption |

## References

- **Functional Requirements**: FR-006 to FR-013, FR-014 to FR-022 in spec.md
- **Success Criteria**: SC-002, SC-003, SC-004, SC-005 in spec.md
- **Research**: research.md sections on width calculations and dialog sizing
- **Implementation**: `internal/tui/finder.go`
