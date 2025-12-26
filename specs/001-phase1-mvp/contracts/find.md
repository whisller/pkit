# Command Contract: find

**Purpose:** Interactive TUI browser for prompts with real-time filtering and tagging

**Priority:** P1 (Primary discovery interface)

## Signature

```bash
pkit find [initial-query]
```

## Arguments

- `[initial-query]` (optional): Pre-populate search with initial query

## Flags

None (interactive mode handles all filtering)

## Behavior

1. **Launch**:
   - Check if running in TTY (terminal)
   - If not TTY (piped), fall back to traditional search
   - Launch Bubbletea TUI interface

2. **Interface** (Bubbletea):
   - **Search box**: Type to filter prompts in real-time
   - **Results list**: Navigate with arrow keys, page up/down
   - **Preview pane**: Show description of selected prompt
   - **Status indicators**: Show if prompt is bookmarked, tags
   - **Help footer**: Show keyboard shortcuts

3. **Real-time Filtering**:
   - Filter prompts as user types
   - Fuzzy matching on name, description, tags
   - Update results instantly (<100ms)

4. **Actions**:
   - **Enter**: Show full prompt details and copy ID to clipboard
   - **Ctrl+S**: Bookmark selected prompt (form: alias, tags, notes)
   - **Ctrl+T**: Edit tags on bookmarked prompt
   - **Ctrl+G**: Get prompt content (output to stdout after exit)
   - **Ctrl+B**: Toggle bookmarks-only filter
   - **Ctrl+R**: Clear search / Reset filters
   - **Ctrl+C / Esc**: Exit

## TUI Layout

### Main View

```
┌──────────────────────────────────────────────────────────────────────┐
│ Search: code review                                           [287]   │
├──────────────────────────────────────────────────────────────────────┤
│ ▸ fabric:code-review                                         [★]      │
│   fabric:code-reviewer-pro                                           │
│   fabric:security-review                                             │
│   awesome:code-helper                                                │
│   awesome:review-assistant                                           │
│   internal:go-review                                        [★]      │
│                                                                      │
├──────────────────────────────────────────────────────────────────────┤
│ Preview: fabric:code-review                              [★ review]  │
│                                                                      │
│ Expert code reviewer analyzing code for bugs, security issues, and   │
│ best practices. Provides detailed feedback with suggestions.         │
│                                                                      │
│ Tags: dev, security, go                                              │
│ Source: fabric                                                       │
├──────────────────────────────────────────────────────────────────────┤
│ ↑/↓: Nav • Enter: Details • ^S: Bookmark • ^T: Tags • ^G: Get • Esc │
└──────────────────────────────────────────────────────────────────────┘
```

**Legend:**
- `[★]` = Bookmarked prompt
- `[★ alias]` = Bookmarked with alias shown

### Bookmark Form (Ctrl+S on unbookmarked prompt)

```
┌──────────────────────────────────────────────────────────────────────┐
│ Bookmark: fabric:code-review                                         │
├──────────────────────────────────────────────────────────────────────┤
│                                                                      │
│ Alias (required): review_                                           │
│                                                                      │
│ Tags (comma-separated, optional):                                   │
│ dev,security,go_                                                    │
│                                                                      │
│ Notes (optional):                                                   │
│ Use for Go code reviews with security focus_                        │
│                                                                      │
│                                                                      │
│                                      [Tab: Next] [Enter: Save]       │
│                                                  [Esc: Cancel]       │
└──────────────────────────────────────────────────────────────────────┘
```

### Edit Tags Form (Ctrl+T on bookmarked prompt)

```
┌──────────────────────────────────────────────────────────────────────┐
│ Edit Tags: review (fabric:code-review)                               │
├──────────────────────────────────────────────────────────────────────┤
│                                                                      │
│ Current tags: dev, security, go                                      │
│                                                                      │
│ New tags (comma-separated):                                         │
│ dev,security,go,testing_                                            │
│                                                                      │
│ Tips:                                                               │
│  - Remove all tags: leave empty and save                            │
│  - Lowercase alphanumeric with hyphens/underscores only             │
│  - Press Tab for tag suggestions from other bookmarks               │
│                                                                      │
│                                      [Tab: Suggest] [Enter: Save]    │
│                                                     [Esc: Cancel]    │
└──────────────────────────────────────────────────────────────────────┘
```

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| Type | Filter prompts in real-time |
| ↑/↓ | Navigate results |
| PgUp/PgDn | Page through results |
| Enter | Show full details and copy ID |
| **Ctrl+S** | **Bookmark selected prompt** (form: alias, tags, notes) |
| **Ctrl+T** | **Edit tags** (only if prompt is bookmarked) |
| Ctrl+G | Get prompt content (output after exit) |
| Ctrl+B | Toggle bookmarks-only filter |
| Ctrl+R | Clear search / Reset filters |
| Esc / Ctrl+C | Exit |
| ? | Show help |

## Output

### Interactive Mode (TTY)

Launches full TUI interface (no direct output)

### After Bookmark (Ctrl+S pressed)

TUI shows success message, updates UI:

```
┌──────────────────────────────────────────────────────────────────────┐
│ ✓ Bookmarked as 'review'                                             │
│                                                                      │
│ Use it:                                                              │
│   pkit review | claude -p "analyse me ~/main.go"                    │
│   pkit get review | llm -m claude-3-sonnet                          │
│                                                                      │
│                                               [Press any key to close]│
└──────────────────────────────────────────────────────────────────────┘
```

Then returns to main view with `[★ review]` indicator on prompt.

### After Edit Tags (Ctrl+T pressed)

```
┌──────────────────────────────────────────────────────────────────────┐
│ ✓ Tags updated for 'review'                                          │
│                                                                      │
│ Old: dev, security, go                                               │
│ New: dev, security, go, testing                                      │
│                                                                      │
│                                               [Press any key to close]│
└──────────────────────────────────────────────────────────────────────┘
```

### After Selection (Enter pressed)

```
fabric:code-review [★ review]

Description:
Expert code reviewer analyzing code for bugs, security issues, and best practices.
Provides detailed feedback with suggestions for improvement.

Tags: dev, security, go, testing
Source: fabric
File: patterns/code-review/system.md
Bookmarked: Yes (alias: review, usage: 12 times)

Prompt ID copied to clipboard!

Actions:
  pkit get review | claude
  pkit review | claude              # shorthand
  pkit tag review dev,security      # edit tags from CLI
```

### After Get (Ctrl+G pressed)

Exits TUI, outputs prompt content to stdout:

```
# IDENTITY AND PURPOSE

You are an expert code reviewer...
```

### Non-TTY Fallback

```bash
$ pkit find "code review" | grep fabric
Warning: Not a TTY, falling back to search mode

fabric:code-review - Expert code reviewer analyzing code...
fabric:security-review - Security-focused code review...

Use 'pkit search "code review"' for non-interactive search.
```

## Examples

### Basic launch

```bash
$ pkit find
# Opens interactive TUI
```

### Launch with initial query

```bash
$ pkit find "summarize"
# Opens TUI with "summarize" pre-populated in search box
```

### Workflow: Find and bookmark with tags

```bash
$ pkit find
# 1. User types "code review"
# 2. Selects fabric:code-review with arrow keys
# 3. Presses Ctrl+S
# 4. Form appears:
#    - Alias: "review"
#    - Tags: "dev,security,go"
#    - Notes: "Use for Go code reviews"
# 5. Presses Enter to save
# 6. Success message shown
# 7. Returns to main view, prompt now shows [★ review]
```

### Workflow: Edit tags on bookmarked prompt

```bash
$ pkit find
# 1. User types "review"
# 2. Selects fabric:code-review [★ review]
# 3. Presses Ctrl+T
# 4. Form shows current tags: "dev,security,go"
# 5. User edits to: "dev,security,go,testing"
# 6. Presses Enter to save
# 7. Success message shown
# 8. Tags updated in bookmarks.yml
```

### Workflow: Find and use immediately

```bash
$ pkit find
# 1. User types "summarize"
# 2. Selects fabric:summarize
# 3. Presses Ctrl+G
# 4. TUI exits, prompt content output to stdout

$ pkit find | claude
# Same workflow, pipes directly to claude
```

### Workflow: Browse only bookmarks

```bash
$ pkit find
# 1. User presses Ctrl+B
# 2. Filter applied: shows only bookmarked prompts
# 3. User can still search within bookmarks
# 4. Press Ctrl+B again to show all prompts
```

## Error Handling

### Not a TTY (piped)

```bash
$ pkit find | grep summary
Warning: Not a TTY, falling back to traditional search

Showing results for all prompts...
fabric:summarize - Expert content summarizer...
awesome:summary-helper - Summarize text content...

For better experience, use:
  pkit search <query>  # Traditional search
  pkit find            # Interactive TUI (requires terminal)

Exit code: 0
```

### No prompts indexed

```bash
$ pkit find
Error: No prompts indexed yet

Get started:
  pkit subscribe fabric/patterns
  pkit subscribe f/awesome-chatgpt-prompts

Exit code: 1
```

### Invalid alias in bookmark form

TUI shows validation error inline:

```
┌──────────────────────────────────────────────────────────────────────┐
│ Bookmark: fabric:code-review                                         │
├──────────────────────────────────────────────────────────────────────┤
│                                                                      │
│ Alias (required): Review                                            │
│ ❌ Error: Alias must be lowercase alphanumeric with hyphens/underscores│
│                                                                      │
│ Tags (comma-separated, optional):                                   │
│ dev,security,go                                                     │
│                                                                      │
└──────────────────────────────────────────────────────────────────────┘
```

### Duplicate alias

```
┌──────────────────────────────────────────────────────────────────────┐
│ Bookmark: fabric:code-review                                         │
├──────────────────────────────────────────────────────────────────────┤
│                                                                      │
│ Alias (required): review                                            │
│ ⚠️  Warning: Alias 'review' already exists (fabric:improve-code)    │
│                                                                      │
│ Options:                                                            │
│  - Choose different alias                                           │
│  - Press Ctrl+F to force overwrite                                  │
│                                                                      │
└──────────────────────────────────────────────────────────────────────┘
```

### Ctrl+T on unbookmarked prompt

TUI shows message:

```
┌──────────────────────────────────────────────────────────────────────┐
│ ℹ️  Not Bookmarked                                                    │
│                                                                      │
│ This prompt is not bookmarked yet.                                  │
│                                                                      │
│ Press Ctrl+S to bookmark it first.                                  │
│                                                                      │
│                                               [Press any key to close]│
└──────────────────────────────────────────────────────────────────────┘
```

### Search returns no results

TUI shows:

```
┌──────────────────────────────────────────────────────────────────────┐
│ Search: xyz                                                     [0]   │
├──────────────────────────────────────────────────────────────────────┤
│                                                                      │
│ No prompts found matching "xyz"                                     │
│                                                                      │
│ Try:                                                                │
│  - Check spelling                                                   │
│  - Use fewer keywords                                               │
│  - Press Ctrl+R to clear search                                     │
│  - Press Ctrl+B if bookmark filter is active                        │
│                                                                      │
├──────────────────────────────────────────────────────────────────────┤
│ ↑/↓: Navigate • ^R: Clear • ^B: Toggle Bookmarks • Esc: Exit        │
└──────────────────────────────────────────────────────────────────────┘
```

## Performance

**Real-time Filtering Target**: <100ms per keystroke

**Strategy**:
- Use bleve index for fast full-text search
- Debounce search queries (50ms)
- Limit displayed results to 100
- Lazy load prompt details only for visible items
- Cache bookmark status lookups

## Form Validation

### Alias Validation

```go
func validateAlias(alias string) error {
    if alias == "" {
        return errors.New("alias is required")
    }

    if !validAliasRegex.MatchString(alias) {
        return errors.New("alias must be lowercase alphanumeric with hyphens/underscores")
    }

    // Check reserved commands
    for _, cmd := range reservedCommands {
        if alias == cmd {
            return fmt.Errorf("alias %q conflicts with built-in command", alias)
        }
    }

    return nil
}
```

### Tags Validation

```go
func validateTags(tagsStr string) ([]string, error) {
    if tagsStr == "" {
        return []string{}, nil // Empty tags OK
    }

    tags := strings.Split(tagsStr, ",")
    var cleaned []string
    seen := make(map[string]bool)

    for _, tag := range tags {
        tag = strings.TrimSpace(tag)
        tag = strings.ToLower(tag)

        if tag == "" || seen[tag] {
            continue // Skip empty or duplicate
        }

        if !regexp.MustCompile(`^[a-z0-9_-]+$`).MatchString(tag) {
            return nil, fmt.Errorf("invalid tag %q: must be lowercase alphanumeric with hyphens/underscores", tag)
        }

        cleaned = append(cleaned, tag)
        seen[tag] = true
    }

    return cleaned, nil
}
```

## Bubbletea Implementation

### Model

```go
type Model struct {
    // Search state
    searchInput   textinput.Model
    searchResults []Prompt
    cursor        int
    offset        int
    lastQuery     string

    // Filters
    bookmarksOnly bool
    sourceFilter  string

    // UI state
    width  int
    height int
    mode   ViewMode // main, bookmarkForm, editTagsForm, details, help

    // Forms
    bookmarkForm BookmarkForm
    editTagsForm EditTagsForm

    // Selected prompt
    selected *Prompt

    // Output for Ctrl+G
    getPrompt *Prompt

    // Bookmarks cache
    bookmarks map[string]Bookmark
}

type BookmarkForm struct {
    aliasInput textinput.Model
    tagsInput  textinput.Model
    notesInput textarea.Model
    focusIndex int
    error      string
}

type EditTagsForm struct {
    tagsInput textinput.Model
    bookmark  Bookmark
    error     string
}
```

### Update Loop

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch m.mode {
    case mainView:
        return m.updateMain(msg)
    case bookmarkFormView:
        return m.updateBookmarkForm(msg)
    case editTagsFormView:
        return m.updateEditTagsForm(msg)
    }
    return m, nil
}

func (m Model) updateMain(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.Type {
        case tea.KeyCtrlC, tea.KeyEsc:
            return m, tea.Quit

        case tea.KeyCtrlG:
            m.getPrompt = m.selected
            return m, tea.Quit

        case tea.KeyCtrlS:
            // Open bookmark form
            if m.selected != nil {
                if isBookmarked(m.selected.ID) {
                    // Already bookmarked, show message
                    return m.showMessage("Already bookmarked. Use Ctrl+T to edit tags."), nil
                }
                m.mode = bookmarkFormView
                m.bookmarkForm = newBookmarkForm(m.selected)
                return m, nil
            }

        case tea.KeyCtrlT:
            // Open edit tags form
            if m.selected != nil {
                bm, ok := m.bookmarks[m.selected.ID]
                if !ok {
                    return m.showMessage("Not bookmarked. Press Ctrl+S to bookmark first."), nil
                }
                m.mode = editTagsFormView
                m.editTagsForm = newEditTagsForm(bm)
                return m, nil
            }

        case tea.KeyCtrlB:
            m.bookmarksOnly = !m.bookmarksOnly
            m.searchResults = m.runSearch(m.searchInput.Value())
            m.cursor = 0

        case tea.KeyCtrlR:
            m.searchInput.SetValue("")
            m.bookmarksOnly = false
            m.searchResults = m.runSearch("")
            m.cursor = 0

        case tea.KeyEnter:
            if m.selected != nil {
                return m.showDetails(), nil
            }

        case tea.KeyUp:
            if m.cursor > 0 {
                m.cursor--
                m.updateSelected()
            }

        case tea.KeyDown:
            if m.cursor < len(m.searchResults)-1 {
                m.cursor++
                m.updateSelected()
            }
        }
    }

    // Update search input
    var cmd tea.Cmd
    m.searchInput, cmd = m.searchInput.Update(msg)

    // Re-run search if query changed
    if m.searchInput.Value() != m.lastQuery {
        m.searchResults = m.runSearch(m.searchInput.Value())
        m.lastQuery = m.searchInput.Value()
        m.cursor = 0
        m.updateSelected()
    }

    return m, cmd
}

func (m Model) updateBookmarkForm(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.Type {
        case tea.KeyEsc:
            m.mode = mainView
            return m, nil

        case tea.KeyEnter:
            // Validate and save
            alias := m.bookmarkForm.aliasInput.Value()
            tagsStr := m.bookmarkForm.tagsInput.Value()
            notes := m.bookmarkForm.notesInput.Value()

            // Validate alias
            if err := validateAlias(alias); err != nil {
                m.bookmarkForm.error = err.Error()
                return m, nil
            }

            // Check duplicate
            if _, exists := m.bookmarks[alias]; exists {
                m.bookmarkForm.error = fmt.Sprintf("Alias %q already exists. Press Ctrl+F to force overwrite.", alias)
                return m, nil
            }

            // Validate tags
            tags, err := validateTags(tagsStr)
            if err != nil {
                m.bookmarkForm.error = err.Error()
                return m, nil
            }

            // Save bookmark
            bookmark := Bookmark{
                Alias:      alias,
                PromptID:   m.selected.ID,
                SourceID:   m.selected.SourceID,
                PromptName: m.selected.Name,
                Tags:       tags,
                Notes:      notes,
                CreatedAt:  time.Now(),
                UpdatedAt:  time.Now(),
            }

            if err := SaveBookmark(bookmark); err != nil {
                m.bookmarkForm.error = err.Error()
                return m, nil
            }

            // Update cache
            m.bookmarks[m.selected.ID] = bookmark

            // Show success and return to main
            m.mode = mainView
            return m.showSuccessMessage(fmt.Sprintf("✓ Bookmarked as '%s'", alias)), nil

        case tea.KeyTab:
            // Cycle focus between inputs
            m.bookmarkForm.focusIndex = (m.bookmarkForm.focusIndex + 1) % 3
            return m, nil
        }
    }

    // Update focused input
    switch m.bookmarkForm.focusIndex {
    case 0:
        m.bookmarkForm.aliasInput, _ = m.bookmarkForm.aliasInput.Update(msg)
    case 1:
        m.bookmarkForm.tagsInput, _ = m.bookmarkForm.tagsInput.Update(msg)
    case 2:
        m.bookmarkForm.notesInput, _ = m.bookmarkForm.notesInput.Update(msg)
    }

    return m, nil
}

func (m Model) updateEditTagsForm(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.Type {
        case tea.KeyEsc:
            m.mode = mainView
            return m, nil

        case tea.KeyEnter:
            tagsStr := m.editTagsForm.tagsInput.Value()

            // Validate tags
            tags, err := validateTags(tagsStr)
            if err != nil {
                m.editTagsForm.error = err.Error()
                return m, nil
            }

            // Update bookmark
            bm := m.editTagsForm.bookmark
            oldTags := bm.Tags
            bm.Tags = tags
            bm.UpdatedAt = time.Now()

            if err := UpdateBookmark(bm); err != nil {
                m.editTagsForm.error = err.Error()
                return m, nil
            }

            // Update cache
            m.bookmarks[bm.PromptID] = bm

            // Show success
            m.mode = mainView
            return m.showSuccessMessage(fmt.Sprintf("✓ Tags updated: %s → %s",
                strings.Join(oldTags, ","), strings.Join(tags, ","))), nil
        }
    }

    m.editTagsForm.tagsInput, _ = m.editTagsForm.tagsInput.Update(msg)
    return m, nil
}
```

## Edge Cases

1. **Terminal too small**: Show warning, require minimum 80x24
2. **Rapid typing**: Debounce search to avoid lag
3. **Unicode in prompts**: Handle properly with runewidth
4. **Very long descriptions**: Wrap text properly in preview pane
5. **Clipboard not available**: Show prompt ID in output, skip clipboard
6. **No search results**: Show helpful message with suggestions
7. **Index rebuilding**: Show loading indicator, block interaction
8. **Form validation errors**: Show inline, keep form open
9. **Concurrent bookmark saves**: Last write wins (acceptable for Phase 1)
10. **Very long alias/tags**: Validate max lengths (alias: 50, tag: 30)

## Exit Codes

- `0`: Success (even if user cancels)
- `1`: Error (no prompts indexed, etc.)
- `2`: Not a TTY (fell back to search)

## Related Commands

- `pkit search <query>`: Traditional non-interactive search
- `pkit bookmarks`: Show only bookmarked prompts
- `pkit get <alias>`: Get prompt content for piping
- `pkit save <prompt-id> --as <alias>`: Bookmark via CLI
- `pkit tag <alias> <tags>`: Edit tags via CLI

## Requirements Mapping

- FR-011: Cross-source search functionality
- FR-012: Display results with source identifier, name, description
- FR-015: Detailed prompt view
- FR-016: Results clearly indicate source
- FR-017: Bookmark prompts with custom aliases
- FR-018: Tag bookmarks with multiple tags
- FR-021: Update tags on existing bookmarks
- SC-002: Search results in <1 second

## Dependencies

- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/lipgloss` - Styling
- `github.com/charmbracelet/bubbles/textinput` - Text input widget
- `github.com/charmbracelet/bubbles/textarea` - Multi-line text input
- `github.com/atotto/clipboard` - Clipboard support (optional)
- `golang.org/x/term` - TTY detection
- `github.com/mattn/go-isatty` - TTY detection

## Testing Checklist

- [ ] Launches in terminal
- [ ] Falls back to search when not TTY
- [ ] Real-time filtering works
- [ ] Keyboard navigation works
- [ ] Ctrl+S bookmarking form works
- [ ] Bookmark form validates alias
- [ ] Bookmark form validates tags
- [ ] Bookmark form handles duplicates
- [ ] Ctrl+T edit tags works
- [ ] Edit tags validates input
- [ ] Ctrl+G get works and outputs clean content
- [ ] Ctrl+B bookmark filter toggle works
- [ ] Enter shows details
- [ ] Esc/Ctrl+C exits cleanly
- [ ] Performance <100ms per keystroke
- [ ] Handles empty results
- [ ] Handles large result sets (1000+)
- [ ] Unicode renders correctly
- [ ] Works in various terminal sizes
- [ ] Form Tab navigation works
- [ ] Error messages display correctly
- [ ] Success messages display correctly
- [ ] Bookmark indicators [★] show correctly
