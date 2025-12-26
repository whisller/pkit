# pkit - Your Prompt Toolkit

> A bookmark and package manager for AI prompts from any source

**Status:** Pre-MVP / Planning Phase
**Domain:** pkit.ai (available)
**Created:** 2025-12-23

---

## 🎯 Project Vision

pkit is a **bookmark and package manager for AI prompts**. Think of it as **Homebrew meets Pocket for AI prompts** - an aggregator and organizer, NOT an execution engine or prompt creator.

### Core Philosophy

- **Organization-first** - Bookmark, tag, and search prompts from multiple sources
- **Aggregator, not creator** - We don't create prompts, we organize them
- **Bookmark, not execute** - Phase 1 focuses purely on organization (execution is Phase 2+)
- **CLI-first** - Unix pipes and command-line workflow
- **Multi-source** - Subscribe to any GitHub repo with prompts
- **Personal workspace** - Tag, rename, modify prompts locally
- **Version tracking** - Get notified when prompt sources update
- **Team-ready** - Path to shared configs and team sync

### Key Differentiation

**pkit is NOT competing with Fabric, llm, or mods for execution.**

Instead, pkit makes those tools MORE useful by:
- Aggregating prompts from multiple sources
- Providing cross-source search and organization
- Tracking versions of subscribed prompt libraries
- Enabling team-wide prompt standardization

---

## 🚀 What Problem Does This Solve?

### Current Pain Points

**Today's workflow is broken:**
```bash
# Users currently do this:
cd ~/repos/fabric && git pull
cd ~/repos/awesome-chatgpt-prompts && git pull
cd ~/repos/company-prompts
grep -r "summarize" ../fabric ../awesome-chatgpt-prompts
# Copy-paste prompt manually
# Run with llm or fabric
```

**With pkit:**
```bash
# One-time setup
pkit subscribe fabric/patterns
pkit subscribe f/awesome-chatgpt-prompts

# Find and bookmark
pkit search summarize
pkit save fabric:summarize --as sum

# Use with your preferred tool
cat article.txt | pkit get sum | llm -m claude-3-sonnet
# Or: cat article.txt | pkit get sum | fabric
# Or: cat article.txt | pkit get sum | mods
```

### Target Users

**Primary (Free Tier):**
- Developers using CLI tools (llm, fabric, aichat)
- Power users managing multiple prompt sources
- Teams wanting to standardize prompts
- DevRel/technical writers

**Secondary (Paid Tier - Future):**
- Engineering teams (5-50 developers)
- Content teams using AI heavily
- Agencies managing client prompts

---

## 🏆 Competition Analysis

### The CLI Ecosystem

**Execution Tools (Run prompts with LLMs):**
- **Fabric** - 23k+ stars, 300+ curated patterns, Unix pipes
- **llm** (Simon Willison) - Multi-provider, template system
- **mods** (Charm) - Beautiful CLI, roles system, MCP support
- **shell-gpt** - AI assistant with roles/templates
- **AIChat** - All-in-one with roles and sessions
- **chatblade** - Swiss Army Knife for ChatGPT

**Testing/Evaluation Tools:**
- **promptfoo** - Test and compare prompts, CI/CD integration

**Organization Tools:**
- **pkit** ← YOU ARE HERE

### What Makes pkit Different

| Tool | Multi-source | CLI | Outputs Text | Reusable Prompts | Version Tracking | Cross-source Tags |
|------|--------------|-----|--------------|------------------|------------------|-------------------|
| **pkit** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Fabric | ❌ | ✅ | ❌ (executes) | ✅ (300+) | ❌ | ❌ |
| llm | ❌ | ✅ | ❌ (executes) | ✅ (templates) | ❌ | ❌ |
| mods | ❌ | ✅ | ❌ (executes) | ✅ (roles) | ❌ | ❌ |
| shell-gpt | ❌ | ✅ | ❌ (executes) | ✅ (roles) | ❌ | ❌ |
| AIChat | ❌ | ✅ | ❌ (executes) | ✅ (roles) | ❌ | ❌ |

### Why This Will Work

1. **No direct competition** - Nobody is doing multi-source prompt aggregation with CLI
2. **Complementary, not competitive** - We make Fabric, llm, mods MORE useful
3. **Clear value from day 1** - Users can subscribe to existing sources immediately
4. **No cold start** - We aggregate, don't create content
5. **Network effects work FOR us** - More sources = more value
6. **Tool-agnostic** - Works with ANY execution backend (llm, fabric, mods, etc.)

### Strategic Positioning

**pkit is the package manager for prompts.**

Just like:
- Homebrew manages packages from multiple taps
- npm manages packages from multiple registries
- Pocket manages bookmarks from multiple sources

**pkit manages prompts from multiple sources and lets YOU choose the execution tool.**

**Likelihood of adoption:** 70-80%
**Likelihood of profitability:** 40-50%
**Likelihood of personal usefulness:** 95%

---

## 📦 Features & Development Phases

### Phase 1: Pure Bookmark Manager (Months 1-3) - MVP

**Goal:** Validate core value proposition - multi-source prompt organization

**Core Commands (NO EXECUTION):**
```bash
# Subscribe to sources
pkit subscribe <repo-url>
pkit subscribe fabric/patterns
pkit subscribe f/awesome-chatgpt-prompts
pkit subscribe https://github.com/company/internal-prompts

# Search and browse
pkit search <query>                # Search across all sources
pkit list                         # List all prompts
pkit list --source fabric         # Filter by source
pkit show <prompt-id>             # View prompt details

# Bookmark and organize
pkit save <prompt-id> --as <alias> --tags <tags>
pkit tag <alias> <tags>           # Add/update tags
pkit alias <alias> <new-alias>    # Rename bookmark

# Get prompt text (for piping to tools)
pkit get <alias>                  # Output prompt text to stdout
cat file.txt | pkit get review | llm -m claude-3-sonnet
cat file.txt | pkit get review | fabric
cat file.txt | pkit get review | mods

# Version management
pkit update                       # Check for updates
pkit status                       # Show outdated sources
pkit upgrade <source>             # Update specific source
pkit upgrade --all                # Update all sources
```

**What Phase 1 Does:**
- ✅ Aggregates prompts from multiple GitHub repos
- ✅ Provides unified search across all sources
- ✅ Bookmarking with custom aliases and tags
- ✅ Version tracking and notifications
- ✅ **Outputs prompt text ONLY** (no execution)

**What Phase 1 Does NOT Do:**
- ❌ Execute prompts (that's for llm/fabric/mods)
- ❌ Call LLM APIs
- ❌ Manage API keys or models

**Success Metrics:**
- 100 users
- 10+ GitHub issues/discussions
- Proof that organization-only is valuable
- Users requesting execution features

---

### Phase 2: Web UI + Polish (Months 4-6)

**Goal:** Add visual browsing/organizing interface

**Primary Feature: Web UI**
```bash
# Launch local web interface
pkit web                          # Opens http://localhost:3000

# Or run on custom port
pkit web --port 8080

# Web UI features:
# - Visual browse/search across all sources
# - Rich prompt preview with syntax highlighting
# - Drag-and-drop tagging
# - Side-by-side comparison
# - Usage analytics and history
# - Visual bookmarking
# - Filter/sort by source, tags, date
```

**Why Web UI in Phase 2?**
- ✅ Better for browsing 300+ prompts than CLI
- ✅ Visual organization (tags, drag-drop)
- ✅ Foundation for Phase 3 (cloud sync, teams)
- ✅ Follows successful patterns (Homebrew formulae, npm, Docker Hub)

**Web Architecture:**
```
Local-first approach:
- Web UI reads from ~/.pkit/ directly
- No cloud dependency
- Optional cloud sync in Phase 3
```

**Secondary Feature: Optional Execution Wrapper**
```bash
# Optional: Configure preferred backend
pkit config set backend llm
pkit config set model claude-3-sonnet

# Optional convenience
pkit run review
# Internally: llm -m claude-3-sonnet -s "$(pkit get review)"

# CLI still works as before
pkit get review | mods
```

**Additional Polish:**
- Local modifications (fork/edit prompts)
- Better CLI search (fuzzy, full-text)
- Shell completions (bash, zsh, fish)
- Analytics (most-used prompts)
- Improved documentation

**Success Metrics:**
- 1000 users
- 500 GitHub stars
- 50%+ users try web UI
- Validate web adds value for browsing

---

### Phase 3: Unique Execution Features (Months 7-12)

**Goal:** Add features that Fabric/llm/mods DON'T have

**Advanced Features:**

**1. Cross-source Composition**
```bash
# Combine prompts from different sources
pkit compose fabric:summarize + awesome:improve-writing --as my-combo
cat file.txt | pkit get my-combo | llm
```

**2. Version Locking & Management**
```bash
# Pin specific versions
pkit lock fabric:summarize@v1.2.3
pkit lock --all                   # Lock all current versions
pkit diff my-sum fabric:summarize # Compare local vs upstream
```

**3. Local Modifications with Tracking**
```bash
# Fork and modify
pkit fork fabric:summarize --as my-sum
pkit edit my-sum
pkit diff my-sum fabric:summarize # See what changed
pkit sync my-sum                  # Pull upstream updates
```

**4. Team Configuration**
```bash
# .pkit/team.yml (checked into git)
subscriptions:
  - fabric/patterns@v2.1.0
  - company/prompts@main
bookmarks:
  review: fabric:code-review
  sum: company:custom-summarize

# Team members just:
pkit sync                         # Pull team config
```

**5. Prompt Chaining**
```bash
# Chain multiple prompts (single LLM call)
pkit chain review improve translate-es --output combined
cat file.txt | pkit get combined | llm
```

**Success Metrics:**
- 5000 users
- 10 paying teams
- Team features validated
- $500-1000 MRR

---

### Future Ideas (Phase 4+)

- Cloud sync (paid feature)
- Team workspace web UI
- Private prompt storage
- Variable templating
- A/B testing framework
- Performance analytics
- CI/CD integration
- VS Code/JetBrains extensions
- Claude Code MCP integration
- Prompt testing integration (with promptfoo)

---

## 🎨 Name: pkit

### Why "pkit"?

**Evaluation Criteria:**
- ✅ Short (4 letters - easy to type)
- ✅ Clear readability (p, k, i, t are distinct in all fonts)
- ✅ Meaningful ("prompt kit" is self-explanatory)
- ✅ Professional (sounds like a real dev tool)
- ✅ Available (domain pkit.ai available)
- ✅ Memorable and brandable

**Rejected alternatives:**
- `quiver` - Too many similar-looking letters (q, u)
- `stash` - Too associated with git
- `prmt` - Hard to pronounce
- `deck`, `shelf`, `cache` - Less distinctive

**Taglines:**
- "Your prompt toolkit"
- "Prompts, packaged"
- "The prompt kit for developers"

---

## 🏗️ Technical Architecture

### Tech Stack

**Language:** Go
- Single binary distribution
- Excellent CLI tooling (Cobra)
- Fast compilation and iteration
- Great for system tools

**CLI Framework:** Cobra (standard Go CLI framework)
- Used by kubectl, hugo, github CLI
- Subcommands, flags, help generation
- No TUI complexity - simple text I/O

**UI Strategy:** CLI-first, Web later
- Phase 1: Simple CLI only (text output)
- Phase 2: Add Web UI for browsing/organizing
- Phase 3: Cloud sync via Web

**Why Simple CLI + Web (not TUI)?**
1. ✅ **Faster MVP** - 2-3 weeks vs 8-10 weeks for TUI
2. ✅ **Web is better for browsing** - Rich formatting, visual organization
3. ✅ **Follows successful patterns** - Homebrew, npm, Docker all use simple CLI + Web
4. ✅ **Need Web anyway** - For team features, monetization (Phase 3)
5. ✅ **Best of both** - CLI for automation/piping, Web for exploration
6. ✅ **Unix-friendly** - Simple text I/O works with pipes and scripts

**Comparison:**

| Approach | Dev Time | Browsing UX | Piping | Team Features | Mobile |
|----------|----------|-------------|--------|---------------|---------|
| **Simple CLI + Web** | 2-3 weeks + 4-6 weeks | ⭐⭐⭐⭐⭐ (web) | ✅ | ✅ (web) | ✅ (web) |
| **TUI (Bubble Tea)** | 8-10 weeks | ⭐⭐⭐ | ⚠️ | ❌ | ❌ |
| **CLI only** | 2-3 weeks | ⭐⭐ | ✅ | ❌ | ❌ |

**Successful examples of this pattern:**
- **Homebrew**: Simple CLI + [formulae.brew.sh](https://formulae.brew.sh/)
- **npm**: Basic CLI + [npmjs.com](https://npmjs.com/)
- **Docker Hub**: CLI search + [hub.docker.com](https://hub.docker.com/)
- **GitHub**: `gh` CLI + [github.com](https://github.com/)

### Directory Structure

**User data (~/.pkit/):**
```
~/.pkit/
├── config.yml          # User configuration
├── sources/           # Subscribed repos (git clones)
│   ├── fabric/
│   ├── awesome-chatgpt-prompts/
│   └── company-prompts/
├── bookmarks.yml      # User's saved prompts
├── tags.yml          # Tag mappings
└── cache/            # Search index
```

**Project structure (Phase 1 - Simple CLI):**
```
pkit/
├── cmd/
│   └── pkit/         # CLI entry point (Cobra)
├── internal/
│   ├── config/       # Config management
│   ├── sources/      # Source adapters (git clone, index)
│   ├── search/       # Cross-source search
│   ├── bookmarks/    # Bookmark/tag management
│   ├── display/      # Text formatting (tables, lists)
│   └── execution/    # Thin wrapper (Phase 2+)
├── adapters/         # Source format parsers
│   ├── fabric/       # Fabric pattern parser
│   ├── markdown/     # Generic markdown
│   └── yaml/         # YAML prompts
├── docs/
├── examples/
└── scripts/
```

**Project structure (Phase 2+ with Web UI):**
```
pkit/
├── cmd/
│   └── pkit/         # CLI (unchanged)
├── internal/         # Shared core logic
│   └── ...
├── web/              # Web UI (Phase 2+)
│   ├── server/       # HTTP server (serves local UI)
│   ├── frontend/     # React/Vue/Svelte UI
│   └── api/          # JSON API for UI
├── adapters/
├── docs/
└── examples/
```

### CLI Output Style (Phase 1)

**Simple, parseable text output** - like git, docker, npm:

```bash
$ pkit search "review"

Found 12 prompts matching "review":

  1. fabric:code-review          Review code for bugs and style
  2. fabric:pr-review            Review GitHub pull requests
  3. awesome:security-review     Security-focused code review

Use 'pkit show <id>' for details
Use 'pkit save <id> --as <alias>' to bookmark

$ pkit list --tags security

Your bookmarks (tagged: security):

  review-sec   → fabric:security-review
  audit        → company:security-audit

$ pkit get review-sec
[Outputs prompt text to stdout - ready to pipe]
```

**Benefits:**
- ✅ Fast to build (2-3 weeks)
- ✅ Unix-friendly (pipes, scripts)
- ✅ No TUI complexity
- ✅ Familiar to developers

### Execution Strategy by Phase

**Phase 1 (MVP - Simple CLI):**
- Text-based commands only
- pkit ONLY outputs prompt text
- No execution, no backend configuration
- No TUI, no web UI yet
- Users pipe to their tool of choice

```bash
# Core Phase 1 usage
pkit subscribe fabric/patterns
pkit search "code review"
pkit show fabric:code-review
pkit save fabric:code-review --as review
pkit get review | llm -m claude-3-sonnet
```

**Phase 2 (Web UI + Optional Execution):**

1. **Web UI for browsing/organizing**
```bash
# Launch local web interface
pkit web                          # Opens http://localhost:3000

# Web UI provides:
# - Visual browse/search
# - Rich prompt preview
# - Drag-and-drop tagging
# - Comparison view
# - Usage analytics
```

2. **Optional thin execution wrapper**
```yaml
# ~/.pkit/config.yml (optional)
backend: llm
model: claude-3-sonnet
```

```bash
# Optional convenience
pkit run review                   # Wrapper around backend
# Internally: llm -m claude-3-sonnet -s "$(pkit get review)"

# CLI approach still works
pkit get review | mods
```

**Architecture (Phase 2):**
```
┌──────────┐
│   CLI    │──┐
└──────────┘  │
              ├──▶ ~/.pkit/ (local data)
┌──────────┐  │
│  Web UI  │──┘   (runs locally on localhost)
└──────────┘
```

**Phase 3 (Cloud Sync + Advanced Features):**
- Cloud sync via web
- Team workspace
- Cross-source composition
- Version locking
- Team configurations

**Architecture (Phase 3):**
```
┌──────────┐
│   CLI    │──┐
└──────────┘  │
              ├──▶ Local + Cloud sync
┌──────────┐  │
│  Web UI  │──┘   (app.pkit.ai or localhost)
└──────────┘
```

### Source Format Support

**Phase 1 (MVP):**
- Fabric patterns (Markdown with frontmatter)
- awesome-chatgpt-prompts (CSV/Markdown)
- Simple Markdown files

**Phase 2:**
- YAML prompt definitions
- JSON API responses
- Custom formats via plugin system

---

## 📊 Go-to-Market Strategy

### Phase 1: MVP Launch (Months 1-3)

**Goal:** Validate concept

**Tactics:**
- Build 5 core commands
- Support 2-3 major sources (Fabric, awesome-chatgpt-prompts)
- Launch on Hacker News: "Show HN: Bookmark manager for AI prompts"
- Post to r/LocalLLaMA, r/ClaudeAI, r/commandline
- Tweet to @simonw (llm creator), mention Fabric community
- Write blog post about the Unix philosophy + AI prompts

**Success:** 100 users in first month

### Phase 2: Growth (Months 4-6)

**Goal:** Build community

**Tactics:**
- Add polish and missing features
- Create great documentation
- Integration guides (llm, fabric, aichat)
- Video demos
- Community prompt source directory

**Success:** 1000 users, 500 stars

### Phase 3: Monetization Test (Months 7-12)

**Goal:** Find paying customers

**Tactics:**
- Identify paying use cases
- Build team features
- Create landing page on pkit.ai
- Test pricing ($49/mo for teams)
- Enterprise outreach

**Success:** 10 paying teams, $500-1000 MRR

---

## 💰 Monetization Strategy

### Free Tier (Forever)

- CLI tool (full-featured)
- Subscribe to unlimited sources
- Local bookmarks/tags
- Community support

### Paid Tier ($9/mo - Individual)

- Cloud sync across machines
- Advanced search/analytics
- Priority support
- Early access to features

### Team Tier ($49/mo - Up to 10 users)

- Shared bookmarks/configs
- Team workspace web UI
- Private prompt storage
- Version history
- Team analytics

### Enterprise ($499/mo)

- SSO/SAML
- Audit logs
- Custom integrations
- Dedicated support
- On-premise deployment

---

## 🗂️ Repository Strategy

### Phase 1 (Now): Single Personal Repo

```
github.com/yourusername/pkit
```

**Why:**
- Faster to start
- Easier to iterate
- Full control during experimentation
- Can transfer to org later (GitHub makes this easy)

**Repo Structure:**
```
pkit/
├── cmd/              # CLI
├── internal/         # Core logic
├── adapters/         # Source parsers
├── docs/            # Documentation
├── examples/        # Example configs
└── README.md
```

### Phase 2 (After 500+ stars): Organization

```
github.com/pkit-dev/
├── pkit              # Main CLI
├── pkit.ai          # Website/docs
├── adapters         # Community adapters (optional)
└── .github          # Org profile
```

**When to migrate:**
- 500+ GitHub stars
- Regular contributors
- Need team access controls
- Want to add separate website repo

---

## 🌐 Domain & Branding

### Domain

**Primary:** pkit.ai (available - register now!)

**Subdomain Strategy:**
- `pkit.ai` - Main website/docs
- `app.pkit.ai` - Team sync web app (future)
- `registry.pkit.ai` - Prompt registry (future)
- `api.pkit.ai` - Sync API (future)

### Logo/Brand Ideas

- Minimalist toolkit icon
- Developer-focused aesthetic
- Color scheme: Terminal-friendly (blues, greens)
- Avoid over-designed logos

---

## 📋 Immediate Next Steps

### Week 1: Validation

- [ ] Talk to 10 developers who use AI prompts
- [ ] Validate they have the organization problem
- [ ] Confirm they'd use a CLI tool
- [ ] Ask about pricing willingness

### Week 2: Setup

- [ ] Register pkit.ai domain
- [ ] Create GitHub repo: yourusername/pkit
- [ ] Set up Go project with Cobra
- [ ] Set up basic project structure (cmd/, internal/, adapters/)
- [ ] Write initial README
- [ ] Set up GitHub Actions CI

### Weeks 3-4: MVP Development (Phase 1 - Simple CLI Only)

**Week 3:**
- [ ] Implement `pkit subscribe` - Clone and index repos
- [ ] Implement `pkit list` - List all prompts (simple table)
- [ ] Implement `pkit show` - View prompt details
- [ ] Support Fabric pattern format parser

**Week 4:**
- [ ] Implement `pkit search` - Cross-source search
- [ ] Implement `pkit save` - Bookmark with aliases/tags
- [ ] Implement `pkit get` - Output prompt text to stdout
- [ ] Implement `pkit update` - Version tracking
- [ ] Support awesome-chatgpt-prompts format
- [ ] Add pterm for pretty output
- [ ] Write basic docs

**Important:** NO EXECUTION, NO TUI - Phase 1 is simple CLI bookmark-only

### Week 5: Polish & Launch

- [ ] Polish README
- [ ] Create demo video
- [ ] Post to Hacker News
- [ ] Share on Reddit
- [ ] Tweet and tag relevant people

---

## 🎯 Success Metrics

### MVP Validation (Month 3)
- ✅ 100 users
- ✅ 10+ GitHub issues/discussions
- ✅ 2-3 external contributors
- ✅ Proof people find it useful

### Growth (Month 6)
- ✅ 1,000 users
- ✅ 500 GitHub stars
- ✅ 10+ sources supported
- ✅ Mentioned in blogs/podcasts

### Monetization (Month 12)
- ✅ 5,000 users
- ✅ 10 paying teams
- ✅ $500-1,000 MRR
- ✅ Clear product-market fit

### Long-term (18+ months)
- 🎯 10,000+ users
- 🎯 50-100 paying teams
- 🎯 $5,000-10,000 MRR
- 🎯 Sustainable side income or full-time

---

## 🤔 Key Risks & Mitigations

### Risk: Low Adoption

**Mitigation:**
- Make onboarding EXTREMELY easy
- Support popular sources from day 1
- Integrate with existing tools (llm, fabric)
- Show value in first 5 minutes

### Risk: Source Format Fragmentation

**Mitigation:**
- Start with 2-3 formats
- Plugin architecture for community parsers
- Document format spec clearly

### Risk: Execution Backend Complexity

**Mitigation:**
- Don't reinvent the wheel - delegate to existing tools
- Let users configure their own backend
- Focus on organization, not execution

### Risk: Competition from Fabric/llm

**Mitigation:**
- Position as complementary, not competitive
- Integrate deeply with their tools
- Focus on multi-source aggregation (our unique value)

### Risk: Monetization Difficulty

**Mitigation:**
- Free tier is permanently valuable
- Build for yourself first
- Don't count on revenue for 12+ months
- Multiple monetization paths (individual, team, enterprise)

---

## 📝 Key Decisions Log

### 2025-12-23: Project Conception & Strategic Direction

**Core Concept:**
- ✅ **Vision:** Bookmark/package manager for prompts (NOT prompt creator)
- ✅ **Name:** pkit (chosen for clarity and readability over quiver)
- ✅ **Domain:** pkit.ai (available)
- ✅ **Strategic Positioning:** Organization tool, NOT execution engine

**Critical Decision: Three-Phase Approach**
- ✅ **Phase 1 (Months 1-3):** Pure bookmark manager - NO EXECUTION
  - Only outputs prompt text via `pkit get`
  - Users pipe to their preferred tool (llm, fabric, mods, etc.)
  - Validates core value: multi-source organization
  - Success metric: 100 users who find bookmark-only valuable

- ✅ **Phase 2 (Months 4-6):** Optional thin wrapper - CONVENIENCE ONLY
  - Add `pkit run` as thin wrapper (if Phase 1 users request it)
  - Still just pipes to configured backend
  - Core value remains organization
  - Success metric: Validate if execution wrapper adds value

- ✅ **Phase 3 (Months 7-12):** Unique features - DIFFERENTIATION
  - Cross-source composition
  - Version locking and tracking
  - Team configurations
  - Features that Fabric/llm/mods DON'T have

**Why This Matters:**
- Avoids competing with Fabric (23k stars, established)
- Focuses on unique value (multi-source aggregation)
- Validates organization-only is valuable before adding execution
- Keeps tool simple and focused

**Technical Decisions:**
- ✅ **Tech Stack:** Go with Cobra (standard CLI framework)
- ✅ **UI Strategy:** Simple CLI + Web (NOT TUI)
  - Phase 1: Text-based CLI only (2-3 weeks to MVP)
  - Phase 2: Add Web UI for browsing (local-first)
  - Phase 3: Cloud sync via Web
- ✅ **Why not TUI (Bubble Tea)?**
  - Web is better for browsing/organizing
  - TUI delays MVP by 6+ weeks
  - Need Web anyway for team features
  - Simple CLI works better with Unix pipes
- ✅ **Repo Strategy:** Start personal, move to org after 500+ stars
- ✅ **Primary Sources:** Fabric patterns, awesome-chatgpt-prompts
- ✅ **MVP Timeline:** 2-3 weeks to Phase 1 launch

**Positioning in Ecosystem:**
- ✅ Complementary to Fabric/llm/mods (not competitive)
- ✅ Tool-agnostic (works with ANY execution backend)
- ✅ "Package manager for prompts" (like Homebrew, npm, Pocket)
- ✅ CLI for power users, Web for visual tasks

---

## 🔗 Important Links

### Prompt Sources (To Aggregate)

- [Fabric Patterns](https://github.com/danielmiessler/fabric) - 300+ curated prompt patterns
- [awesome-chatgpt-prompts](https://github.com/f/awesome-chatgpt-prompts) - Popular prompt collection
- [Awesome Prompts](https://github.com/ai-boost/awesome-prompts) - Collection of prompts
- [ChatGPT System Prompts](https://github.com/LouisShark/chatgpt_system_prompt) - System prompts

### Execution Tools (pkit works WITH these)

- [Fabric](https://github.com/danielmiessler/fabric) - 23k+ stars, 300+ patterns, Unix pipes
- [llm](https://github.com/simonw/llm) - Simon Willison's multi-provider CLI
- [mods](https://github.com/charmbracelet/mods) - Charm's AI CLI with MCP support
- [shell-gpt](https://github.com/TheR1D/shell_gpt) - AI assistant with roles/templates
- [AIChat](https://github.com/sigoden/aichat) - All-in-one LLM CLI
- [chatblade](https://github.com/npiv/chatblade) - Swiss Army Knife for ChatGPT

### Testing/Evaluation Tools

- [promptfoo](https://github.com/promptfoo/promptfoo) - Test and compare prompts, CI/CD

### Related Tools

- [Elia](https://github.com/darrenburns/elia) - Terminal UI for LLM chat
- [Model Context Protocol](https://docs.claude.com/en/docs/mcp) - For future Claude Code integration

### Development Resources

**Go CLI Frameworks:**
- [Cobra](https://github.com/spf13/cobra) - CLI framework (used by kubectl, hugo, github)
- [Viper](https://github.com/spf13/viper) - Configuration management
- [go-git](https://github.com/go-git/go-git) - Git operations in Go

**Terminal Output:**
- [pterm](https://github.com/pterm/pterm) - Pretty terminal output
- [tablewriter](https://github.com/olekukonko/tablewriter) - ASCII table rendering
- [color](https://github.com/fatih/color) - Colorized output

**Web UI (Phase 2):**
- Consider: React, Vue, or Svelte for frontend
- Embed UI in Go binary using [embed](https://pkg.go.dev/embed)
- Local API server with [chi](https://github.com/go-chi/chi) or standard lib

**Other:**
- [pkit.ai domain](https://pkit.ai) - Register this ASAP
- [Model Context Protocol](https://docs.claude.com/en/docs/mcp) - For future integration

---

## 📞 Contact & Contributions

**Creator:** [Your Name]
**Status:** Pre-alpha, not accepting contributions yet
**Timeline:** Planning to launch MVP in Q1 2026

---

## 📄 License

*TBD - Consider MIT or Apache 2.0 for CLI tool*

---

**Last Updated:** 2025-12-23
**Next Review:** After MVP completion
