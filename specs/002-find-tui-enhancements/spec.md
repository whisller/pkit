# Feature Specification: Enhance pkit Find TUI

**Feature Branch**: `002-find-tui-enhancements`
**Created**: 2025-12-28
**Status**: Draft
**Input**: User description: "We will introduce few improvements for "pkit find" command.

1. We somehow lost possibility to full text search across filtered results. We need to get it back
2. We need to show to users that they can use arrows up/down (go through prompts) and left/right (go through pages)
3. We need to provide reasonable limit on tag size, to avoid situations like "[ ] bldfbakfdsahjkfdsahjkfdash jkdsahjkfdashjkfdhsajkfhdsajk │"
4. I believe more readable for pagination instead of dots would be numbers. Let's check this.
5. For "fabric:review_code" during preview of prompt it still is too high.
6. From preview of prompt we should be able to remove it from bookmarks
7. If prompt had tag and user pressed "ctrl + t" they should be able to see tags assigned for specified prompt, so they can modify it."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Search Within Filtered Results (Priority: P1)

Users need to perform full-text search across prompts that have already been filtered by source, tags, or bookmark status. This allows quick refinement of results without losing applied filters.

**Why this priority**: Core usability feature that was lost. Without search, users must manually scan through filtered lists, which defeats the purpose of having a powerful filter system.

**Independent Test**: Can be fully tested by applying any filter (source, tag, or bookmark), then typing text to search within those results. Delivers immediate value by allowing quick navigation through large filtered sets.

**Acceptance Scenarios**:

1. **Given** user has filtered prompts by source "fabric", **When** user types "/code" in search field, **Then** only fabric prompts containing "code" are shown
2. **Given** user has 50 filtered prompts displayed, **When** user searches for "review", **Then** list narrows to only matching prompts while maintaining filter panel visibility
3. **Given** user has searched within filtered results, **When** user clears search, **Then** full filtered list is restored
4. **Given** user has no filters applied, **When** user searches, **Then** search works across all prompts

---

### User Story 2 - Clear Navigation Instructions (Priority: P2)

Users need to understand available navigation options (arrow keys, page navigation) without trial and error. Clear help text guides efficient TUI usage.

**Why this priority**: Improves discoverability and reduces learning curve. While not blocking functionality, unclear navigation frustrates new users and reduces productivity.

**Independent Test**: Can be tested by launching find TUI and observing help text. Delivers value by reducing user confusion and support questions.

**Acceptance Scenarios**:

1. **Given** user launches find command, **When** viewing the help text, **Then** arrow key navigation (↑/↓ for prompts, ←/→ for pages) is clearly documented
2. **Given** user is in filters panel, **When** viewing help text, **Then** panel-specific navigation hints are shown
3. **Given** user is in prompts list, **When** viewing help text, **Then** list-specific navigation and actions are shown
4. **Given** list has multiple pages, **When** user views help, **Then** page navigation instructions are visible

---

### User Story 3 - Tag Display Truncation (Priority: P2)

Users with long tag names need readable filter lists without layout breaking. Tag names should be truncated with ellipsis when exceeding reasonable length.

**Why this priority**: Prevents TUI layout corruption and maintains professional appearance. Medium priority as it's a display issue, not a functional blocker.

**Independent Test**: Can be tested by creating tags with various lengths and verifying display. Delivers value by maintaining clean UI regardless of user data.

**Acceptance Scenarios**:

1. **Given** user has a tag named "very-long-tag-name-that-exceeds-display-limit", **When** viewing filters panel, **Then** tag displays as "very-long-tag-n..." with consistent checkbox alignment
2. **Given** user has multiple long tags, **When** viewing filters list, **Then** all tags maintain consistent width and alignment
3. **Given** user hovers over truncated tag, **When** tooltip/full text is available (if technically feasible), **Then** full tag name is shown
4. **Given** tag is shorter than truncation limit, **When** viewing filters, **Then** full tag name is displayed without ellipsis

---

### User Story 4 - Numeric Pagination Display (Priority: P3)

Users need clear indication of current page and total pages instead of abstract dot indicators. Numeric display (e.g., "Page 2/5") provides concrete navigation context.

**Why this priority**: Usability improvement but not critical. Dots work but numbers are more intuitive. Lowest priority as existing pagination functions correctly.

**Independent Test**: Can be tested by navigating through multi-page lists and verifying numeric display. Delivers value through improved orientation.

**Acceptance Scenarios**:

1. **Given** user has 100 prompts with 20 per page, **When** viewing page 3, **Then** display shows "3/5" or "Page 3 of 5"
2. **Given** user navigates between pages, **When** page changes, **Then** numeric indicator updates in real-time
3. **Given** user filters prompts reducing to single page, **When** viewing list, **Then** pagination indicator shows "1/1" or is hidden
4. **Given** user is on last page, **When** attempting to go forward, **Then** clear feedback indicates end of list

---

### User Story 5 - Optimized Preview Height (Priority: P2)

Users viewing long prompts like "fabric:review_code" need properly sized preview dialogs that don't overwhelm the screen while remaining scrollable for full content access.

**Why this priority**: Impacts user experience but current implementation is functional (just suboptimal). Medium priority as it affects daily usage comfort.

**Independent Test**: Can be tested by previewing various prompts including long ones. Delivers value through better screen real estate utilization.

**Acceptance Scenarios**:

1. **Given** user previews "fabric:review_code" prompt, **When** dialog opens, **Then** dialog height is proportional to content (not fixed 60%) with maximum of 50% screen height
2. **Given** prompt content is shorter than max height, **When** previewing, **Then** dialog sizes to fit content without excess whitespace
3. **Given** prompt content exceeds dialog height, **When** previewing, **Then** scroll indicators and instructions are clearly visible
4. **Given** user has small terminal window, **When** previewing, **Then** dialog maintains minimum readable size

---

### User Story 6 - Bookmark Management from Preview (Priority: P2)

Users viewing prompt preview need ability to remove bookmarks without closing preview dialog. Streamlines bookmark management workflow.

**Why this priority**: Convenience feature that reduces navigation steps. Medium priority as users can currently exit preview and use separate bookmark command.

**Independent Test**: Can be tested by previewing bookmarked prompt and using removal action. Delivers value through workflow efficiency.

**Acceptance Scenarios**:

1. **Given** user previews a bookmarked prompt, **When** preview dialog is open, **Then** bookmark removal option is clearly indicated in help text
2. **Given** user presses bookmark removal key in preview, **When** action completes, **Then** visual feedback confirms removal and prompt remains in list (but unbookmarked)
3. **Given** user removes bookmark from preview, **When** closing preview, **Then** prompt list reflects updated bookmark status
4. **Given** user previews non-bookmarked prompt, **When** viewing controls, **Then** bookmark addition option is available

---

### User Story 7 - View and Edit Tags from Tag Dialog (Priority: P1)

Users adding tags to a prompt need to see existing tags to avoid duplicates and enable modification. Tag dialog should display current tags when invoked.

**Why this priority**: Critical for tag management usability. Without seeing existing tags, users create duplicates or can't efficiently manage prompt tagging. Should have been part of original implementation.

**Independent Test**: Can be tested by pressing ctrl+t on tagged prompt and verifying existing tags are shown. Delivers immediate value for tag workflow.

**Acceptance Scenarios**:

1. **Given** prompt has tags ["code", "review"], **When** user presses ctrl+t, **Then** dialog shows "Existing tags: code, review" before input field
2. **Given** user views existing tags in dialog, **When** adding new tags, **Then** can see what already exists to avoid duplicates
3. **Given** prompt has no tags, **When** user presses ctrl+t, **Then** dialog shows "No existing tags" or similar clear message
4. **Given** prompt has many tags, **When** opening tag dialog, **Then** existing tags are displayed in readable format (wrapped if needed)

---

### Edge Cases

- What happens when search query matches zero filtered prompts?
- How does system handle tag truncation for tags that are exactly at the character limit?
- What happens when user tries to page navigation with only one page of results?
- How does preview height calculation handle terminal windows smaller than minimum size?
- What happens when user removes last bookmark while filtered by "bookmarked only"?
- How does tag display handle special characters or unicode in tag names?
- What happens when user searches for text that exists in metadata but not visible in list view?

## Requirements *(mandatory)*

### Functional Requirements

#### Search Functionality
- **FR-001**: System MUST provide real-time search input field that filters currently displayed prompts
- **FR-002**: Search MUST work within already filtered results (source, tags, bookmarks) without clearing those filters
- **FR-003**: Search MUST match against prompt ID, name, and description fields
- **FR-004**: Search MUST support clearing search text to restore full filtered list
- **FR-005**: Search field MUST be accessible via keyboard shortcut (e.g., "/" key)

#### Navigation Help
- **FR-006**: Help text MUST clearly indicate ↑/↓ keys navigate through prompt list
- **FR-007**: Help text MUST clearly indicate ←/→ keys navigate between pages (when multiple pages exist)
- **FR-008**: Help text MUST be context-aware showing relevant shortcuts for active panel (filters vs prompts)
- **FR-009**: Help text MUST be concise and fit within single line or designated help area

#### Tag Display
- **FR-010**: Tag names in filters panel MUST be truncated at 25 characters
- **FR-011**: Truncated tags MUST display ellipsis ("...") to indicate truncation
- **FR-012**: Tag checkboxes MUST maintain consistent alignment regardless of tag name length
- **FR-013**: Full tag name MUST be accessible (via full tag list or status message when selected)

#### Pagination Display
- **FR-014**: Pagination indicator MUST display numeric format showing current page and total pages
- **FR-015**: Pagination format MUST be "N/M" where N is current page and M is total pages
- **FR-016**: Pagination indicator MUST update in real-time as user navigates pages
- **FR-017**: When only one page exists, pagination indicator MUST show "1/1" or be hidden entirely

#### Preview Sizing
- **FR-018**: Preview dialog height MUST be dynamically calculated based on content length
- **FR-019**: Preview dialog height MUST NOT exceed 50% of terminal height
- **FR-020**: Preview dialog MUST maintain minimum height of 15 lines for readability
- **FR-021**: Preview dialog width MUST remain at 80% of terminal width (existing behavior)
- **FR-022**: Scroll indicators MUST clearly show when content extends beyond visible area

#### Bookmark Management in Preview
- **FR-023**: Preview dialog MUST include keyboard shortcut for bookmark removal (suggest ctrl+x or ctrl+d)
- **FR-024**: Bookmark removal action MUST provide visual confirmation (status message)
- **FR-025**: Bookmark removal MUST update prompt list bookmark indicators immediately
- **FR-026**: Preview dialog MUST also support adding bookmarks for non-bookmarked prompts
- **FR-027**: Help text in preview MUST clearly document bookmark management shortcuts

#### Tag Editing Dialog
- **FR-028**: Tag addition dialog (ctrl+t) MUST display existing tags before input field
- **FR-029**: Existing tags MUST be displayed as comma-separated list with clear label (e.g., "Current tags: code, review")
- **FR-030**: When prompt has no tags, dialog MUST clearly indicate this (e.g., "No tags assigned")
- **FR-031**: Tag dialog MUST support adding tags without removing existing ones
- **FR-032**: Tag dialog MUST wrap long existing tag lists for readability

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can search within filtered results and see matching prompts in under 1 second
- **SC-002**: Users can identify navigation options (arrow keys, page keys) within 5 seconds of viewing help text
- **SC-003**: Tag display maintains consistent layout with tags up to 50 characters long (truncated at 25)
- **SC-004**: Users can determine current page position within 2 seconds by reading numeric pagination
- **SC-005**: Preview dialog for long prompts (>100 lines) occupies no more than 50% of screen height
- **SC-006**: Users can remove bookmarks from preview without closing dialog, saving 3+ navigation steps
- **SC-007**: Users can view existing tags when adding new tags, reducing duplicate tag creation by 80%
- **SC-008**: Search functionality works across all filtering combinations (source + tags + bookmarks)
