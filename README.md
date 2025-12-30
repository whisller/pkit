<div align="center">
    <h1>pkit</h1>
    <p><strong>Multi-source AI prompt bookmark manager</strong></p>
    <p>
        <a href="#features">Features</a> •
        <a href="#quick-start">Quick Start</a> •
        <a href="#installation">Installation</a> •
        <a href="#usage">Usage</a> •
        <a href="#configuration">Configuration</a>
    </p>
</div>

`pkit` is a CLI tool for subscribing to, organizing, and using AI prompts from multiple GitHub sources. It provides full-text search, interactive browsing, and seamless piping to execution tools like `claude`, `llm`, `fabric`, and `mods`.

**Status:** WIP - Phase 1 MVP in active development

---

<div align="center">
    <img src="docs/gifs/terminal.gif" alt="pkit demo" width="240"/>
    <img src="docs/gifs/web.gif" alt="pkit demo" width="240"/>
    <p><em>Interactive TUI and Web for browsing and searching prompts</em></p>
</div>

---

## Features

- **Multi-source aggregation**: Subscribe to GitHub repositories (Fabric, awesome-chatgpt-prompts, etc.)
- **Full-text search**: Fast bleve-powered indexing and search
- **Interactive TUI**: Real-time filtering with Bubbletea
- **Web Interface**: Browse prompts in your browser with a clean, responsive UI
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

# Or download pre-built binary from releases
# https://github.com/whisller/pkit/releases
```

## Usage

### Basic Usage

```bash
# Subscribe to a prompt source
pkit subscribe fabric/patterns

# or multiple sources at once
pkit subscribe fabric/patterns f/awesome-chatgpt-prompts

# Interactive browser (TUI) where you can interactively do all actions
pkit find

# Start web server, where you can use your browser to manage your prompts
pkit serve --port 8080
# Open http://127.0.0.1:8080 in your browser

# Search for prompts
pkit search "code review"

# Bookmark a prompt
pkit save fabric:code-review --as review --tags dev,security

# Get prompt content (pipe to execution tools)
pkit get review | claude -p "analyse me ~/main.go"

# Shorthand form
pkit review | claude -p "analyse me ~/main.go"
```


#### Bookmarks and Tags
```bash
# Add bookmark with alias and tags
pkit bookmark add fabric:code-review --alias review --tags dev,security

# List all bookmarks
pkit bookmark list

# Tag an existing prompt
pkit tag add fabric:summarize productivity,writing

# Search by tags
pkit search --tags dev,security
```

#### Web Interface
```bash
# Start web server on custom port
pkit serve --port 3000 --host 0.0.0.0

# Access at http://localhost:3000
# Features:
# - Browse all prompts with filters
# - Search in real-time
# - Manage bookmarks and tags
# - Copy prompt content with one click
```

#### Piping to AI Tools
```bash
# With Claude CLI
pkit get fabric:summarize | claude < article.txt

# With Simon Willison's llm
cat code.go | pkit review | llm -m claude-3-opus

# With fabric
pkit get fabric:extract-wisdom | fabric --stream

# With mods
echo "Explain Docker" | pkit get teacher | mods
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

## Configuration

`pkit` stores all data in `~/.pkit/`:

```
~/.pkit/
├── config.yml          # Main configuration
├── bookmarks.yml       # Bookmarked prompts
├── aliases.yml         # Custom aliases
├── tags.yml           # Prompt tags
├── sources/           # Cloned repositories
│   ├── fabric/
│   └── awesome-chatgpt/
└── index/             # Bleve search index
    └── prompts.bleve/
```

### Configuration File

Default `~/.pkit/config.yml`:

```yaml
sources:
  - id: fabric
    url: https://github.com/danielmiessler/fabric
    format: fabric_pattern
    prompt_count: 150
    commit_sha: abc123...
    last_indexed: 2024-01-15T10:30:00Z

display:
  table_style: rounded    # simple, rounded, unicode
  date_format: relative   # relative, rfc3339, short

search:
  default_max_results: 50
```

## Architecture

### How It Works

1. **Subscribe**: Clone GitHub repositories to `~/.pkit/sources/`
2. **Parse**: Extract prompts based on repository format (Fabric patterns, awesome-chatgpt CSV, etc.)
3. **Index**: Build full-text search index using Bleve
4. **Search**: Query indexed prompts with fuzzy matching
5. **Bookmark**: Save frequently used prompts with aliases and tags
6. **Execute**: Pipe prompt content to AI tools

### Supported Sources

| Source | Format | Example |
|--------|--------|---------|
| [Fabric](https://github.com/danielmiessler/fabric) | Markdown patterns | `pkit subscribe fabric/patterns` |
| [awesome-chatgpt-prompts](https://github.com/f/awesome-chatgpt-prompts) | CSV | `pkit subscribe f/awesome-chatgpt-prompts` |
| Custom Markdown | Frontmatter-based | Any GitHub repo with markdown files |

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
