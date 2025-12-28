# Implementation Plan: Local Web Interface for pkit

**Branch**: `003-web-interface` | **Date**: 2025-12-28 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/003-web-interface/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Implement a local web interface for pkit that provides the same functionality as `pkit find` (browse, search, filter, bookmark, tag prompts) through a browser-based UI. The interface must be lightweight with minimal dependencies, using Go's standard library `html/template` for server-side rendering and embedded static assets. The web server binds only to localhost, serves HTML/CSS/minimal JS, and uses the same data storage as the CLI for seamless integration.

## Technical Context

**Language/Version**: Go 1.25.4 (existing project version)
**Primary Dependencies**:
- Go standard library `html/template` (template rendering)
- Go standard library `net/http` (HTTP server)
- Go standard library `embed` (for embedding HTML/CSS/JS assets)
- Existing pkit internal packages (`internal/bookmark`, `internal/tag`, `internal/source`, `internal/index`)

**Storage**:
- YAML files in `~/.pkit/` (existing: `bookmarks.yml`, `tags.yml`, `config.yml`)
- Git clones in `~/.pkit/sources/` (existing)
- Bleve search index in `~/.pkit/cache/` (existing)

**Testing**: Go testing with `go test` (existing project pattern)
**Target Platform**: Cross-platform (macOS, Linux, Windows) - desktop browsers
**Project Type**: Single project (add web server to existing CLI)
**Performance Goals**:
- Server start: < 2 seconds
- Page load: < 2 seconds (first render)
- Search response: < 500ms for 1000+ prompts
- Memory: < 100MB with 1000+ prompts loaded

**Constraints**:
- Localhost-only binding (127.0.0.1)
- No external CDN dependencies (all assets embedded)
- Minimal JavaScript (progressive enhancement)
- Server-side rendering preferred over SPA
- Must use existing storage layer (no new data files)

**Scale/Scope**:
- Support 1000+ prompts in memory
- Single user (no authentication needed)
- Desktop browser support (Chrome, Firefox, Safari, Edge - last 2 versions)
- 50 prompts per page (pagination)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Core Principles Compliance

| Principle | Status | Justification |
|-----------|--------|---------------|
| **I. Organization-First Architecture** | ⚠️ **CONCERN** | Web UI is explicitly scoped to Phase 2 per constitution. **JUSTIFICATION**: User specifically requested web interface as alternative to TUI. This is planned Phase 2 work. Web UI will be view/organize only - NO execution, consistent with organization-first mandate. |
| **II. CLI-First Interface** | ✅ **PASS** | Web UI does NOT replace CLI. CLI remains primary interface. Web is browsing/organization complement, exactly as constitution describes for Phase 2. |
| **III. Tool Agnosticism** | ✅ **PASS** | Web UI does not add execution capabilities. Users still pipe CLI output to their chosen tools. Web is purely for browsing/organizing. |
| **IV. Multi-Source Aggregation** | ✅ **PASS** | Web UI displays all subscribed sources using existing index. No changes to multi-source aggregation logic. |
| **V. Simple Output Protocol** | ✅ **PASS** | Web UI does not change CLI output behavior. CLI remains text I/O. Web uses same data via read operations. |
| **VI. Phase-Gated Development** | ⚠️ **CONCERN** | Constitution places web UI in Phase 2 (months 4-6). **JUSTIFICATION**: This IS Phase 2 work. Spec explicitly states "MUST add local web UI for visual browsing/organizing" in Phase 2 description. Phase 1 CLI features are already complete. |
| **VII. Simplicity & Focus** | ✅ **PASS** | Spec requires minimal dependencies, no over-engineering, standard library usage. YAGNI enforced. |

### Gate Decision

**PROCEED** with following constraints:

1. **Phase Verification**: Confirm Phase 1 CLI features are complete and validated before implementing web UI
2. **No Execution**: Web UI MUST NOT add prompt execution - view/organize only
3. **Simplicity**: Use Go stdlib, embed assets, server-side rendering - no complex build toolchains
4. **CLI Primacy**: Document that web UI is complementary browser for existing CLI-managed data

**Constitution Alignment**: This feature aligns with Phase 2 goals: "MUST add local web UI for visual browsing/organizing". The web UI provides discovery and organization value without competing with execution tools or undermining CLI-first philosophy.

## Project Structure

### Documentation (this feature)

```text
specs/003-web-interface/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
│   ├── http-api.md      # HTTP endpoints for web UI
│   └── templates.md     # HTML template structure
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
cmd/
├── pkit/
│   └── main.go          # Add 'serve' subcommand

internal/
├── web/                 # NEW: Web server implementation
│   ├── server.go        # HTTP server setup
│   ├── handlers.go      # Request handlers (list, detail, search, bookmark, tag)
│   ├── middleware.go    # Logging, CORS, localhost enforcement
│   ├── templates/       # HTML templates (embedded)
│   │   ├── layout.html
│   │   ├── list.html
│   │   ├── detail.html
│   │   └── components/
│   ├── static/          # CSS/JS assets (embedded)
│   │   ├── style.css
│   │   └── app.js       # Minimal JS for copy, filters
│   └── embed.go         # Go embed directives
├── bookmark/            # EXISTING: Reuse for web handlers
├── tag/                 # EXISTING: Reuse for web handlers
├── source/              # EXISTING: Reuse for web handlers
├── index/               # EXISTING: Reuse for search
└── ... (other existing packages)

tests/
├── web/                 # NEW: Web server tests
│   ├── handlers_test.go
│   └── integration_test.go
└── ... (other existing tests)
```

**Structure Decision**: Single project structure (existing pkit codebase). Add new `internal/web` package for web server implementation. Reuse all existing data layer packages (`internal/bookmark`, `internal/tag`, `internal/source`, `internal/index`). No new storage, no new data models - pure UI layer on top of existing functionality.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Concern | Justification | Mitigation | Status |
|---------|---------------|------------|--------|
| **Phase 2 Timing** | Web UI belongs to Phase 2 (months 4-6) per constitution. Implementing now requires validation that Phase 1 is complete. | Verify Phase 1 features are shipped and validated: `subscribe`, `list`, `search`, `save`, `get` commands all functional with 2+ sources supported. Check git history confirms Phase 1 completion. | ✅ **ACCEPTED**: Phase 1 CLI complete based on recent commits (TUI enhancements, bookmark/tag management). |
| **Web UI Scope** | Must ensure web UI remains view/organize only, no execution. | Explicitly exclude execution from requirements. Web UI provides: browse, search, filter, bookmark, tag - NO "run" or "execute" functionality. | ✅ **ACCEPTED**: Spec explicitly states "Out of Scope: Prompt creation or modification (viewing and organizing only)". |

## Phase 0: Research & Technical Investigation

**Status**: PENDING

### Research Tasks

#### R001: Go HTTP Server Patterns for Single Binary Distribution
**Question**: What's the best practice for embedding and serving static assets (HTML/CSS/JS) in a Go binary for cross-platform distribution?

**Key Areas**:
- `embed` package usage patterns (Go 1.16+)
- File system abstraction for embedded vs. development mode
- Asset caching strategies
- Template compilation and reload during development

**Deliverable**: Recommended approach for asset embedding and serving

---

#### R002: Server-Side Rendering Architecture
**Question**: How should we structure HTML templates to minimize JavaScript while maintaining interactive filtering and search?

**Key Areas**:
- Template inheritance patterns in `html/template`
- Progressive enhancement strategies (works without JS, better with JS)
- Form submission vs. AJAX for filters/search
- URL query parameter handling for bookmarkable state

**Deliverable**: Template architecture and progressive enhancement strategy

---

#### R003: Localhost Security & Port Management
**Question**: Best practices for ensuring server only binds to localhost and handling port conflicts gracefully?

**Key Areas**:
- Binding to 127.0.0.1 vs. localhost (IPv6 considerations)
- Port conflict detection and user-friendly error messages
- Graceful shutdown on SIGINT/SIGTERM
- Browser auto-launch considerations

**Deliverable**: Server security configuration and port handling strategy

---

#### R004: State Management Without Database
**Question**: How to efficiently manage filter/search state and pagination using existing YAML storage?

**Key Areas**:
- In-memory caching of prompt data
- Incremental loading vs. full load at startup
- Search index integration (existing Bleve index)
- Bookmark/tag write operations without race conditions

**Deliverable**: Data loading and caching strategy

---

#### R005: Client-Side Clipboard API
**Question**: What's the most reliable cross-browser approach for copying prompt content to clipboard?

**Key Areas**:
- Modern Clipboard API (`navigator.clipboard.writeText()`)
- Fallback for older browsers (`document.execCommand('copy')`)
- Permission handling
- User feedback (success/failure messages)

**Deliverable**: Clipboard implementation pattern

**Output**: `research.md` with decisions, rationale, and alternatives for each research task

## Phase 1: Design & Contracts

**Status**: PENDING

**Prerequisites**: `research.md` complete

### D001: Data Model

**Deliverable**: `data-model.md`

**Content**:
- **No new entities** - reuse existing models from `pkg/models`:
  - `Prompt` (ID, Name, Description, Content, Source, Tags)
  - `Bookmark` (PromptID, Notes, CreatedAt, UpdatedAt, UsageCount, LastUsedAt)
  - `Source` (ID, URL, Type, Path, LastUpdate)
  - `Tag` (associations stored in `tags.yml`)

- **View models** (for rendering):
  - `PromptListItem`: Prompt + bookmarked status + tag list (for list view)
  - `PromptDetail`: Full prompt + metadata + bookmarked status
  - `FilterState`: Active source filters, tag filters, search query, bookmark filter, page number

- **No database migrations** - uses existing YAML storage
- **No new storage formats** - reads existing `bookmarks.yml`, `tags.yml`

---

### D002: HTTP API Contracts

**Deliverable**: `contracts/http-api.md`

**Endpoints**:

| Method | Path | Purpose | Request | Response |
|--------|------|---------|---------|----------|
| GET | `/` | List prompts (paginated, filtered) | Query params: `?search=`, `?source=`, `?tags=`, `?bookmarked=`, `?page=` | HTML page with prompt list |
| GET | `/prompts/:id` | View prompt detail | Path param: prompt ID | HTML page with full prompt |
| POST | `/api/bookmarks` | Toggle bookmark | JSON: `{prompt_id}` | JSON: `{bookmarked: bool}` |
| POST | `/api/tags` | Update tags | JSON: `{prompt_id, tags: []}` | JSON: `{success: bool}` |
| GET | `/api/search` | Real-time search (optional AJAX) | Query: `?q=` | JSON: `{results: [...]}` |
| GET | `/health` | Health check | None | `200 OK` |

**Design Principles**:
- Server-side rendering for core functionality (GET endpoints return HTML)
- JSON API for interactive features (bookmark, tags)
- Progressive enhancement: forms work with POST, enhanced with JS fetch
- RESTful where possible, pragmatic where needed

---

### D003: HTML Template Structure

**Deliverable**: `contracts/templates.md`

**Template Hierarchy**:
```
layout.html (base template)
├── list.html (prompt list page)
│   ├── components/prompt-card.html
│   ├── components/filters.html
│   └── components/pagination.html
└── detail.html (prompt detail page)
    ├── components/prompt-header.html
    ├── components/prompt-content.html
    └── components/tag-editor.html
```

**Template Data Contracts**:
- Each template receives specific data structure
- Partials for reusable components (filters, cards, pagination)
- CSS classes for styling (no inline styles)
- Minimal JS hooks (data attributes for interactive elements)

---

### D004: Development Quickstart

**Deliverable**: `quickstart.md`

**Content**:
- How to run web server in development: `go run cmd/pkit/main.go serve --port 8080`
- How to access: `http://localhost:8080`
- How to test with sample data
- How to modify templates (watch mode considerations)
- How to build and test embedded assets

---

### D005: Agent Context Update

**Action**: Run `.specify/scripts/bash/update-agent-context.sh claude`

**Updates**:
- Add Go web server patterns to context
- Add `html/template` usage
- Add embedded assets pattern
- Preserve existing CLI context

**Output**: Updated `.claude/context.md` or equivalent agent-specific file

## Implementation Phases (Post-Planning)

**Note**: These phases are executed by `/speckit.tasks` and NOT part of this planning document.

The following is provided for reference only:

### Phase 2: Core Implementation (via `/speckit.tasks`)
- Implement `internal/web` package
- Add `serve` command to CLI
- Create HTML templates
- Implement handlers (list, detail, search)

### Phase 3: Interactive Features (via `/speckit.tasks`)
- Implement bookmark toggle (API + JS)
- Implement tag editing (API + JS)
- Implement clipboard copy
- Add loading/error states

### Phase 4: Testing & Polish (via `/speckit.tasks`)
- Unit tests for handlers
- Integration tests for server
- Browser compatibility testing
- Performance optimization

## Appendix

### Related Files

- **Specification**: `/Users/danielancuta/Sites/whisller/pkit/specs/003-web-interface/spec.md`
- **Constitution**: `/Users/danielancuta/Sites/whisller/pkit/.specify/memory/constitution.md`
- **Existing TUI**: `/Users/danielancuta/Sites/whisller/pkit/internal/tui/finder.go` (reference for UI patterns)

### Dependencies

**Existing (reused)**:
- `github.com/whisller/pkit/internal/bookmark`
- `github.com/whisller/pkit/internal/tag`
- `github.com/whisller/pkit/internal/source`
- `github.com/whisller/pkit/internal/index`
- `github.com/whisller/pkit/pkg/models`

**New (standard library)**:
- `net/http`
- `html/template`
- `embed`

**No external dependencies added** - aligns with "lightweight, minimal dependencies" requirement.
