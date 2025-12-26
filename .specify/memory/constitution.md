# pkit Constitution

<!--
SYNC IMPACT REPORT
==================
Version Change: Initial → 1.0.0
Constitution Type: MAJOR (initial ratification)

Core Principles Established:
- I. Organization-First Architecture
- II. CLI-First Interface
- III. Tool Agnosticism
- IV. Multi-Source Aggregation
- V. Simple Output Protocol
- VI. Phase-Gated Development
- VII. Simplicity & Focus

Additional Sections:
- Technical Constraints
- Development Workflow
- Governance

Templates Requiring Updates:
✅ plan-template.md - Constitution Check section aligned
✅ spec-template.md - User scenarios align with CLI-first, organization-first principles
✅ tasks-template.md - Phase-based organization aligns with constitution phases

Follow-up Actions:
- None - all placeholders resolved
- Constitution ready for initial ratification
-->

## Core Principles

### I. Organization-First Architecture

**pkit is a bookmark and package manager for AI prompts, NOT an execution engine or prompt creator.**

- MUST focus on aggregation, organization, and discovery of prompts from multiple sources
- MUST NOT compete with execution tools (Fabric, llm, mods, claude-code) - complement them instead
- Phase 1 (MVP) MUST deliver organization-only functionality with NO execution
- Optional execution wrappers (Phase 2+) MUST be thin convenience layers that delegate to user-chosen backends
- Every feature MUST answer: "Does this help users organize and find prompts better?"

**Rationale**: Clear positioning avoids competing with established tools and focuses on unique value proposition - multi-source prompt aggregation. Organization-first validates core value before adding execution complexity.

### II. CLI-First Interface

**Simple CLI with Unix philosophy: text in, text out.**

- MUST provide simple text-based commands that work with pipes and scripts
- MUST support command chaining and Unix pipes
- Output formats MUST be parseable and composable (tables, lists, JSON when needed)
- Web UI (Phase 2+) is for browsing/visual organization only - CLI remains primary interface
- MUST follow successful patterns: Homebrew, npm, Docker (simple CLI + web for browsing)

**Example workflows:**
```bash
# With Claude
pkit get code-review | claude -p "analyse me ~/myproject/main.go"
cat ~/mytext.txt | pkit get summarize | claude

# With llm
pkit get security-audit | llm -m claude-3-sonnet

# With Fabric
pkit get improve-writing | fabric

# With mods
pkit get explain-code | mods
```

**Rationale**: CLI-first ensures automation and scripting remain first-class. Works seamlessly with any execution tool through pipes. Faster MVP, better Unix integration.

### III. Tool Agnosticism

**pkit works WITH any execution tool - user's choice of backend.**

- MUST output prompt text to stdout for piping to ANY tool (claude, llm, fabric, mods, aichat, etc.)
- MUST NOT require specific LLM providers or API keys in Phase 1
- MUST NOT manage model selection or LLM configuration in Phase 1
- Optional execution wrapper (Phase 2+) MUST support configurable backends
- Source format adapters MUST be pluggable and extensible

**Rationale**: Maximizes adoption by working with users' existing workflows. No lock-in to specific tools or providers. Makes pkit more valuable to execution tools rather than competing with them.

### IV. Multi-Source Aggregation

**Aggregate prompts from multiple GitHub repositories with unified search and version tracking.**

- MUST support subscribing to multiple GitHub repos as prompt sources
- MUST provide cross-source search and unified indexing
- MUST track source versions and notify users of updates
- MUST support personal bookmarking with custom aliases and tags across all sources
- Phase 1 MUST support at minimum: Fabric patterns, awesome-chatgpt-prompts, generic markdown

**Example workflow:**
```bash
# Subscribe to multiple sources
pkit subscribe fabric/patterns
pkit subscribe f/awesome-chatgpt-prompts
pkit subscribe company/internal-prompts

# Search across all sources
pkit search "code review"

# Bookmark and use
pkit save fabric:code-review --as review --tags dev,security
pkit get review | claude -p "analyse me ~/app/auth.go"
```

**Rationale**: This is pkit's unique value proposition - no other tool does multi-source prompt organization. Network effects work FOR us: more sources = more value. No cold start problem since we aggregate existing content.

### V. Simple Output Protocol

**Text I/O ensures debuggability and composability.**

- Commands MUST output clean, parseable text (similar to git, docker, npm)
- `pkit get` MUST output ONLY prompt text to stdout (no headers, no formatting) - ready to pipe
- Search/list commands MUST use clear table/list formats for human readability
- Errors MUST go to stderr
- JSON output MUST be available via `--json` flag for programmatic use

**Rationale**: Unix-friendly output enables automation, scripting, and piping. Clean separation of concerns: stdout for data, stderr for errors. Follows established CLI tool conventions.

### VI. Phase-Gated Development

**Three-phase approach validates value before adding complexity.**

**Phase 1 (Months 1-3) - Pure Bookmark Manager:**
- MUST deliver organization-only: subscribe, search, save, get (output text)
- MUST NOT include execution, API calls, or model management
- Success gate: Core workflow validated - users can subscribe, search, bookmark, and pipe to execution tools
- MUST validate core hypothesis: multi-source organization is useful

**Phase 2 (Months 4-6) - Web UI + Optional Wrapper:**
- MUST add local web UI for visual browsing/organizing
- Features: rich prompt preview, drag-and-drop tagging, side-by-side comparison, usage analytics
- MAY add thin execution wrapper (`pkit run`) if Phase 1 users request it
- MAY add local modifications (fork/edit prompts)
- Success gate: Web UI provides measurable value for prompt discovery and organization
- MUST validate web UI adds value beyond CLI

**Phase 3 (Months 7-12) - Unique Features:**

This phase adds features that Fabric, llm, and mods DON'T have - true differentiation:

**Cross-source Composition:**
- Combine prompts from different sources into custom workflows
- Example: `pkit compose fabric:summarize + awesome:improve-writing --as my-combo`
- Compose multi-step prompt chains that execute sequentially
- Enable building complex workflows from simple prompt building blocks

**Version Locking & Management:**
- Pin specific versions of prompts for reproducibility
- Track upstream changes without automatic updates
- Example: `pkit lock fabric:summarize@v1.2.3`
- Compare local modifications against upstream: `pkit diff my-sum fabric:summarize`
- Lock all current versions: `pkit lock --all` for stable environments

**Local Modifications with Tracking:**
- Fork upstream prompts and modify locally
- Track differences from upstream source
- Optionally sync upstream updates into local forks
- Example: `pkit fork fabric:summarize --as my-sum`, edit, then `pkit sync my-sum`
- Maintain personal variations while tracking upstream improvements

**Team Configuration & Sharing:**
- Shared team configs checked into git (`.pkit/team.yml`)
- Standardize prompt usage across teams
- Version-locked subscriptions for consistency
- Team members sync with single command: `pkit sync`
- Enables reproducible prompt workflows across team members

**Prompt Chaining:**
- Chain multiple prompts in single execution
- Example: `pkit chain review improve translate-es --output combined`
- Enables complex multi-step workflows
- Single command for multi-stage prompt processing

**Advanced Search & Discovery:**
- Fuzzy search across prompt content, not just titles
- Full-text search within prompt bodies
- Tag-based filtering and recommendations
- Usage analytics: most-used prompts, effectiveness tracking
- Smart suggestions based on usage patterns
- "Similar prompts" recommendations

Success gate: Features validate uniqueness - users choose pkit for capabilities unavailable elsewhere. Team features demonstrate collaboration value. Advanced features show clear differentiation from execution-only tools.

**Rationale**: Phase 3 features differentiate pkit from execution tools. Focus on organization, versioning, and team collaboration - areas where execution-focused tools fall short. Each feature builds on validated Phase 1/2 foundation.

### VII. Simplicity & Focus

**Codebase simplicity wins. Avoid over-engineering.**

- MUST start simple - no abstractions until pattern emerges multiple times (Rule of Three)
- MUST avoid backwards-compatibility hacks - if something is unused, delete it completely
- MUST NOT add features beyond what was requested or clearly necessary
- MUST NOT create helpers/utilities for one-time operations
- MUST NOT design for hypothetical future requirements
- Three similar lines of code is better than a premature abstraction
- MUST delete dead code immediately - no commenting out "for later"
- Simple, readable code beats clever code every time
- MUST refactor when duplication becomes problematic, not preemptively
- MUST keep functions short and focused - extract when intent becomes unclear

**Rationale**: YAGNI (You Aren't Gonna Need It) principle. Complexity is the enemy of maintainability. Start simple, refactor when actual need emerges. Deleted code has no bugs. Readable code is maintainable code.

## Technical Constraints

### Distribution

- **Distribution**: Single binary with no runtime dependencies
- **Platforms**: macOS, Linux, Windows support from Phase 1
- **Installation**: Installable via package managers (Homebrew, etc.)

### User Data Structure

**All user data stored in ~/.pkit/:**
```
~/.pkit/
├── config.yml          # User configuration
├── sources/            # Subscribed repos (git clones)
├── bookmarks.yml       # User's saved prompts
├── tags.yml            # Tag mappings
└── cache/              # Search index
```

### Source Format Support

**Phase 1 (MVP):**
- Fabric patterns (Markdown with frontmatter)
- awesome-chatgpt-prompts (CSV/Markdown)
- Simple Markdown files

**Phase 2+:**
- YAML prompt definitions
- Plugin system for custom formats

### Performance Standards

- Subscribe command: <30 seconds for typical repo (~300 patterns)
- Search across all sources: <1 second for indexed search
- `pkit get`: <100ms (prompt text retrieval)
- Binary size: Target <20MB
- Memory: <50MB typical usage

## Development Workflow

### MVP Timeline (Phase 1 - Weeks 3-4)

**Week 3 Core Commands:**
- `pkit subscribe` - Clone and index repos
- `pkit list` - List all prompts
- `pkit show` - View prompt details
- Support Fabric pattern format

**Week 4 Organization Features:**
- `pkit search` - Cross-source search
- `pkit save` - Bookmark with aliases/tags
- `pkit get` - Output prompt text to stdout (pipeable)
- `pkit update` - Version tracking
- Support awesome-chatgpt-prompts format
- Basic documentation

### Quality Gates

**Before Phase 1 Launch:**
- All core commands functional
- 2+ source formats supported (Fabric, awesome-chatgpt-prompts minimum)
- Basic documentation complete
- `pkit get` outputs ONLY prompt text (pipeable to claude, llm, fabric, mods)
- Examples work: `pkit get review | claude -p "analyse me ~/file.go"`
- No obvious performance issues (meets performance standards)
- Code is simple, readable, and maintainable

**Before Phase 2 (Web UI):**
- Phase 1 functionality validated by real usage
- User feedback collected and analyzed
- Architecture stable and extensible
- All Phase 1 tests passing
- Codebase remains simple and maintainable

**Before Phase 3 (Advanced Features):**
- Phase 2 features validated
- Feature differentiation strategy validated
- Web UI adoption and usage patterns understood
- Team features have identified use cases
- No technical debt blocking advanced features

### Code Review Requirements

- All PRs MUST verify compliance with Core Principles
- Constitution violations MUST be justified in plan.md Complexity Tracking table
- New commands MUST follow Simple Output Protocol
- Source adapters MUST be pluggable and tested
- Phase boundaries MUST NOT be crossed without gate validation
- Simplicity MUST be enforced - reject premature abstractions
- Error handling MUST be explicit
- Tests MUST be included for new functionality
- Code MUST be readable and maintainable - reject clever/obscure implementations

## Governance

**This constitution supersedes all other development practices.**

### Amendment Process

1. Proposed changes MUST be documented with rationale
2. Constitution version MUST increment according to semantic versioning:
   - **MAJOR**: Backward incompatible governance/principle removals or redefinitions
   - **MINOR**: New principle/section added or materially expanded guidance
   - **PATCH**: Clarifications, wording, typo fixes, non-semantic refinements
3. Changes MUST include migration plan if affecting existing work
4. All dependent templates MUST be updated (.specify/templates/*.md)

### Compliance Review

- All PRs and reviews MUST verify compliance with this constitution
- Complexity MUST be justified against Core Principles (especially Principle VII: Simplicity)
- Phase-gating MUST be enforced - no execution in Phase 1, no advanced features in Phase 2
- Use this constitution for runtime development guidance

### Strategic Positioning Enforcement

- pkit MUST position as "package manager for prompts" (like Homebrew, npm, Pocket)
- Marketing/docs MUST emphasize complementary relationship with claude, fabric, llm, mods
- Feature decisions MUST prioritize organization and discovery over execution
- Success metrics MUST track feature validation and user workflow improvements
- Examples MUST show integration with multiple tools (claude, llm, fabric, mods)
- Codebase simplicity MUST be prioritized over feature completeness

**Version**: 1.0.0 | **Ratified**: 2025-12-25 | **Last Amended**: 2025-12-25
