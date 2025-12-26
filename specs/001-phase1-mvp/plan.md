# Implementation Plan: pkit Phase 1 MVP - Pure Bookmark Manager

**Branch**: `001-phase1-mvp` | **Date**: 2025-12-25 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-phase1-mvp/spec.md`

## Summary

pkit Phase 1 MVP is a CLI-based bookmark manager for AI prompts that aggregates prompts from multiple GitHub repositories, provides unified search and indexing, and outputs clean text suitable for piping to execution tools (claude, llm, fabric, mods). This phase focuses purely on organization and discovery - NO execution, NO API calls, NO model management. The system must handle GitHub rate limits gracefully, sandbox parsers for security, support parallel source operations for performance, and maintain data integrity with fail-safe error handling.

**Technical Approach**: Single-binary CLI tool built with Cobra (command structure) + Bubbletea (interactive finder). Uses go-git for repository operations, file-based storage in ~/.pkit/, and secure token storage via system keychain. Emphasis on simplicity, Unix philosophy (stdin/stdout/stderr separation), and tool agnosticism. Follows idiomatic Go patterns and community standards throughout.

**UX Pattern**: Three-tier command resolution for maximum convenience:
1. **Explicit commands**: `pkit subscribe`, `pkit search`, `pkit find` → Execute command
2. **Shorthand get**: `pkit review` → Falls back to `pkit get review` (no need to type "get")
3. **Interactive finder**: `pkit find` → Real-time fuzzy search with selection

**Examples:**
```bash
pkit review | claude               # Shorthand: pkit get review | claude
pkit fabric:code-review | llm      # Shorthand: pkit get fabric:code-review | llm
pkit find --get | claude           # Interactive search + auto-get
```

## Technical Context

**Language/Version**: Go 1.23+ (latest stable as of Dec 2025)

**Primary Dependencies**:
- **Cobra** (github.com/spf13/cobra) - CLI command framework and routing
- **Bubbletea** (github.com/charmbracelet/bubbletea) - Interactive TUI for `pkit find` command
- **Bubbles** (github.com/charmbracelet/bubbles) - Bubbletea components (list, textinput, viewport)
- **Lipgloss** (github.com/charmbracelet/lipgloss) - TUI styling
- **go-git** (github.com/go-git/go-git/v5) - Git operations without system git dependency
- **goccy/go-yaml** (github.com/goccy/go-yaml) - Fast YAML parsing for bookmarks/config/frontmatter
- **zalando/go-keyring** (github.com/zalando/go-keyring) - Cross-platform secure token storage
- **bleve** (github.com/blevesearch/bleve/v2) - Full-text search indexing
- **isatty** (github.com/mattn/go-isatty) - TTY detection for interactive mode switching

**Storage**: File-based in ~/.pkit/ directory
- config.yml (user configuration - NO plain text tokens)
- bookmarks.yml (user bookmarks with aliases and tags)
- sources/ (cloned Git repositories)
- cache/ (bleve search index, rate limit tracking)

**Security**:
- GitHub tokens stored in system keychain via zalando/go-keyring
  - macOS: Keychain
  - Linux: Secret Service API / libsecret
  - Windows: Credential Manager
- Fallback to environment variable GITHUB_TOKEN if keychain unavailable
- NEVER store tokens in plain text config files

**Testing**:
- Go standard library testing (testing package)
- Table-driven tests for parsers and business logic
- Integration tests for CLI commands
- Contract tests for source format adapters
- TUI testing with test terminal buffers
- Use `testdata/` directory for test fixtures

**Target Platform**: Cross-platform CLI
- macOS (primary development target)
- Linux (server/developer usage)
- Windows (Windows 10+ with PowerShell/CMD support)

**Project Type**: Single binary CLI application with hybrid UX (traditional + interactive)

**Performance Goals**:
- Subscribe & index: <30 seconds for ~300 prompts
- Search: <1 second across all sources
- Interactive find: Real-time filtering (<50ms per keystroke)
- Get command: <100ms for prompt retrieval
- Shorthand resolution: <10ms to check if arg is bookmark/prompt
- Parallel source operations: ~O(slowest source) not O(sum of sources)

**Constraints**:
- Binary size: <20MB
- Memory usage: <50MB during typical operations
- No execution of prompts or LLM API calls
- No code execution from subscribed repositories (sandboxed parsers)

**Scale/Scope**:
- Support for 10-20 subscribed sources initially
- Handle repositories with 300-1000 prompts efficiently
- Cross-source search across all subscribed prompts
- Phase 1 supports 2-3 prompt formats (Fabric, awesome-chatgpt-prompts, generic Markdown)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Principle I: Organization-First Architecture ✅

**Status**: PASS

- Feature delivers organization-only functionality (subscribe, search/find, bookmark, get)
- NO execution of prompts or LLM API calls (FR-030, FR-031, FR-032)
- Complements execution tools via stdout piping (FR-025, FR-026)
- Shorthand syntax makes piping even easier: `pkit review | claude`
- Focuses on multi-source aggregation value proposition

### Principle II: CLI-First Interface ✅

**Status**: PASS

- Cobra-based CLI commands work with pipes and scripts
- Shorthand resolution: `pkit <prompt>` → `pkit get <prompt>` reduces typing
- Interactive `find` only activates in TTY, falls back to traditional output when piped
- Unix-friendly I/O: stdout for data, stderr for errors/logs (FR-027, FR-033)
- Supports piping to any execution tool (FR-026)
- JSON output available via --json flag (FR-029)
- Hybrid approach: traditional commands + interactive TUI where helpful

### Principle III: Tool Agnosticism ✅

**Status**: PASS

- Outputs clean prompt text to stdout for piping (FR-025)
- Shorthand works with any tool: `pkit review | claude`, `pkit review | llm`, etc.
- Works with claude, llm, fabric, mods, any tool (SC-004, SC-006)
- No LLM provider requirements in Phase 1 (FR-030, FR-031)
- Pluggable source format adapters (FR-003)

### Principle IV: Multi-Source Aggregation ✅

**Status**: PASS

- Subscribes to multiple GitHub repos (FR-001, FR-002)
- Cross-source search with unified indexing (FR-011, FR-012)
- Shorthand works across sources: `pkit fabric:review | claude`
- Interactive find searches across all sources in real-time
- Source namespace prefixes prevent conflicts (edge case handling)
- Tracks source versions and updates (FR-005, FR-006)
- Supports Fabric, awesome-chatgpt-prompts, generic Markdown (FR-003)

### Principle V: Simple Output Protocol ✅

**Status**: PASS

- `pkit get` outputs ONLY prompt text to stdout (FR-025)
- Shorthand `pkit <prompt>` behaves identically to `pkit get <prompt>`
- `pkit find` outputs selection to stdout (prompt ID or text with --get flag)
- Errors to stderr (FR-027)
- Exit codes: 0 for success, non-zero for errors (FR-028)
- --verbose and --debug flags for troubleshooting (FR-034, FR-035)
- Clean separation: stdout = data, stderr = logs/errors
- TTY detection ensures interactive mode doesn't break pipes

### Principle VI: Phase-Gated Development ✅

**Status**: PASS - Phase 1 Scope

- Delivers organization-only (subscribe, search, find, save, get/shorthand)
- NO execution wrappers in Phase 1 (FR-030)
- Shorthand syntax is UX improvement, not execution complexity
- Interactive TUI enhances UX but doesn't add execution complexity
- Validates core hypothesis: multi-source organization is useful
- Success gate: Core workflow functional (SC-005, SC-010)

### Principle VII: Simplicity & Focus ✅

**Status**: PASS

- Single binary CLI (no microservices, no complex architecture)
- File-based storage (no database for Phase 1)
- Standard Go patterns (Cobra for commands, Bubbletea for interactive only)
- Shorthand resolution is simple: check if arg matches bookmark/prompt, fall back to get
- Hybrid UX adds value without adding complexity (TTY detection is simple)
- No premature abstractions - implement as needed
- Focus on working software over clever architecture
- Rule of Three: No abstractions until pattern emerges multiple times

### Principle VIII: Go Best Practices ✅

**Status**: PASS - Implementation Standards

**Error Handling**:
- MUST handle errors explicitly - never ignore error returns
- MUST wrap errors with context: `fmt.Errorf("operation failed: %w", err)`
- MUST use sentinel errors for expected error conditions (e.g., `ErrNotFound`, `ErrInvalidAlias`)
- MUST return errors, not panic (except for programmer errors in init)
- Example: `if err := subscribeSource(url); err != nil { return fmt.Errorf("subscribe failed for %s: %w", url, err) }`

**Code Organization**:
- MUST follow standard Go project layout: `cmd/`, `internal/`, `pkg/`
- MUST use `internal/` for code that should not be imported externally (used in plan)
- MUST keep packages focused and cohesive - single responsibility
- MUST avoid circular dependencies between packages
- Structure already follows this: `cmd/pkit/`, `internal/config/`, `internal/source/`, etc.

**Interfaces & Abstractions**:
- MUST accept interfaces, return structs (except when interface is the contract)
- MUST keep interfaces small and focused (single method when possible)
- MUST define interfaces at point of use, not point of implementation
- MUST NOT create interfaces until multiple implementations exist or are clearly needed
- Example: `parser.Parser` interface only when we have multiple format parsers

**Concurrency**:
- MUST use goroutines for I/O-bound operations (git clones, file reads) - already planned for parallel source operations
- MUST use channels for goroutine communication, not shared memory
- MUST use `context.Context` for cancellation and timeouts
- MUST avoid premature concurrency - profile first
- Parallel source operations already leverage goroutines with proper synchronization

**Testing**:
- MUST write table-driven tests for multiple test cases
- MUST use `testdata/` directory for test fixtures (already in structure)
- MUST test exported API, not internal implementation
- MUST use subtests (`t.Run()`) for organizing test cases
- Example: Parser tests with table-driven approach for different input formats

**Code Style**:
- MUST run `go fmt` before commit - no exceptions
- MUST run `go vet` and address all warnings
- MUST use meaningful variable names - avoid single-letter except for short scopes
- MUST prefer early returns over nested conditionals
- MUST keep functions short and focused - extract when >50 lines
- Documented in quickstart.md code standards

**Dependencies**:
- MUST use Go modules for dependency management (go.mod already planned)
- MUST minimize external dependencies - standard library first
- Dependencies selected are well-maintained: Cobra (standard), Bubbletea (actively developed), go-git (stable)
- MUST review dependencies for maintenance status and security

**Performance**:
- MUST NOT optimize prematurely - measure first with benchmarks
- MUST use `strings.Builder` for string concatenation in loops
- MUST reuse buffers and reduce allocations in hot paths (search/index operations)
- MUST use `sync.Pool` for frequently allocated objects (only after profiling)
- Performance goals documented: <1s search, <100ms get

**Documentation**:
- MUST write godoc comments for all exported types, functions, constants
- MUST start comments with the name being documented
- MUST include examples for non-trivial API usage
- MUST document expected invariants and edge cases
- Example: Each parser implementation should have godoc explaining format support

**Implementation Enforcement**:
- Code reviews MUST verify Go best practices compliance
- CI pipeline MUST run: `go fmt`, `go vet`, `golangci-lint`
- PRs rejected if style violations exist
- Developer quickstart.md documents all code standards

**Overall Assessment**: ✅ **ALL PRINCIPLES SATISFIED** - No violations requiring justification.

## Project Structure

### Documentation (this feature)

```text
specs/001-phase1-mvp/
├── spec.md              # Feature specification
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (technology decisions, patterns)
├── data-model.md        # Phase 1 output (entities, relationships)
├── quickstart.md        # Phase 1 output (developer setup guide)
├── contracts/           # Phase 1 output (CLI command specifications)
│   ├── subscribe.md     # pkit subscribe command contract
│   ├── search.md        # pkit search command (traditional list)
│   ├── find.md          # pkit find command (interactive TUI)
│   ├── list.md          # pkit list command
│   ├── show.md          # pkit show command
│   ├── bookmark.md      # pkit save/alias/tag command contracts
│   ├── get.md           # pkit get command (+ shorthand behavior)
│   └── update.md        # pkit update/upgrade/status command contracts
└── checklists/          # Quality validation checklists
    └── requirements.md  # Spec quality checklist (already exists)
```

### Source Code (repository root)

```text
pkit/
├── cmd/
│   └── pkit/            # CLI entry point
│       ├── main.go      # Cobra root command setup + shorthand resolution
│       ├── subscribe.go # Subscribe command
│       ├── search.go    # Traditional search command
│       ├── find.go      # Interactive find command (bubbletea)
│       ├── list.go      # List command
│       ├── show.go      # Show command
│       ├── save.go      # Save bookmark command
│       ├── alias.go     # Alias command
│       ├── tag.go       # Tag command
│       ├── get.go       # Get command (used by shorthand too)
│       ├── update.go    # Update command
│       ├── upgrade.go   # Upgrade command
│       ├── status.go    # Status command
│       └── init.go      # Init command
│
├── internal/            # Private application code
│   ├── config/          # Configuration management
│   │   ├── config.go    # Load/save ~/.pkit/config.yml
│   │   ├── keyring.go   # Secure token storage (zalando/go-keyring)
│   │   └── ratelimit.go # GitHub rate limit tracking
│   │
│   ├── source/          # Source repository management
│   │   ├── manager.go   # Subscribe, clone, update sources
│   │   ├── git.go       # Git operations (go-git)
│   │   └── github.go    # GitHub API integration (rate limits)
│   │
│   ├── parser/          # Prompt format parsers
│   │   ├── parser.go    # Parser interface
│   │   ├── fabric.go    # Fabric pattern parser
│   │   ├── awesome.go   # awesome-chatgpt-prompts parser
│   │   └── markdown.go  # Generic Markdown parser
│   │
│   ├── index/           # Search indexing
│   │   ├── indexer.go   # Build and maintain bleve index
│   │   └── search.go    # Query bleve index
│   │
│   ├── bookmark/        # Bookmark management
│   │   ├── manager.go   # CRUD operations on bookmarks
│   │   ├── resolver.go  # Resolve shorthand args to prompts
│   │   └── validator.go # YAML validation, integrity checks
│   │
│   ├── tui/             # Terminal UI components (bubbletea)
│   │   ├── finder.go    # Interactive fuzzy finder model
│   │   ├── preview.go   # Prompt preview component
│   │   └── keybinds.go  # Keyboard shortcuts
│   │
│   └── display/         # Output formatting (traditional commands)
│       ├── table.go     # Table formatting for list/search
│       ├── text.go      # Plain text output for get
│       └── json.go      # JSON output for --json flag
│
├── pkg/                 # Public libraries (if needed for extensibility)
│   └── models/          # Shared data models
│       ├── source.go    # Source entity
│       ├── prompt.go    # Prompt entity
│       └── bookmark.go  # Bookmark entity
│
├── tests/
│   ├── integration/     # Integration tests (CLI command tests)
│   ├── contract/        # Contract tests for parsers
│   └── fixtures/        # Test data (sample repos, prompts)
│       └── testdata/    # Go standard testdata directory
│
├── .specify/            # Specify framework files
├── .github/             # GitHub workflows (CI/CD)
│   └── workflows/
│       ├── test.yml     # Run tests, go fmt, go vet
│       └── release.yml  # Build and release binaries
├── go.mod               # Go module definition
├── go.sum               # Dependency checksums
├── Makefile             # Build targets (build, test, lint, install)
└── README.md            # Project overview and quick start
```

**Structure Decision**: Single project structure selected following standard Go project layout. This is a CLI application with no frontend/backend split. The internal/ directory keeps implementation private (cannot be imported by other projects), while pkg/ provides public interfaces if needed for future extensibility. The `internal/tui/` package isolates Bubbletea interactive components from traditional CLI commands. The `internal/bookmark/resolver.go` handles shorthand argument resolution to check if an arg is a bookmark alias or prompt ID before falling back to get command. This follows Go best practices for project organization.

## Complexity Tracking

**No violations to justify** - all 8 Constitution principles satisfied without exceptions.

---

## Phase 0: Outline & Research

*Proceeding to generate research.md...*

### Research Topics

The following areas require research to resolve technical decisions and establish best practices:

1. **Git operations with go-git**
   - Decision needed: Authentication handling, clone strategies, update detection
   - Requirements: Clone repos, check for updates, handle GitHub auth without system git

2. **Bleve search indexing**
   - Decision needed: Index structure, query syntax, persistence strategy
   - Requirements: <1 second search, cross-source, persistent index, support for fuzzy matching

3. **GitHub API rate limiting strategy**
   - Decision needed: When to use API vs git clone, rate limit tracking/caching
   - Requirements: 60 req/h unauthenticated, 5000 req/h authenticated, graceful warnings

4. **Secure token storage with zalando/go-keyring**
   - Decision needed: Fallback strategies, cross-platform testing, error handling
   - Requirements: macOS Keychain, Windows Credential Manager, Linux Secret Service
   - Fallback: GITHUB_TOKEN environment variable

5. **Prompt parser security (sandboxing)**
   - Decision needed: Safe YAML/Markdown/CSV parsing without code execution
   - Requirements: No eval/exec, validate input safely, handle malicious YAML/frontmatter
   - Use goccy/go-yaml safely

6. **Bubbletea interactive finder UX**
   - Decision needed: Component selection (list vs textinput+filter), keybindings, preview pane
   - Requirements: Real-time filtering (<50ms), keyboard shortcuts, TTY detection for fallback

7. **Parallel operations with goroutines**
   - Decision needed: Worker pool pattern, error aggregation, progress reporting
   - Requirements: Parallel source ops, per-source progress, wait for all, context cancellation
   - Follow Go concurrency best practices

8. **Cross-platform binary distribution**
   - Decision needed: Build targets, release automation (goreleaser?), package managers
   - Requirements: macOS, Linux, Windows support, <20MB binary, Homebrew formula

9. **Shorthand command resolution pattern**
   - Decision needed: Cobra custom arg handling, performance of lookup
   - Requirements: <10ms resolution, clear error messages, preserve all explicit commands

---

*Continuing to Phase 0 research.md generation...*
