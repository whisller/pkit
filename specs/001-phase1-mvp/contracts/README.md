# CLI Command Contracts

**Purpose:** Detailed specifications for all pkit CLI commands

**Status:** Phase 1 MVP

## Overview

This directory contains detailed contracts for each CLI command, including:
- Command signature and arguments
- Flags and options
- Input/output specifications
- Examples and workflows
- Error handling
- Edge cases
- Requirements mapping

## Command Categories

### 1. Source Management

Commands for subscribing to and managing prompt sources.

| Command | Status | Description |
|---------|--------|-------------|
| **subscribe** | ✅ Complete | Subscribe to GitHub repository as prompt source |
| **upgrade** | 📝 Stub | Upgrade subscribed sources to latest version (git pull + re-index) |
| **unsubscribe** | 📝 Stub | Remove a subscribed source |
| **status** | 📝 Stub | Show status of sources (prompts, updates available, rate limits) |

### 2. Discovery & Search

Commands for finding and browsing prompts.

| Command | Status | Description |
|---------|--------|-------------|
| **find** | ✅ Complete | Interactive TUI browser with real-time filtering and tagging |
| **search** | 📝 Stub | Traditional keyword search (pipeable, non-interactive) |
| **show** | 📝 Stub | Show detailed prompt information |

### 3. Bookmarking & Organization

Commands for bookmarking and organizing prompts.

| Command | Status | Description |
|---------|--------|-------------|
| **save** | ✅ Complete | Bookmark a prompt with alias and tags |
| **bookmarks** | 📝 Stub | Show all bookmarked prompts |
| **tag** | 📝 Stub | Update tags on bookmarked prompt (CLI alternative to TUI) |
| **alias** | 📝 Stub | Rename a bookmark alias |
| **rm** | 📝 Stub | Remove a bookmark |

### 4. Output & Execution

Commands for retrieving prompt content.

| Command | Status | Description |
|---------|--------|-------------|
| **get** | ✅ Complete | Get prompt content for piping to execution tools |
| **<alias>** | ✅ Complete | Shorthand form of `get` (e.g., `pkit review` = `pkit get review`) |

### 5. Configuration

Commands for managing pkit configuration.

| Command | Status | Description |
|---------|--------|-------------|
| **init** | 📝 Stub | Initialize pkit (create ~/.pkit/ structure) |
| **config** | 📝 Stub | View or edit configuration |
| **config set-token** | 📝 Stub | Securely store GitHub token in keyring |

### 6. Utilities

| Command | Status | Description |
|---------|--------|-------------|
| **help** | 📝 Stub | Show help information |
| **version** | 📝 Stub | Show pkit version |

## Completed Contracts

### subscribe.md

**Command:** `pkit subscribe <source> [flags]`

Subscribe to a GitHub repository as a prompt source. Clones the repository, auto-detects format (Fabric, awesome-chatgpt-prompts, generic markdown), indexes prompts, and adds to config.

**Key Features:**
- Short form: `fabric/patterns` or full URL
- Auto-detect format or force with `--format`
- Parallel subscribe for multiple sources (FR-041, FR-042, FR-043)
- GitHub rate limit tracking and warnings (FR-008, FR-009)
- Authenticated and unauthenticated modes (FR-007, FR-010)

**Performance:** <30 seconds for ~300 prompts (SC-001)

**Examples:**
```bash
$ pkit subscribe fabric/patterns
$ pkit subscribe https://github.com/f/awesome-chatgpt-prompts
$ pkit subscribe fabric/patterns f/awesome-chatgpt-prompts  # parallel
```

---

### get.md

**Command:** `pkit get <alias|prompt-id>` or `pkit <alias>`

Get prompt content for piping to execution tools. **Critical command** for unix pipe workflow.

**Key Features:**
- **Stdout contains ONLY prompt content** (no headers, formatting) (FR-025)
- Errors to stderr (FR-027)
- Shorthand form: `pkit review` = `pkit get review`
- JSON output mode: `--json`
- Usage tracking (increment count, update timestamp)

**Performance:** <100ms retrieval (SC-003)

**Testing Priority:** **CRITICAL** - Must verify pipe compatibility with claude, llm, fabric, mods (SC-004)

**Examples:**
```bash
$ pkit get review | claude -p "analyse me ~/main.go"
$ pkit review | claude -p "analyse me ~/main.go"  # shorthand
$ cat article.txt | pkit sum | claude
$ pkit get review --json | jq '.content'
```

---

### save.md

**Command:** `pkit save <prompt-id> --as <alias> [flags]`

Bookmark a prompt with custom alias and tags.

**Key Features:**
- Required `--as` flag for alias
- Optional `--tags` (comma-separated)
- Optional `--notes`
- Alias validation (lowercase alphanumeric with hyphens/underscores)
- Duplicate prevention (FR-023)
- Force overwrite with `--force`

**Validation:**
- Alias uniqueness
- No conflicts with reserved commands
- Proper prompt ID format (`<source>:<name>`)

**Examples:**
```bash
$ pkit save fabric:summarize --as sum
$ pkit save fabric:code-review --as review --tags dev,security,go
$ pkit save fabric:audit --as audit --tags security --notes "For sensitive code"
```

---

### find.md

**Command:** `pkit find [initial-query]`

Interactive TUI browser powered by Bubbletea with real-time filtering and in-TUI tagging.

**Key Features:**
- Real-time fuzzy search (<100ms per keystroke)
- Navigate with arrow keys
- **Ctrl+S**: Bookmark with form (alias, tags, notes)
- **Ctrl+T**: Edit tags on bookmarked prompt
- **Ctrl+G**: Get prompt content (output after exit)
- **Ctrl+B**: Toggle bookmarks-only filter
- **Ctrl+R**: Clear search / Reset filters
- Preview pane with description
- Bookmark indicators `[★]`
- TTY detection (fall back to search if not terminal)

**Performance:** <100ms filtering (SC-002)

**Keyboard Shortcuts:**
- Type: Filter in real-time
- ↑/↓: Navigate results
- Enter: Show full details
- Ctrl+S: Bookmark
- Ctrl+T: Edit tags
- Ctrl+G: Get content
- Ctrl+B: Toggle bookmarks filter
- Esc: Exit

**Examples:**
```bash
$ pkit find                    # Launch interactive browser
$ pkit find "code review"      # Pre-populate search
$ pkit find | claude           # Get selected prompt, pipe to claude
```

---

## Stub Commands (To Be Detailed)

### upgrade

**Command:** `pkit upgrade <source|--all>`

Upgrade subscribed sources to latest version. Performs `git pull` and re-indexes prompts.

**Flags:**
- `--all`: Upgrade all sources

**Examples:**
```bash
$ pkit upgrade fabric
Upgrading fabric...
✓ Pulled latest changes (5 commits)
✓ Re-indexed 287 prompts (3 new, 2 updated)

$ pkit upgrade --all
Upgrading all sources...
[fabric] ✓ Up to date
[awesome] ✓ 5 new prompts
```

**Note:** No separate `update` command. Just upgrade when ready. Use `pkit status` to see if updates are available.

---

### status

**Command:** `pkit status [source]`

Show status of subscribed sources with prompt counts, update availability, rate limits.

**Examples:**
```bash
$ pkit status
Sources (2):
  fabric (287 prompts)          ✓ Up to date
  awesome (163 prompts)         ⚠ Update available (5 new prompts)

Bookmarks: 12
GitHub API: 4823/5000 requests remaining (96%)
Index: ~/.pkit/cache/index.bleve (1.2 MB)

$ pkit status fabric
Source: fabric
URL: https://github.com/danielmiessler/fabric
Format: fabric_pattern
Prompts: 287
Last indexed: 2025-12-26 10:32:00
Commit: abc123def456...
Status: ✓ Up to date
```

---

### search

**Command:** `pkit search <query> [flags]`

Traditional keyword search for prompts. Non-interactive, pipeable.

**Flags:**
- `--source <source>`: Filter by source
- `--tags <tags>`: Filter by tags (comma-separated, OR logic)
- `--format <format>`: Output format (table, list, json)
- `--limit <n>`: Max results (default: 50)

**Output:** Table format by default, JSON with `--json`

**Examples:**
```bash
$ pkit search "code review"
┌─────────────┬──────────────────┬────────────────────────────────────┐
│ SOURCE      │ NAME             │ DESCRIPTION                         │
├─────────────┼──────────────────┼────────────────────────────────────┤
│ fabric      │ code-review      │ Expert code reviewer analyzing...   │
│ fabric      │ security-review  │ Security-focused code review...     │
└─────────────┴──────────────────┴────────────────────────────────────┘

$ pkit search "summarize" --format json | jq '.prompts[0].id'
"fabric:summarize"
```

---

### bookmarks

**Command:** `pkit bookmarks [flags]`

Show all bookmarked prompts. Small set, no pagination needed.

**Flags:**
- `--tags <tags>`: Filter by tags
- `--format <format>`: Output format (table, list, json)
- `--sort <field>`: Sort by (alias, usage, updated)

**Examples:**
```bash
$ pkit bookmarks
┌─────────────┬──────────────────┬──────────────┬────────────────────┐
│ ALIAS       │ PROMPT           │ TAGS         │ USAGE              │
├─────────────┼──────────────────┼──────────────┼────────────────────┤
│ review      │ fabric:code-rev  │ dev,security │ 12 times           │
│ sum         │ fabric:summarize │ analysis     │ 5 times            │
│ audit       │ fabric:security  │ security     │ 3 times            │
└─────────────┴──────────────────┴──────────────┴────────────────────┘

$ pkit bookmarks --tags security
# Shows only bookmarks tagged with 'security'
```

---

### show

**Command:** `pkit show <prompt-id|alias>`

Show detailed prompt information in formatted view (not for piping, use `get` for that).

**Examples:**
```bash
$ pkit show fabric:code-review
$ pkit show review  # By alias

fabric:code-review [★ review]

Description:
Expert code reviewer analyzing code for bugs, security issues...

Content:
# IDENTITY AND PURPOSE
You are an expert code reviewer...

Tags: dev, security, go
Source: fabric
File: patterns/code-review/system.md
Updated: 2025-12-20
Bookmarked: Yes (usage: 12 times, last used: 2025-12-26)
```

---

### tag

**Command:** `pkit tag <alias> <tags>`

Update tags on bookmarked prompt (CLI alternative to TUI Ctrl+T).

**Examples:**
```bash
$ pkit tag review dev,security,testing
✓ Tags updated for 'review'
  Old: dev, security, go
  New: dev, security, testing

$ pkit tag review ""  # Remove all tags
```

---

### alias

**Command:** `pkit alias <old-alias> <new-alias>`

Rename a bookmark alias.

**Examples:**
```bash
$ pkit alias review code-review
✓ Renamed bookmark: review → code-review
```

---

### rm

**Command:** `pkit rm <alias>`

Remove a bookmark.

**Examples:**
```bash
$ pkit rm review
✓ Removed bookmark 'review' (fabric:code-review)
```

---

### unsubscribe

**Command:** `pkit unsubscribe <source>`

Remove a subscribed source.

**Examples:**
```bash
$ pkit unsubscribe fabric
Warning: This will remove the source and all its prompts from the index.
  Bookmarks referencing these prompts will become orphaned.

Remove 'fabric' (287 prompts)? [y/N]: y
✓ Unsubscribed from fabric
```

---

## Design Principles

### 1. Unix Philosophy

**Stdin → Process → Stdout**

Commands follow unix conventions:
- `get` outputs ONLY content to stdout (FR-025)
- Errors to stderr (FR-027)
- Proper exit codes (FR-028)
- Pipeable to any tool (FR-026)

### 2. Three-Tier UX

1. **Explicit commands**: `pkit subscribe`, `pkit search`, `pkit find`
2. **Shorthand get**: `pkit review` → auto-executes `pkit get review`
3. **Interactive finder**: `pkit find` → real-time TUI

### 3. TTY Awareness

Interactive commands detect TTY:
- `pkit find` in terminal → launches TUI
- `pkit find` piped → falls back to traditional search

### 4. Clean Output Protocol

**Critical for piping** (SC-004):
```bash
# ✓ Correct
$ pkit get review | claude
# Stdout contains ONLY prompt content

# ✗ Wrong
$ pkit get review | claude
=== Prompt: review ===  # Headers break pipes
# IDENTITY...
```

### 5. Simplified Version Management

**No separate `update` command** - just upgrade when ready:
- `pkit upgrade` - Pull latest and re-index (fast, lightweight)
- `pkit status` - Shows if updates available
- Git already tracks what changed (`git log` after upgrade)

**Rationale:** Prompts are text files on disk (already cloned), no dependencies, no breaking changes. No reason to check without upgrading.

## Testing Priority

### Critical Path (P0)

1. **subscribe** → **find** → **save** → **get** → pipe to claude/llm/fabric/mods
   - This is the core workflow that MUST work
   - Test with all 4 execution tools: claude, llm, fabric, mods (SC-004)

### High Priority (P1)

2. **search** - Traditional search for scripting
3. **bookmarks** - View saved prompts
4. **status** - Check sources and updates
5. **upgrade** - Keep sources up to date

### Medium Priority (P2)

6. **tag** / **alias** / **rm** - Bookmark management
7. **show** - Detailed view
8. **unsubscribe** - Remove sources

## Requirements Coverage

Each command contract maps to functional requirements (FR-XXX) and success criteria (SC-XXX) from spec.md.

**Example:**
- `subscribe.md`: FR-001, FR-002, FR-003, FR-004, FR-007, FR-008, FR-009, FR-010, FR-041, FR-042, FR-043, SC-001
- `get.md`: FR-025, FR-026, FR-027, FR-028, FR-029, SC-003, SC-004
- `find.md`: FR-011, FR-012, FR-015, FR-016, FR-017, FR-018, FR-021, SC-002

## User Story Mapping

### User Story 1: Subscribe and Discover (P1)

**Commands:** `subscribe`, `find`, `search`, `show`

**Workflow:**
```bash
$ pkit subscribe fabric/patterns
$ pkit find "code review"  # Interactive browser
$ pkit search "summarize" --source fabric  # Traditional search
$ pkit show fabric:summarize
```

### User Story 2: Bookmark and Organize (P2)

**Commands:** `save`, `bookmarks`, `tag`, `alias`, `rm`

**Workflow:**
```bash
$ pkit save fabric:code-review --as review --tags dev,security
$ pkit bookmarks
$ pkit tag review dev,security,go  # Add tag
$ pkit alias review code-review  # Rename
```

### User Story 3: Pipe to Execution Tools (P1)

**Commands:** `get`, shorthand

**Workflow:**
```bash
$ pkit get review | claude -p "analyse me ~/main.go"
$ pkit review | claude -p "analyse me ~/main.go"  # shorthand
$ cat article.txt | pkit sum | claude
$ pkit get audit | llm -m claude-3-sonnet < ~/auth.go
```

### User Story 4: Track Source Updates (P3)

**Commands:** `status`, `upgrade`

**Workflow:**
```bash
$ pkit status  # Shows which sources have updates
$ pkit upgrade fabric  # Upgrade specific source
$ pkit upgrade --all  # Upgrade all
```

## Next Steps

1. ✅ Complete core contracts (subscribe, get, save, find)
2. Create stub command contracts (search, bookmarks, status, upgrade, etc.)
3. Review all contracts for consistency
4. Validate against spec.md requirements
5. Begin implementation based on contracts
6. Write tests based on "Testing Checklist" sections

## Command Summary

**Total Commands:** 15

**Completed Contracts (4):**
- subscribe
- get (+ shorthand)
- save
- find

**Stub Contracts (11):**
- upgrade
- status
- search
- bookmarks
- show
- tag
- alias
- rm
- unsubscribe
- init
- config
- help
- version

## Notes

- **No `list` command**: Use `find` (interactive) or `search` (traditional) instead
- **No separate `update` command**: Just `upgrade` when ready, `status` shows availability
- **Bookmarks-only view**: Use `pkit bookmarks` instead of `pkit list --bookmarks`
- **Source statistics**: Use `pkit status` for overview
