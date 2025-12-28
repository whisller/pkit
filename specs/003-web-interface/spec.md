# Feature Specification: Local Web Interface for pkit

**Feature Branch**: `003-web-interface`
**Created**: 2025-12-28
**Status**: Draft
**Input**: User description: "Apart of pkit find I would like to allow people to run local simple web page that allows to handle subscribed sources, same functionality as pkit find. It needs to be simple, lightweight, with as little dependencies as possible: simple html, template engine, maybe static page generator?"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Browse and Search Prompts via Web (Priority: P1)

As a user, I want to access my prompt library through a web browser running locally, so I can search and filter prompts without using the terminal.

**Why this priority**: This is the core value proposition - providing web-based access to the existing functionality. Without this, the feature has no purpose.

**Independent Test**: Can be fully tested by starting the local web server, opening a browser to the local address, and verifying that all prompts from subscribed sources are displayed with working search and filtering.

**Acceptance Scenarios**:

1. **Given** I have subscribed sources with prompts, **When** I run the local web server and open the browser, **Then** I see all my prompts displayed in a list with their names and descriptions
2. **Given** I am viewing the web interface, **When** I type text in the search box, **Then** the prompt list filters in real-time to show only matching prompts
3. **Given** I have prompts with tags, **When** I select tag filters, **Then** only prompts with those tags are displayed
4. **Given** I have multiple sources, **When** I select a source filter, **Then** only prompts from that source are displayed

---

### User Story 2 - View Prompt Details and Content (Priority: P1)

As a user, I want to click on a prompt to see its full content in the web interface, so I can review prompts before using them.

**Why this priority**: Viewing prompt content is essential - users need to see what they're selecting. This is part of the core browse-and-view workflow.

**Independent Test**: Can be tested by clicking any prompt in the list and verifying the full prompt content displays in a readable format.

**Acceptance Scenarios**:

1. **Given** I see a list of prompts, **When** I click on a prompt, **Then** the full prompt content is displayed in a readable format
2. **Given** I am viewing a prompt, **When** I want to return to the list, **Then** I can navigate back easily
3. **Given** a prompt has metadata (source, tags, description), **When** viewing the prompt, **Then** all metadata is clearly displayed

---

### User Story 3 - Manage Bookmarks via Web (Priority: P2)

As a user, I want to bookmark and unbookmark prompts through the web interface, so I can save favorites without switching to the terminal.

**Why this priority**: Bookmarking is a frequently used feature that significantly improves usability, but the core value (browsing/viewing) works without it.

**Independent Test**: Can be tested by clicking bookmark icons on prompts and verifying bookmarked prompts appear in the bookmarks filter.

**Acceptance Scenarios**:

1. **Given** I am viewing a prompt, **When** I click the bookmark button, **Then** the prompt is saved to my bookmarks and the button shows bookmarked state
2. **Given** a prompt is bookmarked, **When** I click the bookmark button again, **Then** the bookmark is removed
3. **Given** I have bookmarked prompts, **When** I select the "Bookmarked" filter, **Then** only my bookmarked prompts are displayed
4. **Given** I bookmark a prompt in the web interface, **When** I run `pkit find` in the terminal, **Then** the bookmark appears there as well (data consistency)

---

### User Story 4 - Manage Tags via Web (Priority: P2)

As a user, I want to add and remove tags from prompts through the web interface, so I can organize my prompt library without using the terminal.

**Why this priority**: Tag management improves organization but isn't required for basic browsing/viewing functionality.

**Independent Test**: Can be tested by adding/removing tags through the web UI and verifying changes are reflected in both the web interface and terminal.

**Acceptance Scenarios**:

1. **Given** I am viewing a prompt, **When** I click "Edit Tags", **Then** I can add new tags via a text input
2. **Given** a prompt has tags, **When** I edit tags, **Then** existing tags are pre-filled and I can modify or remove them
3. **Given** I update tags via the web interface, **When** I run `pkit find` in terminal, **Then** the tags match (data consistency)
4. **Given** I clear all tags from a prompt, **When** I save, **Then** the prompt has no tags

---

### User Story 5 - Copy Prompt Content (Priority: P3)

As a user, I want to copy a prompt's content to my clipboard with one click, so I can quickly paste it into my AI tool or document.

**Why this priority**: This is a convenience feature that improves workflow but isn't essential to core functionality.

**Independent Test**: Can be tested by clicking a copy button and pasting into another application.

**Acceptance Scenarios**:

1. **Given** I am viewing a prompt, **When** I click the "Copy" button, **Then** the prompt content is copied to my clipboard
2. **Given** prompt content is copied, **When** I paste in another application, **Then** the full prompt text appears
3. **Given** I copy a prompt, **When** the copy succeeds, **Then** I see a brief confirmation message

---

### Edge Cases

- What happens when no sources are subscribed? (Show empty state with instructions to subscribe via CLI)
- What happens when the web server port is already in use? (Show error message with port number and suggest alternatives)
- How does the system handle concurrent updates from both CLI and web interface? (Changes should be visible after page refresh; no real-time sync required)
- What happens when viewing a prompt with very long content (10,000+ characters)? (Content should be scrollable without breaking layout)
- What happens when search returns zero results? (Show "No results found" message)
- What happens when a user has 1000+ prompts? (Implement pagination to display 50 prompts per page)

## Requirements *(mandatory)*

### Functional Requirements

**Server & Core**

- **FR-001**: System MUST start a local HTTP server on a configurable port (default: 8080)
- **FR-002**: System MUST serve the web interface only on localhost/127.0.0.1 (no external network access)
- **FR-003**: System MUST use the same data sources as the CLI (read from same YAML files)
- **FR-004**: System MUST support graceful shutdown on SIGINT/SIGTERM
- **FR-005**: Server MUST log startup information (port, address) to console

**Prompt Display & Search**

- **FR-006**: Web interface MUST display all prompts from subscribed sources
- **FR-007**: Web interface MUST show prompt ID, name, description, source, and tags for each prompt
- **FR-008**: System MUST provide real-time search filtering by prompt name, ID, and description
- **FR-009**: System MUST provide filter options for: source, tags, and bookmarked status
- **FR-010**: System MUST display prompt count (e.g., "232 prompts")
- **FR-011**: System MUST implement pagination for large prompt lists (50 prompts per page)
- **FR-012**: System MUST show visual indicator ([*]) for bookmarked prompts in the list

**Prompt Details**

- **FR-013**: System MUST display full prompt content when a prompt is selected
- **FR-014**: Prompt detail view MUST show: source, tags, description, and full content
- **FR-015**: System MUST provide a way to navigate back from detail view to list view
- **FR-016**: System MUST preserve scroll position when returning to list view

**Bookmark Management**

- **FR-017**: System MUST allow users to bookmark prompts via a clickable button/icon
- **FR-018**: System MUST allow users to remove bookmarks via the same button/icon
- **FR-019**: System MUST persist bookmarks to the same storage used by CLI (YAML file)
- **FR-020**: System MUST show bookmark status immediately after bookmark/unbookmark action
- **FR-021**: Bookmark changes MUST be immediately visible in the bookmarked filter

**Tag Management**

- **FR-022**: System MUST provide an interface to add tags to prompts
- **FR-023**: Tag input MUST support comma-separated tag entry
- **FR-024**: System MUST pre-fill existing tags when editing tags
- **FR-025**: System MUST allow removing all tags (empty input saves as no tags)
- **FR-026**: System MUST persist tags to the same storage used by CLI
- **FR-027**: Tag changes MUST be immediately visible in tag filters

**Content Interaction**

- **FR-028**: System MUST provide a "Copy" button to copy prompt content to clipboard
- **FR-029**: System MUST show confirmation feedback when content is copied
- **FR-030**: Copy function MUST work in modern browsers (Chrome, Firefox, Safari, Edge)

**UI/UX Requirements**

- **FR-031**: Interface MUST be responsive and usable on desktop browsers (minimum 1024px width)
- **FR-032**: Interface MUST use minimal JavaScript for progressive enhancement
- **FR-033**: Interface MUST be fully functional without external CDN dependencies (all assets bundled)
- **FR-034**: Interface MUST show loading states during data operations
- **FR-035**: Interface MUST show error messages for failed operations (bookmark, tag edit, etc.)
- **FR-036**: Interface MUST maintain filter/search state during navigation within the page

### Key Entities

- **Prompt**: Represents a prompt from any source with ID, name, description, content, source ID, and tags
- **Source**: Represents a subscribed prompt source (e.g., fabric, awesome-chatgpt-prompts)
- **Bookmark**: Links a user to a saved prompt with timestamps and usage tracking
- **Tag Association**: Links prompts to user-defined tags for organization

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can start the web server and access the interface in under 30 seconds
- **SC-002**: Search filtering returns results in under 500ms for libraries with up to 1000 prompts
- **SC-003**: Interface loads and displays prompt list in under 2 seconds on first page load
- **SC-004**: Users can complete the flow "search → view prompt → bookmark → return to list" in under 15 seconds
- **SC-005**: Bookmark and tag operations complete and show feedback in under 1 second
- **SC-006**: Web interface consumes less than 100MB of memory with 1000+ prompts loaded
- **SC-007**: All functionality works in the 4 major browsers (Chrome, Firefox, Safari, Edge) without errors
- **SC-008**: Interface remains fully functional with JavaScript disabled for core viewing tasks
- **SC-009**: Server starts successfully on the specified port or returns a clear error message if port is in use
- **SC-010**: Data consistency: changes made in web interface are immediately visible in CLI after CLI restart

## Assumptions

1. **Minimal dependencies**: Will use Go's standard library `html/template` for server-side rendering to avoid external template engine dependencies
2. **Static assets**: CSS and minimal JavaScript will be embedded in the binary or served from a single static directory
3. **No real-time sync**: Changes between CLI and web interface will be visible after page refresh (no WebSocket/SSE required)
4. **Desktop-first**: Primary focus on desktop browsers; mobile support is not a priority for this version
5. **Single user**: Server is designed for single-user local use; no authentication or multi-user support needed
6. **Read-heavy**: Assumes most operations are read (browsing, searching) with occasional writes (bookmarks, tags)
7. **Port conflict handling**: If default port 8080 is in use, user will manually specify an alternative port via command flag

## Constraints

1. **Lightweight**: Must use minimal external dependencies to keep the binary small and compilation fast
2. **CLI compatibility**: Must use the same data structures and storage as the existing CLI to ensure compatibility
3. **Local only**: Must only bind to localhost to prevent accidental exposure to network
4. **Browser compatibility**: Must support evergreen browsers (last 2 versions) without polyfills
5. **No build toolchain**: Should avoid requiring Node.js, npm, webpack, or other JavaScript build tools

## Out of Scope

1. Real-time synchronization between multiple browser tabs or CLI sessions
2. Mobile-responsive design (desktop only for v1)
3. User authentication or multi-user support
4. Advanced text editing features (rich text, markdown preview)
5. Prompt creation or modification (viewing and organizing only)
6. Export functionality (PDF, markdown files)
7. Analytics or usage tracking
8. Dark mode or theme customization
9. Keyboard shortcuts beyond browser defaults
10. Offline mode or service worker caching
