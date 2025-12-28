# Tasks: Local Web Interface for pkit

**Input**: Design documents from `/specs/003-web-interface/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Tests are NOT explicitly requested in the feature specification, so no test tasks are included.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

Project uses existing pkit repository structure:
- `cmd/pkit/` - CLI entry point
- `internal/web/` - NEW web server package
- `internal/bookmark/`, `internal/tag/`, `internal/source/`, `internal/index/` - EXISTING packages (reuse)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic web server structure

- [ ] T001 Create internal/web package directory structure (server.go, handlers.go, middleware.go, embed.go)
- [ ] T002 [P] Create internal/web/templates directory with subdirectories (components/)
- [ ] T003 [P] Create internal/web/static directory for CSS and JavaScript assets

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core web server infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T004 Implement Server struct and NewServer constructor in internal/web/server.go
- [ ] T005 Implement Server.Start() and Server.Shutdown() methods in internal/web/server.go
- [ ] T006 [P] Implement logging middleware in internal/web/middleware.go
- [ ] T007 [P] Implement localhost-only enforcement middleware in internal/web/middleware.go
- [ ] T008 [P] Create base layout.html template in internal/web/templates/layout.html
- [ ] T009 Add 'serve' command to cmd/pkit/main.go with port flag
- [ ] T010 Configure embed directives in internal/web/embed.go for templates and static assets
- [ ] T011 Implement registerRoutes() method in internal/web/server.go (empty routes, to be filled)
- [ ] T012 Add graceful shutdown signal handling in cmd/pkit/main.go serve command
- [ ] T013 Implement health check handler at GET /health in internal/web/handlers.go

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Browse and Search Prompts via Web (Priority: P1) 🎯 MVP

**Goal**: Display all prompts from subscribed sources with search and filtering capabilities in a web browser

**Independent Test**: Start web server with `go run cmd/pkit/main.go serve`, open browser to http://127.0.0.1:8080, verify all prompts display with working search, source filter, tag filter, and bookmarked filter

### Implementation for User Story 1

- [ ] T014 [P] [US1] Create PromptListItem struct in internal/web/handlers.go
- [ ] T015 [P] [US1] Create FilterState struct in internal/web/handlers.go
- [ ] T016 [P] [US1] Create PaginatedResult struct in internal/web/handlers.go
- [ ] T017 [US1] Implement parseFilters() function in internal/web/handlers.go to extract query parameters
- [ ] T018 [US1] Implement applyFilters() function in internal/web/handlers.go for search and filtering logic
- [ ] T019 [US1] Implement paginate() function in internal/web/handlers.go (50 items per page)
- [ ] T020 [US1] Implement handleList() handler for GET / in internal/web/handlers.go
- [ ] T021 [US1] Register GET / route in registerRoutes() method in internal/web/server.go
- [ ] T022 [P] [US1] Create list.html template in internal/web/templates/list.html
- [ ] T023 [P] [US1] Create prompt-card.html component in internal/web/templates/components/prompt-card.html
- [ ] T024 [P] [US1] Create filters.html component in internal/web/templates/components/filters.html
- [ ] T025 [P] [US1] Create pagination.html component in internal/web/templates/components/pagination.html
- [ ] T026 [P] [US1] Create basic CSS styles in internal/web/static/style.css (layout, cards, filters, pagination)
- [ ] T027 [US1] Implement data cache loading at server startup in internal/web/server.go (prompts, bookmarks, tags)
- [ ] T028 [US1] Add sync.RWMutex for concurrent cache access in internal/web/server.go

**Checkpoint**: At this point, User Story 1 should be fully functional - users can browse, search, and filter prompts via web browser

---

## Phase 4: User Story 2 - View Prompt Details and Content (Priority: P1)

**Goal**: Display full prompt content when user clicks on a prompt, with metadata and navigation back to list

**Independent Test**: Click any prompt in the list, verify full content displays with source, tags, description, and a back button works

### Implementation for User Story 2

- [ ] T029 [P] [US2] Create PromptDetail struct in internal/web/handlers.go
- [ ] T030 [US2] Implement handleDetail() handler for GET /prompts/:id in internal/web/handlers.go
- [ ] T031 [US2] Register GET /prompts/ route in registerRoutes() method in internal/web/server.go
- [ ] T032 [P] [US2] Create detail.html template in internal/web/templates/detail.html
- [ ] T033 [P] [US2] Create prompt-header.html component in internal/web/templates/components/prompt-header.html
- [ ] T034 [P] [US2] Create prompt-content.html component in internal/web/templates/components/prompt-content.html
- [ ] T035 [P] [US2] Add detail page styles to internal/web/static/style.css
- [ ] T036 [US2] Implement 404 error page rendering in internal/web/handlers.go for non-existent prompts
- [ ] T037 [P] [US2] Create error.html template in internal/web/templates/error.html

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently - full browse and view workflow complete

---

## Phase 5: User Story 3 - Manage Bookmarks via Web (Priority: P2)

**Goal**: Allow users to bookmark and unbookmark prompts through the web interface with immediate visual feedback

**Independent Test**: Click bookmark button on a prompt detail page, verify bookmark indicator appears, filter by bookmarked prompts in list view

### Implementation for User Story 3

- [ ] T038 [US3] Implement handleBookmarkToggle() handler for POST /api/bookmarks in internal/web/handlers.go
- [ ] T039 [US3] Register POST /api/bookmarks route in registerRoutes() method in internal/web/server.go
- [ ] T040 [US3] Implement bookmark toggle logic using internal/bookmark package in internal/web/handlers.go
- [ ] T041 [US3] Implement cache reload after bookmark write in internal/web/handlers.go
- [ ] T042 [US3] Add bookmark button to prompt-header.html component in internal/web/templates/components/prompt-header.html
- [ ] T043 [P] [US3] Implement bookmark toggle JavaScript in internal/web/static/app.js
- [ ] T044 [P] [US3] Add bookmark button styles to internal/web/static/style.css
- [ ] T045 [US3] Add bookmark indicator ([*]) to prompt-card.html component for list view
- [ ] T046 [US3] Update handleList() to include bookmarked status in PromptListItem

**Checkpoint**: At this point, User Stories 1, 2, AND 3 should all work independently - bookmark management via web is functional

---

## Phase 6: User Story 4 - Manage Tags via Web (Priority: P2)

**Goal**: Allow users to add and remove tags from prompts through the web interface with data consistency

**Independent Test**: Click edit tags on a prompt, add/remove tags, verify changes persist and appear in both web UI and CLI after page refresh

### Implementation for User Story 4

- [ ] T047 [US4] Implement handleTagUpdate() handler for POST /api/tags in internal/web/handlers.go
- [ ] T048 [US4] Register POST /api/tags route in registerRoutes() method in internal/web/server.go
- [ ] T049 [US4] Implement tag update logic using internal/tag package in internal/web/handlers.go
- [ ] T050 [US4] Implement tag validation (lowercase, alphanumeric + hyphens) in internal/web/handlers.go
- [ ] T051 [US4] Implement cache reload after tag write in internal/web/handlers.go
- [ ] T052 [P] [US4] Create tag-editor.html component in internal/web/templates/components/tag-editor.html
- [ ] T053 [P] [US4] Implement tag editing JavaScript in internal/web/static/app.js
- [ ] T054 [P] [US4] Add tag editor styles to internal/web/static/style.css
- [ ] T055 [US4] Add tag editor to detail.html template in internal/web/templates/detail.html
- [ ] T056 [US4] Update prompt-card.html to display tags in list view

**Checkpoint**: All P1 and P2 user stories should now be independently functional - full organization features available

---

## Phase 7: User Story 5 - Copy Prompt Content (Priority: P3)

**Goal**: Provide one-click copy-to-clipboard functionality for prompt content with user feedback

**Independent Test**: Click copy button on prompt detail page, paste into another application, verify full prompt text appears

### Implementation for User Story 5

- [ ] T057 [US5] Add copy button to prompt-content.html component in internal/web/templates/components/prompt-content.html
- [ ] T058 [P] [US5] Implement clipboard copy JavaScript using navigator.clipboard API in internal/web/static/app.js
- [ ] T059 [P] [US5] Implement execCommand fallback for older browsers in internal/web/static/app.js
- [ ] T060 [P] [US5] Add copy confirmation feedback (toast/message) in internal/web/static/app.js
- [ ] T061 [P] [US5] Add copy button styles to internal/web/static/style.css

**Checkpoint**: All user stories should now be independently functional - complete web interface feature set delivered

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories and final validation

- [ ] T062 [P] Add loading states CSS for bookmark and tag operations in internal/web/static/style.css
- [ ] T063 [P] Add error message display CSS in internal/web/static/style.css
- [ ] T064 Implement renderTemplate() helper with error logging in internal/web/handlers.go
- [ ] T065 Add security headers middleware (X-Content-Type-Options, X-Frame-Options, CSP) in internal/web/middleware.go
- [ ] T066 [P] Add cache-control headers for static assets in internal/web/server.go
- [ ] T067 [P] Update README.md with web server usage documentation
- [ ] T068 Verify all functionality works as described in specs/003-web-interface/quickstart.md
- [ ] T069 Manual browser testing checklist: Chrome, Firefox, Safari, Edge
- [ ] T070 Verify data consistency between web interface and CLI (bookmark/tag in both directions)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-7)**: All depend on Foundational phase completion
  - US1 and US2 (both P1) can proceed in parallel after Phase 2
  - US3 and US4 (both P2) can proceed in parallel after Phase 2, but integrate with US2 (detail page)
  - US5 (P3) depends on US2 (adds copy button to detail page)
- **Polish (Phase 8)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 3 (P2)**: Can start after Foundational (Phase 2) - Adds to detail page from US2 but independently testable
- **User Story 4 (P2)**: Can start after Foundational (Phase 2) - Adds to detail page from US2 but independently testable
- **User Story 5 (P3)**: Can start after Foundational (Phase 2) - Adds to detail page from US2 but independently testable

### Within Each User Story

- Structs and data models before handler functions
- Handler functions before route registration
- Templates can be created in parallel with handlers
- CSS can be created in parallel with templates
- JavaScript enhancements after basic HTML functionality works
- Story complete before moving to next priority

### Parallel Opportunities

- **Phase 1**: All 3 setup tasks can run in parallel (different directories)
- **Phase 2**: Tasks T006, T007, T008 can run in parallel (different files)
- **Phase 3 (US1)**:
  - T014, T015, T016 (structs) can run in parallel
  - T022, T023, T024, T025, T026 (templates/CSS) can run in parallel
- **Phase 4 (US2)**:
  - T032, T033, T034, T035, T037 (templates/CSS) can run in parallel
- **Phase 5 (US3)**: T043, T044 (JS/CSS) can run in parallel
- **Phase 6 (US4)**: T052, T053, T054 (template/JS/CSS) can run in parallel
- **Phase 7 (US5)**: T058, T059, T060, T061 (all JS/CSS) can run in parallel
- **Phase 8**: T062, T063, T065, T066, T067 can run in parallel

### Cross-Story Parallelization

Once Phase 2 is complete, the following stories can be worked on simultaneously:
- Developer A: User Story 1 (Phase 3)
- Developer B: User Story 2 (Phase 4)
- Developer C: User Story 3 (Phase 5) - minimal waiting for US2 detail page structure

---

## Parallel Example: User Story 1

```bash
# Launch all templates/assets for User Story 1 together:
Task T022: "Create list.html template in internal/web/templates/list.html"
Task T023: "Create prompt-card.html component in internal/web/templates/components/prompt-card.html"
Task T024: "Create filters.html component in internal/web/templates/components/filters.html"
Task T025: "Create pagination.html component in internal/web/templates/components/pagination.html"
Task T026: "Create basic CSS styles in internal/web/static/style.css"
```

---

## Implementation Strategy

### MVP First (User Stories 1 & 2 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1 (Browse and Search)
4. Complete Phase 4: User Story 2 (View Details)
5. **STOP and VALIDATE**: Test US1 + US2 together (browse → view → back)
6. Deploy/demo if ready

**Result**: Users can browse, search, filter, and view prompts via web browser - core read-only functionality complete

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 + 2 → Test independently → Deploy/Demo (MVP - Browse & View)
3. Add User Story 3 → Test independently → Deploy/Demo (Add Bookmarking)
4. Add User Story 4 → Test independently → Deploy/Demo (Add Tag Management)
5. Add User Story 5 → Test independently → Deploy/Demo (Add Copy Feature)
6. Complete Polish → Final release

Each increment adds value without breaking previous functionality.

### Parallel Team Strategy

With multiple developers:

1. **Team completes Setup + Foundational together** (Phases 1-2)
2. **Once Foundational is done**:
   - Developer A: User Story 1 (Browse/Search)
   - Developer B: User Story 2 (View Details)
3. **After US1 + US2 complete** (MVP checkpoint):
   - Developer A: User Story 3 (Bookmarks)
   - Developer B: User Story 4 (Tags)
   - Developer C: User Story 5 (Copy)
4. **All developers**: Polish phase together

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Tests are NOT included as they were not explicitly requested in spec
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence
- All paths use existing pkit project structure (internal/web is new package)
- Reuse all existing internal packages (bookmark, tag, source, index) - DO NOT reimplement
