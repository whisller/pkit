# Feature Specification: pkit Phase 1 MVP - Pure Bookmark Manager

**Feature Branch**: `001-phase1-mvp`
**Created**: 2025-12-25
**Status**: Draft
**Input**: User description: "Phase 1 MVP - Pure bookmark manager for AI prompts with multi-source aggregation"

## Clarifications

### Session 2025-12-25

- Q: When a user's bookmark file (~/.pkit/bookmarks.yml) becomes corrupted or contains invalid YAML, how should pkit handle this at startup? → A: Fail safely: Refuse to start, display clear error message with file path, suggest manual fix or backup restoration, offer `pkit init --force` to reset
- Q: When subscribing to GitHub repositories, how should pkit handle GitHub API rate limits (60 requests/hour unauthenticated, 5000/hour authenticated)? → A: Graceful degradation: Allow unauthenticated by default, warn when approaching rate limit, provide clear instructions to add token for higher limits
- Q: What level of logging and observability should pkit provide for troubleshooting and debugging? → A: Standard CLI: Silent by default, `--verbose` flag for detailed operation logs (git operations, file I/O, parsing), `--debug` flag for full trace including timing
- Q: How should pkit handle potentially malicious or compromised GitHub repositories? → A: Basic validation: Sandbox prompt parsers (no code execution), warn users if repo contains executable scripts or hooks, skip non-standard files
- Q: Should pkit support concurrent operations when subscribing to multiple sources or performing updates? → A: Parallel source operations: Allow concurrent subscribe/update of different sources, show progress for each, wait for all to complete

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Subscribe and Discover Prompts (Priority: P1)

A developer wants to access prompts from Fabric and awesome-chatgpt-prompts repositories. They subscribe to these sources and can immediately search and view prompts across both sources from a single interface.

**Why this priority**: This is the foundation of pkit's value proposition - aggregating multiple prompt sources. Without this, there's no product.

**Independent Test**: User can run `pkit subscribe fabric/patterns`, then `pkit search ""` and see Fabric prompts. They can search across subscribed sources using `pkit search "code review"` and view details with `pkit show`.

**Acceptance Scenarios**:

1. **Given** user has pkit installed, **When** they run `pkit subscribe fabric/patterns`, **Then** pkit clones the repository to ~/.pkit/sources/ and indexes all prompts
2. **Given** user has subscribed to 2+ sources, **When** they run `pkit search "summarize"`, **Then** they see matching prompts from all subscribed sources with source identifiers (e.g., "fabric:summarize")
3. **Given** user has subscribed sources, **When** they run `pkit search ""`, **Then** they see a table listing all available prompts with source, name, and brief description
4. **Given** user found a prompt via search, **When** they run `pkit show fabric:code-review`, **Then** they see the full prompt text, description, source, and metadata
5. **Given** user subscribes to an invalid or unreachable repository, **When** subscription fails, **Then** they see a clear error message explaining the issue

---

### User Story 2 - Bookmark and Organize Prompts (Priority: P2)

A developer finds useful prompts across multiple sources and wants to save them with memorable aliases and tags for quick access. They can bookmark prompts, rename them, add tags, and filter by tags.

**Why this priority**: Bookmarking is the personal workspace layer that makes pkit useful beyond basic search. Users need this to build their own curated collection.

**Independent Test**: User can run `pkit save fabric:code-review --as review --tags dev,security`, then retrieve it with `pkit list --tags security` and see their bookmark.

**Acceptance Scenarios**:

1. **Given** user has found a prompt they like, **When** they run `pkit save fabric:code-review --as review --tags dev,security`, **Then** the bookmark is saved to ~/.pkit/bookmarks.yml with the alias and tags
2. **Given** user has bookmarked prompts, **When** they run `pkit list --tags security`, **Then** they see only bookmarks tagged with "security"
3. **Given** user wants to rename a bookmark, **When** they run `pkit alias review code-review`, **Then** the bookmark is renamed and accessible by the new alias
4. **Given** user wants to update tags, **When** they run `pkit tag review dev,security,go`, **Then** the bookmark's tags are updated
5. **Given** user tries to save with an existing alias, **When** they run `pkit save`, **Then** they get a warning and option to overwrite or choose a different alias

---

### User Story 3 - Pipe Prompts to Execution Tools (Priority: P1)

A developer wants to use their bookmarked prompts with their preferred execution tool (claude, llm, fabric, mods). They retrieve prompt text via stdout and pipe it to any tool.

**Why this priority**: This validates pkit's tool-agnostic philosophy and integration with the existing ecosystem. Without this, users can't actually use the prompts.

**Independent Test**: User can run `pkit get review | claude -p "analyse me ~/file.go"` and the prompt text flows correctly to claude for execution.

**Acceptance Scenarios**:

1. **Given** user has a bookmarked prompt "review", **When** they run `pkit get review`, **Then** ONLY the prompt text is output to stdout (no headers, formatting, or metadata)
2. **Given** user wants to pipe to claude, **When** they run `pkit get review | claude -p "analyse me ~/file.go"`, **Then** the prompt text is received by claude and execution succeeds
3. **Given** user wants to pipe to llm, **When** they run `cat article.txt | pkit get summarize | llm -m claude-3-sonnet`, **Then** the workflow executes correctly
4. **Given** user requests a non-existent bookmark, **When** they run `pkit get nonexistent`, **Then** an error is written to stderr and exit code is non-zero
5. **Given** user wants JSON output, **When** they run `pkit get review --json`, **Then** prompt metadata and text are output as valid JSON

---

### User Story 4 - Track Source Updates (Priority: P3)

A developer has subscribed to prompt sources that are actively maintained. They want to know when updates are available and selectively upgrade sources to get new or improved prompts.

**Why this priority**: Version tracking adds value but isn't critical for MVP validation. Users can manually re-subscribe if needed in early versions.

**Independent Test**: User runs `pkit update` after upstream changes and sees which sources have updates. They can run `pkit upgrade fabric` to update that specific source.

**Acceptance Scenarios**:

1. **Given** user has subscribed sources, **When** they run `pkit update`, **Then** pkit checks each source's git repository for updates and shows outdated sources
2. **Given** updates are available for a source, **When** they run `pkit status`, **Then** they see which sources are outdated with commit counts/dates
3. **Given** user wants to update a specific source, **When** they run `pkit upgrade fabric`, **Then** that source is pulled to the latest version and re-indexed
4. **Given** user wants to update all sources, **When** they run `pkit upgrade --all`, **Then** all outdated sources are updated
5. **Given** a source update includes new prompts, **When** user upgrades, **Then** new prompts appear in search and list results

---

### Edge Cases

- What happens when a subscribed repository is deleted or becomes inaccessible? System should handle gracefully with error message and option to unsubscribe.
- How does the system handle prompt name conflicts across different sources? Use source namespace prefix (e.g., "fabric:summarize" vs "awesome:summarize").
- What happens when the ~/.pkit/ directory is corrupted or missing? System should detect and offer to reinitialize with warnings about data loss.
- What happens when bookmarks.yml becomes corrupted or contains invalid YAML? System MUST refuse to start, display clear error message with file path, suggest manual fix or backup restoration, and offer `pkit init --force` to reset to empty state.
- How does the system handle GitHub API rate limits? System MUST allow unauthenticated access by default (60 requests/hour), warn users when approaching rate limit threshold (e.g., 80% consumed), and provide clear instructions on how to configure GitHub token for higher limits (5000 requests/hour). Rate limit status should be trackable via `pkit status --verbose`.
- How does the system handle potentially malicious repositories? System MUST sandbox prompt parsers to prevent code execution, warn users if repository contains executable scripts or git hooks, and skip non-standard files during indexing. Only parse recognized prompt formats (Markdown, CSV, YAML).
- How does the system handle very large repositories (1000+ prompts)? Indexing should be optimized with progress indicators and caching.
- What happens when a user bookmarks a prompt and then the source is updated removing that prompt? The bookmark should remain but show a warning that the source prompt no longer exists.
- How does the system handle network connectivity issues during subscribe/update? Clear error messages with retry suggestions.
- What happens when parsing a prompt format fails? Skip that prompt with a warning, don't fail entire indexing.

## Requirements *(mandatory)*

### Functional Requirements

**Source Management:**

- **FR-001**: System MUST support subscribing to GitHub repositories via short syntax (e.g., "fabric/patterns") and full URLs (e.g., "https://github.com/company/internal-prompts")
- **FR-002**: System MUST clone subscribed repositories to ~/.pkit/sources/ with organized directory structure
- **FR-003**: System MUST support at minimum Fabric pattern format (Markdown with frontmatter) and awesome-chatgpt-prompts format (CSV/Markdown)
- **FR-004**: System MUST index prompts from subscribed sources and build searchable metadata (name, description, tags, source)
- **FR-005**: System MUST support checking for updates to subscribed sources via git
- **FR-006**: System MUST allow selective upgrading of individual sources or all sources at once
- **FR-007**: System MUST work with unauthenticated GitHub API access by default (60 requests/hour limit)
- **FR-008**: System MUST track GitHub API rate limit consumption and warn users when threshold exceeded (e.g., 80% of limit used)
- **FR-009**: System MUST provide clear instructions for configuring GitHub personal access token when rate limit warnings appear
- **FR-010**: System MUST support optional GitHub token configuration for higher rate limits (5000 requests/hour) via config file or environment variable
- **FR-041**: System MUST process multiple source operations (subscribe, update) in parallel within a single command
- **FR-042**: System MUST display progress for each source operation when processing multiple sources concurrently
- **FR-043**: System MUST wait for all parallel source operations to complete before returning control to user
- **FR-044**: System SHOULD prevent concurrent pkit command invocations from corrupting shared data (simple approach: fail fast if another instance is running)

**Security:**

- **FR-037**: System MUST sandbox prompt parsers to prevent code execution during parsing
- **FR-038**: System MUST scan subscribed repositories for executable scripts and git hooks, warning users if found
- **FR-039**: System MUST only parse recognized prompt file formats (Markdown, CSV, YAML) and skip other file types
- **FR-040**: System MUST NOT execute any code from subscribed repositories (no eval, exec, or dynamic imports of repo content)

**Search and Discovery:**

- **FR-011**: System MUST provide cross-source search functionality that queries prompt names and descriptions
- **FR-012**: System MUST display search results with source identifier, prompt name, and brief description in table format
- **FR-013**: System MUST support listing all prompts across all sources
- **FR-014**: System MUST support filtering prompts by source (e.g., `--source fabric`)
- **FR-015**: System MUST provide detailed prompt view showing full text, metadata, source, and any tags
- **FR-016**: Search results MUST clearly indicate which source each prompt comes from (e.g., "fabric:code-review")

**Bookmarking and Organization:**

- **FR-017**: Users MUST be able to bookmark prompts with custom aliases
- **FR-018**: Users MUST be able to tag bookmarks with multiple tags (comma-separated)
- **FR-019**: Users MUST be able to filter bookmarks by tags
- **FR-020**: Users MUST be able to rename bookmark aliases
- **FR-021**: Users MUST be able to update tags on existing bookmarks
- **FR-022**: System MUST store bookmarks in ~/.pkit/bookmarks.yml with human-readable format
- **FR-023**: System MUST prevent duplicate aliases and warn users when attempting to create duplicates
- **FR-024**: System MUST validate bookmark file integrity at startup and refuse to start if corrupted, displaying error message with file path and recovery options (manual fix or `pkit init --force`)

**Output and Integration:**

- **FR-025**: System MUST provide `pkit get <alias>` command that outputs ONLY prompt text to stdout
- **FR-026**: Output MUST be pipeable to any external tool (claude, llm, fabric, mods) without formatting interference
- **FR-027**: Errors MUST be written to stderr, not stdout
- **FR-028**: System MUST return appropriate exit codes (0 for success, non-zero for errors)
- **FR-029**: System MUST support `--json` flag for commands to output structured JSON when needed

**Observability:**

- **FR-033**: System MUST be silent by default (no logs to stdout/stderr except errors)
- **FR-034**: System MUST support `--verbose` flag to output detailed operation logs (git operations, file I/O, parsing) to stderr
- **FR-035**: System MUST support `--debug` flag to output full trace logs including timing information to stderr
- **FR-036**: Verbose and debug output MUST NOT interfere with stdout data (e.g., `pkit get --verbose` still outputs only prompt text to stdout)

**Phase 1 Constraints:**

- **FR-030**: System MUST NOT execute prompts or call any LLM APIs in Phase 1
- **FR-031**: System MUST NOT manage API keys, model selection, or backend configuration in Phase 1
- **FR-032**: System MUST focus purely on organization, discovery, and text output

### Key Entities

- **Source**: A GitHub repository containing prompts (e.g., fabric/patterns, f/awesome-chatgpt-prompts). Attributes: URL, short name, local path, last update timestamp, format type.
- **Prompt**: A single prompt template from a source. Attributes: unique ID (source:name), text content, description, metadata, source reference, format-specific fields.
- **Bookmark**: A user's saved reference to a prompt. Attributes: alias (user-defined name), source prompt reference, tags (list), creation date.
- **Index**: Searchable metadata structure. Attributes: prompt mappings, search terms, source references, update timestamps.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can subscribe to a new prompt source and see indexed prompts within 30 seconds for typical repositories (~300 prompts). When subscribing to multiple sources simultaneously, parallel processing should complete in approximately the time of the slowest source, not the sum of all sources.
- **SC-002**: Users can search across all subscribed sources and receive results in under 1 second
- **SC-003**: Users can retrieve a bookmarked prompt via `pkit get` in under 100ms
- **SC-004**: The `pkit get` command outputs clean text that successfully pipes to claude, llm, fabric, and mods without errors
- **SC-005**: Users can complete the full workflow (subscribe → search → bookmark → retrieve → pipe) within 5 minutes on first use
- **SC-006**: 90% of user test scenarios successfully integrate pkit with at least one execution tool (claude/llm/fabric/mods)
- **SC-007**: Search results correctly distinguish between prompts from different sources using clear source identifiers
- **SC-008**: Binary size remains under 20MB for single-file distribution
- **SC-009**: Memory usage stays under 50MB during typical operations
- **SC-010**: Core workflow completes successfully on macOS, Linux, and Windows platforms

## Assumptions

- Users have Git installed on their system (required for cloning repositories)
- Users have internet connectivity to access GitHub repositories
- Subscribed repositories follow standard Git conventions
- Prompt sources use supported formats (Fabric patterns or awesome-chatgpt-prompts initially)
- Users have basic command-line familiarity
- Users already use at least one execution tool (claude, llm, fabric, or mods)
- Local storage (~/.pkit/) has sufficient space for cloned repositories (typically <100MB per source)

## Out of Scope for Phase 1

- Prompt execution or LLM API calls
- API key management
- Model selection or configuration
- Web UI or TUI interface
- Cloud synchronization
- Team features or shared configurations
- Prompt composition or chaining
- Local prompt modifications or forking
- Version locking of specific prompt versions
- Advanced analytics or usage tracking
- Plugin system for custom formats (beyond initial 2-3 formats)
