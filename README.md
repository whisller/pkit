# pkit

**Multi-source AI prompt bookmark manager**

`pkit` is a CLI tool for subscribing to, organizing, and using AI prompts from multiple GitHub sources. It provides full-text search, interactive browsing, and seamless piping to execution tools like `claude`, `llm`, `fabric`, and `mods`.

**Status:** Phase 1 MVP - In Development
**Repository:** https://github.com/whisller/pkit

## Features

- **Multi-source aggregation**: Subscribe to GitHub repositories (Fabric, awesome-chatgpt-prompts, etc.)
- **Full-text search**: Fast bleve-powered indexing and search
- **Interactive TUI**: Real-time filtering with Bubbletea
- **Bookmarking**: Save prompts with custom aliases and tags
- **Unix pipe-friendly**: Clean output protocol for piping to execution tools
- **Tool-agnostic**: Works with claude, llm, fabric, mods, and any CLI tool

## Quick Start

### Installation

```bash
# Clone repository
git clone https://github.com/whisller/pkit.git
cd pkit

# Build binary
make build

# Install to GOPATH/bin
make install
```

### Basic Usage

```bash
# Subscribe to a prompt source
pkit subscribe fabric/patterns

# Interactive browser
pkit find

# Search for prompts
pkit search "code review"

# Bookmark a prompt
pkit save fabric:code-review --as review --tags dev,security

# Get prompt content (pipe to execution tools)
pkit get review | claude -p "analyse me ~/main.go"

# Shorthand form
pkit review | claude -p "analyse me ~/main.go"
```

## What Problem Does This Solve?

**Current workflow** requires managing multiple repositories manually:
```bash
cd ~/repos/fabric && git pull
cd ~/repos/awesome-chatgpt-prompts && git pull
grep -r "summarize" ../fabric ../awesome-chatgpt-prompts
# Copy-paste prompt manually
```

**With pkit:**
```bash
pkit subscribe fabric/patterns f/awesome-chatgpt-prompts
pkit search summarize
pkit save fabric:summarize --as sum
cat article.txt | pkit sum | llm -m claude-3-sonnet
```

## Project Status

**Phase 1 MVP** - In Development

Current implementation:
- ✅ Project setup (Go 1.23+, dependencies, CI)
- 🚧 Foundation (config, models, git operations)
- 🚧 Subscribe & discover (parsers, indexing, search)
- 🚧 Pipe to tools (get command, shorthand resolution)
- 🚧 Bookmark & organize (save, tags, aliases)
- 🚧 Track updates (upgrade sources)
- 🚧 Polish (interactive TUI)

## Development

### Prerequisites

- Go 1.23+
- Git 2.x
- Make (optional)

### Build & Test

```bash
# Build
make build

# Run tests
make test

# Run linters
make lint

# Generate coverage report
make coverage

# Clean artifacts
make clean
```

### Project Structure

```
pkit/
├── cmd/pkit/              # CLI entry point
├── internal/              # Private application code
│   ├── config/            # Configuration management
│   ├── source/            # Source management (git, indexing)
│   ├── parser/            # Format-specific parsers
│   ├── index/             # Search indexing (bleve)
│   ├── bookmark/          # Bookmark management
│   ├── tui/               # Interactive TUI (Bubbletea)
│   └── display/           # Output formatting
├── pkg/models/            # Public data models
├── specs/                 # Specifications
└── tests/                 # Integration tests
```

## Documentation

- [Specification](specs/001-phase1-mvp/spec.md) - Feature requirements
- [Plan](specs/001-phase1-mvp/plan.md) - Implementation plan
- [Tasks](specs/001-phase1-mvp/tasks.md) - Implementation tasks
- [Quickstart](specs/001-phase1-mvp/quickstart.md) - Developer guide
- [Contracts](specs/001-phase1-mvp/contracts/) - Command specifications

## License

See LICENSE file.

## Contributing

This project is in early development (Phase 1 MVP). Contributions welcome once core features are stable.

For development workflow and coding standards, see [quickstart.md](specs/001-phase1-mvp/quickstart.md).
