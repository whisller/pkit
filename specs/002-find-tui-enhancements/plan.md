# Implementation Plan: Enhance pkit Find TUI

**Branch**: `002-find-tui-enhancements` | **Date**: 2025-12-28 | **Spec**: [spec.md](./spec.md)

## Summary

This plan implements 7 enhancements to the `pkit find` TUI command to improve usability, navigation clarity, and interactive management workflows. The features restore lost search functionality, improve visual feedback, and enable in-TUI bookmark and tag management without leaving the interface.

**Technical Approach**: All enhancements modify the existing Bubbletea TUI implementation in `internal/tui/finder.go`. No new architectural components required - this is purely UI/UX refinement using existing Bubbletea patterns, the overlay library for dialogs, and dynamic content loading.

## Technical Context

**Language/Version**: Go 1.25.4
**Primary Dependencies**:
- Bubbletea v1.3.10 (TUI framework with Elm architecture)
- Lipgloss v1.1.0 (Terminal styling and layout)
- Bubbles v0.21.0 (Reusable TUI components - list, textinput, viewport)
- bubbletea-overlay v0.6.3 (Dialog overlay system)
- Bleve v2.5.7 (Full-text search index)

**Storage**:
- User data in `~/.pkit/` (bookmarks.yml, tags.yml, cache/)
- Source repositories cloned to `~/.pkit/sources/`
- Bleve search index in `~/.pkit/cache/`

**Testing**:
- Manual testing workflow (terminal UI requires interactive testing)
- Unit test structure: `*_test.go` files (existing: `pkg/models/validator_test.go`)
- Test coverage focus: State management logic, filter combinations, edge cases

**Target Platform**: CLI application (macOS, Linux, Windows)

**Project Type**: Single project (Go CLI with TUI interface)

**Performance Goals**:
- Search within filtered results: <100ms response time
- Filter application: <50ms UI update
- Tag truncation: instant (pure display logic)
- Preview dialog render: <100ms

**Constraints**:
- Terminal compatibility (must work on standard terminals)
- No external runtime dependencies (single binary)
- Must not break existing keyboard shortcuts
- Maintain accessibility for terminal screen readers

**Scale/Scope**:
- Single TUI file modification (`internal/tui/finder.go`)
- 7 user-facing feature enhancements
- ~500 lines of modified/added code
- No data model changes
- No API changes

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### ✅ Principle I: Organization-First Architecture
**Status**: PASS
**Rationale**: All features improve prompt organization, discovery, and management within the TUI. No execution engine changes. Features support bookmarking, tagging, filtering, and search - all organization-focused.

### ✅ Principle II: CLI-First Interface
**Status**: PASS
**Rationale**: Enhancements to TUI complement existing CLI commands. Users can still use `pkit tag add`, `pkit bookmark add` via CLI. TUI provides visual alternative for same operations. No breaking changes to CLI interface.

### ✅ Principle III: Tool Agnosticism
**Status**: PASS
**Rationale**: Features are internal TUI improvements. No impact on tool agnosticism. Output via `pkit get` remains unchanged - still pipes to any execution tool.

### ✅ Principle IV: Multi-Source Aggregation
**Status**: PASS
**Rationale**: Search and filter improvements enhance cross-source discovery. Features work identically across all subscribed sources. Source filtering maintained and improved.

### ✅ Principle V: Simple Output Protocol
**Status**: PASS
**Rationale**: TUI is interactive interface, not output protocol. `pkit get` output protocol unchanged. Changes are purely visual/interactive within TUI.

### ✅ Principle VI: Phase-Gated Development
**Status**: PASS
**Rationale**: Features align with Phase 1 (organization-only). No execution features. Bookmark and tag management are core organization features already present in CLI - TUI integration is natural extension.

### ✅ Principle VII: Simplicity & Focus
**Status**: PASS
**Rationale**:
- Features refine existing TUI, no new abstractions
- Modifications localized to single file (`internal/tui/finder.go`)
- Reuse existing Bubbletea patterns and overlay library
- No new dependencies beyond already-integrated bubbletea-overlay
- Edge cases explicitly handled in spec
- Simple display logic (truncation, pagination format, help text updates)

**Overall Constitution Status**: ✅ **PASS** - All principles satisfied, no violations requiring justification.

## Project Structure

### Documentation (this feature)

```text
specs/002-find-tui-enhancements/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output - TUI patterns, terminal capabilities
├── data-model.md        # Phase 1 output - State management structures
├── quickstart.md        # Phase 1 output - Developer guide for TUI modifications
├── contracts/           # Phase 1 output - Keyboard shortcuts, state transitions
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
cmd/pkit/                 # CLI commands (unchanged for this feature)
├── find.go               # Entry point for find command
├── bookmark_add.go       # CLI bookmark management
├── bookmark_remove.go
├── tag_add.go            # CLI tag management
├── tag_remove.go
└── ...

internal/
├── tui/
│   └── finder.go         # ⭐ PRIMARY MODIFICATION TARGET
│                         # All 7 enhancements implemented here
│                         # ~2000 lines, contains TUI state machine
│
├── index/
│   ├── indexer.go        # Bleve index management (unchanged)
│   └── search.go         # Search API (minor usage changes)
│
├── storage/
│   ├── bookmarks.go      # Bookmark persistence (used by TUI)
│   └── tags.go           # Tag persistence (used by TUI)
│
└── source/
    └── loader.go         # Prompt content loading (used for preview)

pkg/models/
└── prompt.go             # Data structures (unchanged)

tests/                    # Not yet present, manual testing workflow
└── (future unit tests)
```

**Structure Decision**: Single project structure fits well. All changes localized to `internal/tui/finder.go`. Existing storage and index layers provide necessary APIs. No architectural changes needed - pure UI refinement.

## Complexity Tracking

No constitution violations - table not applicable.

---

## Phase 0: Research & Technical Discovery

### Research Tasks

**R1: Bubbletea Search Input Patterns**
- **Question**: How to implement real-time search within filtered results using Bubbletea?
- **Scope**: Research textinput integration, filter chaining, performance patterns
- **Deliverable**: Pattern for search field that filters already-filtered prompts

**R2: Terminal Width Calculations with ANSI Codes**
- **Question**: How to reliably truncate and pad text containing ANSI escape codes?
- **Scope**: Lipgloss width utilities, visual vs byte length handling
- **Deliverable**: Best practices for tag truncation and layout consistency
- **Note**: Initial implementation used `lipgloss.Width()` - validate this is correct approach

**R3: Help Text Context Awareness**
- **Question**: How to show different help text based on active panel/mode?
- **Scope**: Conditional help rendering, concise formatting for terminal space
- **Deliverable**: Pattern for dynamic help text that updates with UI state

**R4: Preview Dialog Sizing Strategies**
- **Question**: How to dynamically size dialogs based on content length with maximums?
- **Scope**: Viewport scrolling, terminal size detection, responsive sizing
- **Deliverable**: Algorithm for height calculation (proportional to content, max 50%, min 15 lines)
- **Note**: Initial implementation uses fixed percentages - research adaptive sizing

**R5: Bookmark/Tag State Synchronization**
- **Question**: How to update TUI state after bookmark/tag changes in preview?
- **Scope**: Message passing in Bubbletea, state consistency, visual feedback
- **Deliverable**: Pattern for immediate UI updates without closing preview

**R6: Terminal Compatibility Testing**
- **Question**: Which terminals support required features (overlay dialogs, ANSI colors)?
- **Scope**: Test on macOS Terminal, iTerm2, Linux terminals, Windows Terminal
- **Deliverable**: Compatibility matrix and fallback strategies if needed

**R7: Pagination Display Best Practices**
- **Question**: Standard formats for numeric pagination in TUIs?
- **Scope**: Research common patterns (e.g., "3/5" vs "Page 3 of 5" vs "3 | 5")
- **Deliverable**: Recommended format and positioning

### Expected Research Outcomes

**Output**: `research.md` containing:
- Bubbletea search pattern with code examples
- Lipgloss width calculation best practices
- Help text rendering strategy
- Preview dialog sizing algorithm
- State synchronization patterns
- Terminal compatibility findings
- Pagination format recommendation

---

## Phase 1: Design & Contracts

### Data Model

**File**: `data-model.md`

**Key State Structures** (in `internal/tui/finder.go`):

```go
// Model state extensions needed for new features
type Model struct {
    // Existing state
    prompts          []index.SearchResult
    filteredPrompts  []index.SearchResult
    selectedSource   string
    selectedTags     map[string]bool
    bookmarkedOnly   bool

    // NEW: Search state
    searchMode       bool           // Is search field active?
    searchQuery      string         // Current search text
    searchInput      textinput.Model // Bubbles textinput component

    // NEW: Help text state
    helpText         string         // Dynamic help based on context

    // Existing preview state (modified)
    previewDialog    viewport.Model // Scrollable preview
    previewHeight    int           // Dynamic height (NEW calculation)

    // Existing tag display (modified)
    truncatedTags    map[string]string // Tag -> truncated display

    // Existing pagination (modified)
    currentPage      int
    totalPages       int
    paginationFormat string // NEW: "3/5" format
}
```

**State Transitions**:
1. **Normal → Search Mode**: User presses "/" key
2. **Search Mode → Normal**: User presses Esc or Enter
3. **Normal → Preview Mode**: User presses Enter on prompt
4. **Preview Mode → Tag Dialog**: User presses ctrl+t
5. **Preview Mode → Normal**: User presses Esc or removes bookmark

### API Contracts

**File**: `contracts/keyboard-shortcuts.md`

Keyboard shortcut definitions and state machine:

```markdown
# Keyboard Shortcuts Contract

## Main List View
- `/`: Activate search mode
- `↑/↓`: Navigate prompts
- `←/→`: Navigate pages (when multiple pages exist)
- `Enter`: Preview selected prompt
- `ctrl+t`: Open tag dialog for selected prompt
- `ctrl+b`: Toggle bookmark for selected prompt
- `Tab`: Switch between filters panel and prompts list
- `q`: Quit application

## Search Mode
- `Esc`: Exit search, restore filtered list
- `Enter`: Apply search
- (typing): Real-time filter

## Preview Mode
- `↑/↓` or `j/k`: Scroll preview content
- `ctrl+x`: Remove bookmark (NEW)
- `ctrl+b`: Toggle bookmark (NEW - add if not bookmarked)
- `ctrl+t`: Open tag dialog (NEW)
- `Esc`: Close preview

## Tag Dialog Mode
- Shows: "Current tags: tag1, tag2" or "No tags assigned"
- Input: Add tags (comma-separated)
- `Enter`: Save tags
- `Esc`: Cancel
```

**File**: `contracts/display-rules.md`

Visual display contracts:

```markdown
# Display Rules Contract

## Tag Truncation (FR-010 to FR-013)
- Maximum displayed length: 25 characters
- Truncation indicator: "..." (3 chars)
- Truncation point: 22 chars + "..."
- Example: "very-long-tag-name-th..." (25 chars total)
- Alignment: Checkbox alignment must not shift
- Full text access: Status line shows full tag name when selected

## Pagination Display (FR-014 to FR-017)
- Format: "N/M" where N=current page, M=total pages
- Example: "3/5"
- Position: Bottom right of prompts panel
- Single page: Show "1/1" (don't hide - keeps layout consistent)
- Update: Real-time on page change

## Preview Dialog Sizing (FR-018 to FR-022)
- Width: 80% of terminal width (unchanged)
- Height calculation:
  - Count content lines
  - If < 15 lines: Use content height
  - If > 50% terminal: Use 50% terminal
  - Else: Use content height
  - Absolute minimum: 15 lines
  - Absolute maximum: 50% of terminal height
- Scroll indicators: Visible when content exceeds viewport

## Help Text (FR-006 to FR-009)
- Location: Bottom line of TUI
- Context-aware content:
  - Filters panel active: Show filter navigation keys
  - Prompts list active: Show prompt navigation + actions
  - Search mode: Show search instructions
  - Preview mode: Show scroll + bookmark + tag shortcuts
- Format: Single line, concise descriptions
- Example: "↑/↓: navigate | ←/→: pages | /: search | Enter: preview | ctrl+t: tags | ctrl+b: bookmark | q: quit"
```

### Quickstart Guide

**File**: `quickstart.md`

Developer guide for modifying the TUI:

```markdown
# TUI Enhancement Quickstart

## Understanding the Codebase

**File to modify**: `internal/tui/finder.go` (~2000 lines)

**Architecture**: Bubbletea Elm Architecture
- `Init()`: Initialize model and components
- `Update(msg tea.Msg)`: Handle events, update state
- `View()`: Render current state to string

**Key components**:
- `list.Model`: Bubbles list component (prompts display)
- `textinput.Model`: Search input field
- `viewport.Model`: Scrollable preview dialog
- `overlay.Composite()`: Overlay library for dialogs

## Adding a New Feature

### Step 1: Extend Model State
Add new fields to `Model` struct (line ~50-100 in finder.go)

### Step 2: Handle Input in Update()
Add key binding cases in `Update()` method (line ~500-1000)

### Step 3: Update View Rendering
Modify `View()` method to show new state (line ~1200-1500)

### Step 4: Test Manually
```bash
go build -o bin/pkit ./cmd/pkit
./bin/pkit find
# Test your feature interactively
```

## Common Patterns

### Adding a Keyboard Shortcut
```go
// In Update() method
case key.Matches(msg, m.keys.YourNewKey):
    // Update state
    m.someState = true
    return m, someCommand()
```

### Showing a Dialog
```go
// In View() method
if m.showDialog {
    dialog := m.renderDialog()
    return overlay.Composite(dialog, baseView, overlay.Center, overlay.Center, 0, 0)
}
return baseView
```

### Dynamic Text with ANSI Codes
```go
// Use lipgloss.Width() for visual width
visualWidth := lipgloss.Width(text)
if visualWidth > maxWidth {
    // Truncate (preserving ANSI codes is tricky - use runewidth)
}
padding := maxWidth - visualWidth
paddedText := text + strings.Repeat(" ", padding)
```

### Real-time Filtering
```go
// Store original list, filter on each update
filtered := []Prompt{}
for _, p := range m.allPrompts {
    if matches(p, m.searchQuery) {
        filtered = append(filtered, p)
    }
}
m.filteredPrompts = filtered
m.list.SetItems(convertToListItems(filtered))
```

## Testing Workflow

1. Build binary: `make build`
2. Run TUI: `./bin/pkit find`
3. Test feature interactively
4. Check edge cases:
   - Empty results
   - Boundary values (exactly at limit)
   - State transitions (mode changes)
   - Terminal resize
5. Test on multiple terminals (macOS Terminal, iTerm2)

## Debugging Tips

- Add `fmt.Fprintf(os.Stderr, "DEBUG: %v\n", state)` (stderr doesn't corrupt TUI)
- Check ANSI rendering with `lipgloss.Width()` vs `len()`
- Test with different terminal sizes (resize window)
- Verify state cleanup on mode transitions
```

### Agent Context Update

After generating `research.md`, `data-model.md`, `contracts/`, and `quickstart.md`, the agent context will be updated:

**Command**:
```bash
.specify/scripts/bash/update-agent-context.sh claude
```

**Expected updates to `.claude/context.md`** (or similar agent-specific file):
- Add Bubbletea patterns discovered in research
- Add keyboard shortcut reference from contracts
- Add TUI modification patterns from quickstart
- Preserve existing project context

---

## Phase 2: Task Generation

**Note**: Phase 2 (task breakdown) is handled by the `/speckit.tasks` command, not `/speckit.plan`.

The tasks file (`tasks.md`) will be generated in the next phase based on this plan and will include:
- Granular implementation tasks for each of the 7 features
- Dependency ordering (e.g., search infrastructure before search UI)
- Test scenarios for each task
- Estimated complexity per task

---

## Implementation Notes

### Feature Priority for Implementation

Based on spec priorities and dependencies:

**P1 Features (implement first)**:
1. Search within filtered results (FR-001 to FR-005) - Core usability restoration
2. View and edit tags from tag dialog (FR-028 to FR-032) - Critical workflow gap

**P2 Features (implement second)**:
1. Clear navigation instructions (FR-006 to FR-009) - Quick win, improves discoverability
2. Tag display truncation (FR-010 to FR-013) - Prevents layout corruption
3. Optimized preview height (FR-018 to FR-022) - Better UX
4. Bookmark management from preview (FR-023 to FR-027) - Workflow efficiency

**P3 Features (implement last)**:
1. Numeric pagination display (FR-014 to FR-017) - Nice-to-have improvement

### Key Technical Decisions

**Search Implementation**:
- Use Bubbles `textinput.Model` for search field
- Filter `filteredPrompts` slice on each keystroke (already filtered by source/tags/bookmarks)
- Search across prompt ID, name, description fields
- "/" key activates search mode

**Tag Truncation**:
- Use `lipgloss.Width()` for accurate visual width measurement
- Truncate at 22 chars + "..." = 25 chars total
- Store full tag names in status line for accessibility

**Preview Height**:
- Calculate dynamically based on content line count
- Use min(content_lines, terminal_height * 0.5, 30) formula
- Ensure minimum of 15 lines for readability

**State Synchronization**:
- Bookmark/tag changes in preview immediately update underlying model
- No need to close preview - changes reflected in list when preview closes
- Use Bubbletea message passing for state updates

### Success Validation

After implementation, validate against success criteria:
- **SC-001**: Search response time < 1 second (measure with large prompt sets)
- **SC-002**: Navigation options identifiable in 5 seconds (user testing)
- **SC-003**: Tag layout consistent with 50-char tags (create test tags)
- **SC-004**: Page position clear within 2 seconds (user testing)
- **SC-005**: Preview max 50% height for 100+ line prompts (test with fabric:review_code)
- **SC-006**: Bookmark removal saves 3+ steps (compare old vs new workflow)
- **SC-007**: Duplicate tag reduction by 80% (measure after user testing)
- **SC-008**: Search works with all filter combinations (test matrix)

---

## Next Steps

1. ✅ Constitution Check: PASSED
2. 🔄 Phase 0: Execute research tasks → generate `research.md`
3. 🔄 Phase 1: Generate `data-model.md`, `contracts/`, `quickstart.md`
4. 🔄 Phase 1: Update agent context with TUI patterns
5. ⏸️ Phase 2: Run `/speckit.tasks` to generate `tasks.md` with implementation breakdown

**Command to continue**: `/speckit.tasks` (after Phase 0 and Phase 1 complete)
