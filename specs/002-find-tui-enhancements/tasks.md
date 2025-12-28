# Tasks: Enhance pkit Find TUI

**Input**: Design documents from `/specs/002-find-tui-enhancements/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

**Tests**: No automated tests requested in specification. Implementation uses manual testing workflow (interactive TUI validation).

**Organization**: Tasks grouped by user story to enable independent implementation and testing of each enhancement.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different functions/sections, no dependencies)
- **[Story]**: Which user story this task belongs to (US1-US7)
- Include exact file paths and line references in descriptions

## Path Conventions

**Project Type**: Single Go project
**Primary File**: `internal/tui/finder.go` (~2000 lines)
**All modifications** are localized to this single file

---

## Phase 1: Setup & Dependencies

**Purpose**: Verify environment and dependencies are ready

- [x] T001 Verify Go 1.25.4 and all dependencies (Bubbletea v1.3.10, Lipgloss v1.1.0, Bubbles v0.21.0, bubbletea-overlay v0.6.3) are installed
- [x] T002 Read and understand existing finder.go Model structure (lines 50-120 in internal/tui/finder.go)
- [x] T003 Review existing keyboard shortcuts and keybindings (keyMap struct in internal/tui/finder.go)

---

## Phase 2: Foundational Changes

**Purpose**: Core state extensions that multiple user stories depend on

**⚠️ CRITICAL**: Complete this phase before any user story implementation

- [x] T004 Add search state fields to Model struct in internal/tui/finder.go: searchMode bool, searchQuery string, searchInput textinput.Model, preSearchList []index.SearchResult
- [x] T005 [P] Add help text state field to Model struct in internal/tui/finder.go: helpText string
- [x] T006 [P] Add tag truncation fields to Model struct in internal/tui/finder.go: tagTruncateLength int (=25), truncatedTags map[string]string
- [x] T007 [P] Add pagination fields to Model struct in internal/tui/finder.go: currentPage int, totalPages int, pageSize int
- [x] T008 [P] Add preview sizing fields to Model struct in internal/tui/finder.go: previewMinHeight int (=15), previewMaxHeightPct float64 (=0.5), previewDynamicHeight int
- [x] T009 Initialize searchInput textinput component in Model creation function in internal/tui/finder.go with placeholder "Search prompts...", width 40
- [x] T010 Add "/" key binding to keyMap struct in internal/tui/finder.go: Search key.Binding with key "/" and help "search"

**Checkpoint**: Foundation ready - all state fields added, search infrastructure initialized

---

## Phase 3: User Story 1 - Search Within Filtered Results (Priority: P1) 🎯 MVP

**Goal**: Restore ability to perform full-text search across currently filtered prompts without losing active filters

**Independent Test**: Apply source filter "fabric" → press "/" → type "code" → verify only fabric prompts containing "code" shown → press Esc → verify full fabric list restored

### Implementation for User Story 1

- [x] T011 [US1] Implement "/" key handler in Update() method in internal/tui/finder.go to activate search mode: set searchMode=true, focus searchInput, backup filteredPrompts to preSearchList
- [x] T012 [US1] Implement search mode keyboard handlers in Update() method in internal/tui/finder.go: Esc cancels (restore preSearchList), Enter applies (keep current results)
- [x] T013 [US1] Implement applySearchFilter() function in internal/tui/finder.go that searches ID, name, description fields (case-insensitive) and returns filtered []index.SearchResult
- [x] T014 [US1] Implement real-time search filtering in Update() method in internal/tui/finder.go: on searchInput value change, call applySearchFilter() and update list component with results
- [x] T015 [US1] Implement renderSearchBar() function in internal/tui/finder.go to render search input with prompt "Search:" and current query text
- [x] T016 [US1] Modify View() method in internal/tui/finder.go to show search bar above prompts list when searchMode=true using lipgloss.JoinVertical()
- [x] T017 [US1] Update help text generation to show search mode help: "Type to search | Enter: apply | Esc: cancel" when searchMode=true

**Checkpoint**: Search functionality complete - users can search within filtered results, clear search, and apply search

**Manual Test**:
1. Run `go build && ./bin/pkit find`
2. Apply source filter: Select "fabric" source
3. Activate search: Press "/"
4. Type search query: "code"
5. Verify: Only fabric prompts containing "code" shown
6. Cancel: Press Esc
7. Verify: Full fabric list restored
8. Activate search again: Press "/"
9. Type: "review"
10. Apply: Press Enter
11. Verify: Search results persist after exiting search mode

---

## Phase 4: User Story 7 - View and Edit Tags from Tag Dialog (Priority: P1)

**Goal**: Show existing tags when user presses ctrl+t to enable modification and avoid duplicates

**Independent Test**: Press ctrl+t on prompt with tags ["code", "review"] → verify dialog shows "Current tags: code, review" → add "security" → verify all three tags saved

### Implementation for User Story 7

- [x] T018 [US7] Modify renderTagDialog() function in internal/tui/finder.go to fetch existing tags for currentPrompt from userTags map
- [x] T019 [US7] Add existing tags display section to tag dialog in renderTagDialog() in internal/tui/finder.go: show "Current tags: tag1, tag2" if tags exist, "No tags assigned" if empty
- [x] T020 [US7] Format existing tags display in renderTagDialog() with lipgloss faint style and proper text wrapping for long tag lists (using lipgloss.Width())
- [x] T021 [US7] Modify tag save handler in Update() method in internal/tui/finder.go to merge new tags with existing tags (deduplicate, trim whitespace, filter empty)
- [x] T022 [US7] Update tag dialog layout in renderTagDialog() to: title (bold) → existing tags (faint) → blank line → input label → input field → help text

**Checkpoint**: Tag dialog shows existing tags - users can see current tags and add new ones without duplicates

**Manual Test**:
1. Run `go build && ./bin/pkit find`
2. Select prompt with tags: Navigate to tagged prompt
3. Open tag dialog: Press ctrl+t
4. Verify: Shows "Current tags: existing, tags, here"
5. Add new tag: Type "newtag"
6. Save: Press Enter
7. Verify: All tags saved (existing + new)
8. Test empty prompt: Open tag dialog on untagged prompt
9. Verify: Shows "No tags assigned"

---

## Phase 5: User Story 2 - Clear Navigation Instructions (Priority: P2)

**Goal**: Provide context-aware help text showing available navigation options for current mode

**Independent Test**: Launch TUI → verify help shows navigation keys → press "/" → verify help changes to search instructions → preview prompt → verify help shows preview controls

### Implementation for User Story 2

- [x] T023 [P] [US2] Implement generateHelpText() function in internal/tui/finder.go with switch statement for different modes: Normal, Search, Preview, Tag Dialog, Filter Panel
- [x] T024 [P] [US2] Implement Normal mode help text generation in generateHelpText() checking if multiple pages exist: include "←/→: pages" only if totalPages > 1
- [x] T025 [US2] Implement Filter Panel help text in generateHelpText(): "↑/↓: navigate | Space: toggle | Tab: switch panel | /: search | q: quit"
- [x] T026 [US2] Implement Preview mode help text in generateHelpText(): "↑/↓: scroll | ctrl+b: bookmark | ctrl+t: tags | Esc: close"
- [x] T027 [US2] Implement Tag Dialog help text in generateHelpText(): "Enter: save | Esc: cancel"
- [x] T028 [US2] Create renderHelpBar() function in internal/tui/finder.go to render help text with faint lipgloss style at bottom of TUI
- [x] T029 [US2] Modify View() method in internal/tui/finder.go to call generateHelpText() and renderHelpBar() at bottom of all views using lipgloss.JoinVertical()

**Checkpoint**: Context-aware help text complete - help changes based on current mode and panel focus

**Manual Test**:
1. Run `go build && ./bin/pkit find`
2. Normal mode: Verify help shows "↑/↓: navigate | ... | q: quit"
3. Multiple pages: Filter to >20 prompts, verify "←/→: pages" shown
4. Single page: Filter to <20 prompts, verify "←/→: pages" hidden
5. Search mode: Press "/", verify help shows "Type to search | Enter: apply | Esc: cancel"
6. Filter panel: Press Tab, verify help shows filter panel shortcuts
7. Preview: Press Enter on prompt, verify help shows preview shortcuts

---

## Phase 6: User Story 3 - Tag Display Truncation (Priority: P2)

**Goal**: Truncate long tag names at 25 characters with "..." to prevent filter panel layout breaking

**Independent Test**: Create tag "very-long-tag-name-that-exceeds-display-limit" → verify displays as "very-long-tag-name-th..." → verify checkbox alignment consistent

### Implementation for User Story 3

- [x] T030 [US3] Implement truncateTag() function in internal/tui/finder.go using lipgloss.Width() and runewidth.RuneWidth() to truncate at 22 chars + "..."
- [x] T031 [US3] Add truncation logic in truncateTag() to handle: no truncation if ≤25 chars, character boundary truncation (not byte), CJK and emoji width support
- [x] T032 [US3] Implement cacheTruncatedTags() function in internal/tui/finder.go to populate truncatedTags map on initial load of available tags
- [x] T033 [US3] Call cacheTruncatedTags() during Model initialization in Init() or after loading tags from storage in internal/tui/finder.go
- [x] T034 [US3] Modify filter panel rendering in View() method in internal/tui/finder.go to use truncatedTags[tag] instead of full tag name for display
- [x] T035 [US3] Add status bar display for full tag name in Update() when tag is selected/highlighted in filter panel: "Tag: {full_tag_name}"

**Checkpoint**: Tag truncation complete - long tags display correctly without breaking layout

**Manual Test**:
1. Add very long tag via CLI: `./bin/pkit tag add fabric:summarize very-long-tag-name-that-exceeds-display-limit-test`
2. Run TUI: `./bin/pkit find`
3. View filters panel: Verify tag shows "very-long-tag-name-th..."
4. Check alignment: Verify checkbox at same position as short tags
5. Navigate to tag: Verify status bar shows full tag name
6. Test CJK: Add tag "日本語タグ名前長い長い長い長い" → verify truncation handles wide chars
7. Test exact limit: Add tag exactly 25 chars → verify no ellipsis

---

## Phase 7: User Story 5 - Optimized Preview Height (Priority: P2)

**Goal**: Calculate preview dialog height dynamically based on content (min 15 lines, max 50% terminal height)

**Independent Test**: Preview short prompt (10 lines) → verify small dialog → preview long prompt (150 lines) → verify dialog max 50% terminal with scroll → resize terminal → verify recalculation

### Implementation for User Story 5

- [x] T036 [US5] Implement calculatePreviewHeight() function in internal/tui/finder.go with parameters: contentLines int, terminalHeight int, returns int height
- [x] T037 [US5] Add height calculation logic in calculatePreviewHeight(): desired = contentLines + 6 (overhead), maxAllowed = terminalHeight * 0.5, apply min=15, max=maxAllowed
- [x] T038 [US5] Modify preview dialog creation in Update() when entering preview mode in internal/tui/finder.go: count content lines with strings.Count(content, "\n")
- [x] T039 [US5] Call calculatePreviewHeight() before creating viewport in preview mode and use returned height for viewport.New() in internal/tui/finder.go
- [x] T040 [US5] Add scroll indicators to preview rendering in renderPreviewDialog() in internal/tui/finder.go: show "▲ More above ▲" when viewport.ScrollPercent() > 0
- [x] T041 [US5] Add bottom scroll indicator in renderPreviewDialog(): show "▼ More below ▼" when viewport.ScrollPercent() < 1.0
- [x] T042 [US5] Handle terminal resize in Update() tea.WindowSizeMsg case: recalculate preview height if in preview mode and update viewport height

**Checkpoint**: Dynamic preview sizing complete - dialogs adapt to content and terminal size

**Manual Test**:
1. Run `go build && ./bin/pkit find`
2. Short prompt: Preview prompt <15 lines, verify dialog uses content height
3. Medium prompt: Preview ~30 line prompt, verify dialog proportional
4. Long prompt: Preview fabric:review_code (150+ lines), verify dialog max 50% terminal
5. Verify scroll indicators: Check ▲/▼ indicators when content extends beyond viewport
6. Resize terminal: Make terminal smaller, verify dialog recalculates to 50% max
7. Tiny terminal: Resize to 24 lines, verify min 15 lines enforced (may overflow)

---

## Phase 8: User Story 6 - Bookmark Management from Preview (Priority: P2)

**Goal**: Enable bookmark add/remove from preview dialog without closing preview

**Independent Test**: Preview bookmarked prompt → press ctrl+x → verify "✓ Bookmark removed" status → close preview → verify bookmark icon removed from list

### Implementation for User Story 6

- [x] T043 [P] [US6] Add ctrl+x key binding to keyMap struct in internal/tui/finder.go: RemoveBookmark key.Binding with key "ctrl+x" and help "remove bookmark"
- [x] T044 [US6] Implement toggleBookmarkInPreview() function in internal/tui/finder.go: check currentPrompt bookmark status, toggle via storage API, update bookmarks map, set status message
- [x] T045 [US6] Add ctrl+x key handler in Update() when inputMode==ModeViewingPrompt in internal/tui/finder.go: call toggleBookmarkInPreview(), return without closing preview
- [x] T046 [US6] Add ctrl+b key handler in preview mode in Update() in internal/tui/finder.go: call toggleBookmarkInPreview() (handles both add and remove)
- [x] T047 [US6] Implement bookmark status messages in toggleBookmarkInPreview(): "✓ Bookmarked" when adding, "✓ Bookmark removed" when removing, show for 2 seconds
- [x] T048 [US6] Update preview mode help text in generateHelpText() to include ctrl+x and ctrl+b shortcuts: "ctrl+x: remove bookmark | ctrl+b: toggle bookmark"
- [x] T049 [US6] Add bookmarkChanged flag to Model struct and set true after bookmark operations in preview, use to refresh list when preview closes in internal/tui/finder.go

**Checkpoint**: Bookmark management in preview complete - users can toggle bookmarks without closing preview

**Manual Test**:
1. Run `go build && ./bin/pkit find`
2. Preview unbookmarked: Press Enter on unbookmarked prompt
3. Add bookmark: Press ctrl+b
4. Verify: Status shows "✓ Bookmarked"
5. Stay in preview: Verify preview still open
6. Remove bookmark: Press ctrl+x
7. Verify: Status shows "✓ Bookmark removed"
8. Close preview: Press Esc
9. Verify: List shows prompt without bookmark icon
10. Test toggle: Preview bookmarked prompt, press ctrl+b, verify removed

---

## Phase 9: User Story 4 - Numeric Pagination Display (Priority: P3)

**Goal**: Replace dot pagination with numeric format "N/M" showing current page and total pages

**Independent Test**: Navigate to page 3 of 5 → verify displays "3/5" → filter to single page → verify "1/1" → navigate pages → verify real-time updates

### Implementation for User Story 4

- [x] T050 [US4] Implement updatePagination() function in internal/tui/finder.go: calculate totalPages = (len(filteredPrompts) + pageSize - 1) / pageSize, currentPage = list.Paginator.Page + 1
- [x] T051 [US4] Implement getPaginationText() function in internal/tui/finder.go: return formatted string "N/M" where N=currentPage, M=totalPages
- [x] T052 [US4] Call updatePagination() after filter changes in Update() method in internal/tui/finder.go: after applying source/tag/bookmark/search filters
- [x] T053 [US4] Call updatePagination() after page navigation in Update() when handling left/right arrow keys in internal/tui/finder.go
- [x] T054 [US4] Modify prompts panel border rendering in renderBorderedBox() in internal/tui/finder.go to include pagination text in bottom-right corner of border
- [x] T055 [US4] Style pagination text with lipgloss faint/dim color and right-align within border width in renderBorderedBox() using lipgloss.PlaceHorizontal()
- [x] T056 [US4] Disable list's built-in pagination dots in list initialization by setting list.Paginator type and hiding help in internal/tui/finder.go

**Checkpoint**: Numeric pagination complete - users see "N/M" format that updates in real-time

**Manual Test**:
1. Run `go build && ./bin/pkit find`
2. Multiple pages: Filter to show 50+ prompts (>1 page)
3. Verify: Bottom-right shows "1/3" (or similar)
4. Navigate: Press → key
5. Verify: Updates to "2/3"
6. Navigate: Press → again
7. Verify: Updates to "3/3"
8. Single page: Clear filters to show <20 prompts
9. Verify: Shows "1/1"
10. Verify positioning: Pagination in bottom-right, aligned properly

---

## Phase 10: Polish & Cross-Cutting Concerns

**Purpose**: Final refinements affecting multiple user stories

- [x] T057 [P] Add error handling for edge cases: zero search results message "No prompts match your search" in View() when filteredPrompts empty and searchMode true in internal/tui/finder.go
- [x] T058 [P] Handle edge case: removing last bookmark while bookmark filter active, show "No bookmarked prompts" message in View() in internal/tui/finder.go
- [x] T059 [P] Add status bar message cleanup in Update(): clear expired status messages when time.Now().After(statusExpiry) in internal/tui/finder.go
- [x] T060 [P] Optimize search performance: ensure applySearchFilter() runs <100ms for 1000 prompts (use simple string.Contains, no regex) in internal/tui/finder.go
- [x] T061 [P] Test and fix ANSI width calculations: verify all lipgloss.Width() calls correctly measure visual width for styled text in internal/tui/finder.go
- [x] T062 Manual cross-feature testing: test all 7 features together (search + filters + tag dialog + preview + bookmarks + pagination + help) for interaction issues
- [x] T063 Terminal compatibility testing: test on macOS Terminal, iTerm2, and verify graceful degradation with NO_COLOR environment variable
- [x] T064 Performance validation: test search response <100ms, filter update <50ms, verify no lag with 500+ prompts
- [x] T065 [P] Code cleanup: remove any debug fmt.Fprintf(os.Stderr) statements, ensure consistent naming, verify no unused variables in internal/tui/finder.go
- [x] T066 Documentation: add inline comments for complex functions (truncateTag, calculatePreviewHeight, applySearchFilter) in internal/tui/finder.go
- [x] T067 Build and smoke test: run `go build && go vet && go fmt` and verify binary works end-to-end with `./bin/pkit find`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup - BLOCKS all user stories
- **User Stories (Phases 3-9)**: All depend on Foundational completion
  - US1 (Search): Independent, can start after Foundational
  - US7 (Tag Dialog): Independent, can start after Foundational
  - US2 (Help Text): Independent, can start after Foundational
  - US3 (Tag Truncation): Independent, can start after Foundational
  - US5 (Preview Height): Independent, can start after Foundational
  - US6 (Bookmark in Preview): Independent, can start after Foundational
  - US4 (Pagination): Independent, can start after Foundational
- **Polish (Phase 10)**: Depends on all user stories being complete

### User Story Dependencies

All user stories are **INDEPENDENT** - they can be implemented in parallel or any order after Foundational phase:

- **US1 (Search)**: No dependencies on other stories
- **US7 (Tag Dialog)**: No dependencies on other stories
- **US2 (Help Text)**: No dependencies (generates help for all modes)
- **US3 (Tag Truncation)**: No dependencies on other stories
- **US5 (Preview Height)**: No dependencies on other stories
- **US6 (Bookmark in Preview)**: No dependencies on other stories
- **US4 (Pagination)**: No dependencies on other stories

### Within Each User Story

- Tasks within a story have minimal dependencies
- Tasks marked [P] can run in parallel (different functions/sections)
- Sequential tasks modify same functions in order

### Parallel Opportunities

- **Setup Phase**: T001, T002, T003 can all run in parallel (read-only)
- **Foundational Phase**: T005-T008 can run in parallel (adding different fields)
- **After Foundational**: ALL user stories (US1-US7) can be worked on in parallel by different developers
- **Within stories**: Tasks marked [P] are parallelizable
- **Polish Phase**: T057-T061, T063, T065, T066 can run in parallel

---

## Parallel Example: User Story 1 (Search)

```bash
# After Foundational Phase completes, launch US1 tasks in parallel:

# Parallel group 1 (different functions):
Task T013: "Implement applySearchFilter() function"
Task T015: "Implement renderSearchBar() function"

# Sequential after group 1:
Task T011: "Implement / key handler" (uses searchInput from T009)
Task T012: "Implement search mode keyboard handlers"
Task T014: "Implement real-time filtering" (uses applySearchFilter from T013)
Task T016: "Modify View() to show search bar" (uses renderSearchBar from T015)
Task T017: "Update help text for search mode"
```

---

## Parallel Example: Multiple User Stories

```bash
# After Foundational Phase completes, assign stories to different developers:

Developer A: Phase 3 (US1 - Search) - T011 through T017
Developer B: Phase 4 (US7 - Tag Dialog) - T018 through T022
Developer C: Phase 5 (US2 - Help Text) - T023 through T029

# All three developers work in parallel on independent features
# Stories integrate cleanly since they modify different functions
```

---

## Implementation Strategy

### MVP First (User Story 1 + User Story 7 Only)

Priority P1 stories provide maximum value:

1. Complete Phase 1: Setup (T001-T003)
2. Complete Phase 2: Foundational (T004-T010) - CRITICAL BLOCKER
3. Complete Phase 3: US1 - Search (T011-T017)
4. Complete Phase 4: US7 - Tag Dialog (T018-T022)
5. **STOP and VALIDATE**: Test both features independently and together
6. Build and demo MVP with search and tag management

### Incremental Delivery

1. Foundation: Setup + Foundational → Core state ready
2. MVP: Add US1 + US7 → Test independently → Demo (P1 features!)
3. Enhancement 1: Add US2 (Help Text) → Test → Demo
4. Enhancement 2: Add US3 (Tag Truncation) → Test → Demo
5. Enhancement 3: Add US5 (Preview Height) → Test → Demo
6. Enhancement 4: Add US6 (Bookmark Preview) → Test → Demo
7. Enhancement 5: Add US4 (Pagination) → Test → Demo
8. Polish: Phase 10 → Final validation → Release

Each story adds value without breaking previous stories.

### Parallel Team Strategy

With multiple developers working simultaneously:

**Week 1**:
- Team: Complete Setup + Foundational together (T001-T010)

**Week 2** (parallel work begins):
- Dev A: US1 (Search) - T011-T017
- Dev B: US7 (Tag Dialog) - T018-T022
- Dev C: US2 (Help Text) - T023-T029

**Week 3** (parallel work continues):
- Dev A: US3 (Tag Truncation) - T030-T035
- Dev B: US5 (Preview Height) - T036-T042
- Dev C: US6 (Bookmark Preview) - T043-T049

**Week 4**:
- Dev A: US4 (Pagination) - T050-T056
- Dev B & C: Polish (T057-T067)

---

## Notes

- **[P]** = Parallelizable tasks (different functions/sections, no dependencies)
- **[US#]** = User Story mapping for traceability
- **All tasks** modify `internal/tui/finder.go` - single file architecture
- **Manual testing** required for TUI features (interactive validation)
- **No automated tests** per specification (terminal UI complexity)
- **Performance targets**: Search <100ms, filter <50ms, instant display
- **Commit strategy**: Commit after completing each user story phase
- **Checkpoint validation**: Stop after each user story to test independently
- **Edge cases**: Explicitly handled in Polish phase (T057-T058)
- **Constitution aligned**: All tasks follow simplicity principle (single file, existing patterns, no new abstractions)

---

## Task Count Summary

- **Total Tasks**: 67
- **Setup**: 3 tasks
- **Foundational**: 7 tasks
- **US1 (Search)**: 7 tasks
- **US7 (Tag Dialog)**: 5 tasks
- **US2 (Help Text)**: 7 tasks
- **US3 (Tag Truncation)**: 6 tasks
- **US5 (Preview Height)**: 7 tasks
- **US6 (Bookmark Preview)**: 7 tasks
- **US4 (Pagination)**: 7 tasks
- **Polish**: 11 tasks

**Parallelization**: 20+ tasks marked [P] for parallel execution
**MVP Scope**: Phase 1-4 (22 tasks) delivers both P1 features
**Independent Stories**: All 7 user stories testable independently
