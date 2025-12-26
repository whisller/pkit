# Developer Quickstart Guide

**Feature Branch**: `001-phase1-mvp`
**Created**: 2025-12-26
**Purpose**: Developer setup and contribution guide for pkit Phase 1 MVP

## Overview

This guide helps developers set up their environment and start contributing to pkit. It covers local development, testing, and contribution workflow.

## Prerequisites

- **Go 1.23+** (latest stable)
- **Git** 2.x
- **Make** (optional but recommended)
- **GitHub account** (for testing authenticated features)
- Terminal with UTF-8 support

### Verify Prerequisites

```bash
$ go version
go version go1.23.0 darwin/arm64

$ git --version
git version 2.43.0
```

## Quick Start (5 minutes)

### 1. Clone Repository

```bash
$ git clone https://github.com/whisller/pkit.git
$ cd pkit
```

### 2. Install Dependencies

```bash
$ go mod download
```

### 3. Build Binary

```bash
$ go build -o bin/pkit cmd/pkit/main.go
```

### 4. Run Tests

```bash
$ go test ./...
```

### 5. Try It Out

```bash
$ ./bin/pkit version
pkit version 0.1.0-dev

$ ./bin/pkit subscribe fabric/patterns
✓ Subscribed to fabric/patterns
  Format: fabric_pattern
  Prompts: 287

$ ./bin/pkit find
# Opens interactive TUI browser
```

## Development Environment

### Directory Structure

```
pkit/
├── cmd/pkit/              # CLI entry point
│   ├── main.go            # Cobra root + shorthand resolution
│   ├── subscribe.go       # subscribe command
│   ├── find.go            # find command (Bubbletea TUI)
│   ├── get.go             # get command
│   ├── save.go            # save command
│   └── ...
├── internal/              # Private application code
│   ├── config/            # Configuration management
│   │   ├── config.go
│   │   └── validate.go
│   ├── source/            # Source management (git, indexing)
│   │   ├── manager.go
│   │   ├── git.go
│   │   └── format.go
│   ├── parser/            # Format-specific parsers
│   │   ├── parser.go      # Interface
│   │   ├── fabric.go      # Fabric pattern parser
│   │   ├── awesome.go     # awesome-chatgpt-prompts parser
│   │   └── markdown.go    # Generic markdown parser
│   ├── index/             # Search indexing (bleve)
│   │   ├── index.go
│   │   └── search.go
│   ├── bookmark/          # Bookmark management
│   │   ├── bookmark.go
│   │   ├── resolver.go    # Shorthand resolution
│   │   └── validate.go
│   ├── tui/               # Bubbletea TUI components
│   │   ├── finder.go      # Main find TUI
│   │   ├── forms.go       # Bookmark/tag forms
│   │   └── styles.go
│   ├── display/           # Output formatting
│   │   ├── table.go
│   │   └── json.go
│   └── ratelimit/         # GitHub API rate limiting
│       └── tracker.go
├── pkg/models/            # Public data models
│   ├── source.go
│   ├── prompt.go
│   ├── bookmark.go
│   └── config.go
├── tests/                 # Integration tests
│   ├── fixtures/          # Test data
│   └── e2e/               # End-to-end tests
├── specs/                 # Specifications
│   └── 001-phase1-mvp/
├── go.mod
├── go.sum
├── Makefile               # Build automation
└── README.md
```

### Key Dependencies

```go
module github.com/whisller/pkit

go 1.23

require (
    github.com/spf13/cobra v1.8.1              // CLI framework
    github.com/charmbracelet/bubbletea v1.2.4  // TUI framework
    github.com/charmbracelet/lipgloss v1.0.0   // TUI styling
    github.com/charmbracelet/bubbles v0.20.0   // TUI widgets
    github.com/go-git/go-git/v5 v5.12.0        // Git operations
    github.com/goccy/go-yaml v1.15.13          // YAML parsing
    github.com/zalando/go-keyring v0.2.6       // Secure token storage
    github.com/blevesearch/bleve/v2 v2.4.5     // Full-text search
    github.com/mattn/go-isatty v0.0.20         // TTY detection
    golang.org/x/sync v0.10.0                  // errgroup for concurrency
    golang.org/x/term v0.27.0                  // Terminal operations
)
```

**Note:** Use `go get -u` to update to latest versions.

## Development Workflow

### Branch Strategy

```bash
# Create feature branch from main
$ git checkout -b feature/my-feature main

# Make changes, commit
$ git add .
$ git commit -m "feat: add new feature"

# Push and create PR
$ git push origin feature/my-feature
```

### Code Style

Follow Go best practices (Constitution Principle VIII):

1. **Error Handling**: Explicit, wrapped with context
   ```go
   if err != nil {
       return fmt.Errorf("failed to load config: %w", err)
   }
   ```

2. **Package Organization**: cmd/, internal/, pkg/
   ```go
   // internal/ - private to pkit
   // pkg/ - public API (models)
   ```

3. **Interfaces**: Keep small, focused
   ```go
   type Parser interface {
       ParsePrompts(source *Source) ([]Prompt, error)
   }
   ```

4. **Testing**: Table-driven tests
   ```go
   func TestValidateAlias(t *testing.T) {
       tests := []struct {
           name    string
           alias   string
           wantErr bool
       }{
           {"valid", "review", false},
           {"uppercase", "Review", true},
       }
       // ...
   }
   ```

### Running Tests

```bash
# All tests
$ go test ./...

# With coverage
$ go test -cover ./...

# Specific package
$ go test ./internal/bookmark/

# Verbose output
$ go test -v ./...

# Integration tests only
$ go test -tags=integration ./tests/

# With race detector
$ go test -race ./...
```

### Building

```bash
# Development build
$ go build -o bin/pkit cmd/pkit/main.go

# With debug info
$ go build -gcflags="all=-N -l" -o bin/pkit cmd/pkit/main.go

# Production build (optimized, stripped)
$ go build -ldflags="-s -w" -o bin/pkit cmd/pkit/main.go

# Cross-compile
$ GOOS=linux GOARCH=amd64 go build -o bin/pkit-linux-amd64 cmd/pkit/main.go
$ GOOS=darwin GOARCH=arm64 go build -o bin/pkit-darwin-arm64 cmd/pkit/main.go
$ GOOS=windows GOARCH=amd64 go build -o bin/pkit-windows-amd64.exe cmd/pkit/main.go
```

### Makefile Targets

```makefile
.PHONY: build test install clean

build:
	@echo "Building pkit..."
	@go build -o bin/pkit cmd/pkit/main.go

test:
	@echo "Running tests..."
	@go test -v -cover ./...

install:
	@echo "Installing pkit..."
	@go install cmd/pkit/main.go

clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -rf ~/.pkit/

lint:
	@echo "Running linters..."
	@golangci-lint run

coverage:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out

fmt:
	@go fmt ./...
	@goimports -w .
```

## Testing Strategy

### Unit Tests

Located alongside source files (`*_test.go`).

```go
// internal/bookmark/validate_test.go
package bookmark

import "testing"

func TestValidateAlias(t *testing.T) {
    tests := []struct {
        name    string
        alias   string
        wantErr bool
    }{
        {"valid lowercase", "review", false},
        {"valid with hyphen", "code-review", false},
        {"valid with underscore", "code_review", false},
        {"invalid uppercase", "Review", true},
        {"invalid special char", "review!", true},
        {"invalid reserved", "search", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateAlias(tt.alias)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateAlias(%q) error = %v, wantErr %v",
                    tt.alias, err, tt.wantErr)
            }
        })
    }
}
```

### Integration Tests

Test end-to-end workflows.

```go
// tests/e2e/subscribe_test.go
//go:build integration

package e2e

import (
    "os"
    "testing"
)

func TestSubscribeWorkflow(t *testing.T) {
    // Setup test environment
    tmpDir := t.TempDir()
    os.Setenv("PKIT_HOME", tmpDir)

    // Subscribe to test source
    err := RunCommand("subscribe", "file://./fixtures/test-source")
    if err != nil {
        t.Fatalf("subscribe failed: %v", err)
    }

    // Verify source was added
    cfg, _ := LoadConfig(tmpDir)
    if len(cfg.Sources) != 1 {
        t.Errorf("expected 1 source, got %d", len(cfg.Sources))
    }

    // Verify prompts were indexed
    prompts, _ := SearchPrompts("")
    if len(prompts) != 5 {
        t.Errorf("expected 5 prompts, got %d", len(prompts))
    }
}
```

### Test Fixtures

```
tests/fixtures/
├── test-source/           # Git repo for testing
│   ├── .git/
│   ├── patterns/
│   │   ├── test-prompt-1/
│   │   │   └── system.md
│   │   └── test-prompt-2/
│   │       └── system.md
│   └── README.md
├── bookmarks.yml          # Test bookmarks
└── config.yml             # Test config
```

## Debugging

### Enable Verbose Logging

```bash
$ pkit subscribe fabric/patterns --verbose
→ Resolving source: fabric/patterns
→ Full URL: https://github.com/danielmiessler/fabric
→ Cloning repository...
...

$ pkit get review --debug
→ Resolving prompt: review
→ Found bookmark: review → fabric:code-review
→ Loading from: ~/.pkit/sources/fabric/patterns/code-review/system.md
→ Complete in 12ms
```

### Use Delve Debugger

```bash
# Install delve
$ go install github.com/go-delve/delve/cmd/dlv@latest

# Debug pkit
$ dlv debug cmd/pkit/main.go -- subscribe fabric/patterns

# Set breakpoints
(dlv) break internal/source/manager.go:42
(dlv) continue
```

### Test Isolation

Use `PKIT_HOME` environment variable to isolate test data:

```bash
# Create isolated test environment
$ export PKIT_HOME=/tmp/pkit-test
$ ./bin/pkit subscribe fabric/patterns

# Clean up
$ rm -rf /tmp/pkit-test
$ unset PKIT_HOME
```

## Common Development Tasks

### Add New Command

1. Create command file in `cmd/pkit/`:
   ```go
   // cmd/pkit/newcmd.go
   package main

   import (
       "github.com/spf13/cobra"
   )

   var newCmd = &cobra.Command{
       Use:   "newcmd",
       Short: "Description",
       RunE: func(cmd *cobra.Command, args []string) error {
           // Implementation
           return nil
       },
   }

   func init() {
       rootCmd.AddCommand(newCmd)
   }
   ```

2. Add tests:
   ```go
   // cmd/pkit/newcmd_test.go
   func TestNewCmd(t *testing.T) {
       // Test implementation
   }
   ```

3. Update contracts documentation

### Add New Parser

1. Implement `Parser` interface:
   ```go
   // internal/parser/myformat.go
   package parser

   type MyFormatParser struct{}

   func (p *MyFormatParser) ParsePrompts(source *Source) ([]Prompt, error) {
       // Implementation
       return prompts, nil
   }

   func (p *MyFormatParser) CanParse(sourcePath string) bool {
       // Detection logic
       return true
   }
   ```

2. Register in format detector:
   ```go
   // internal/source/format.go
   func DetectFormat(path string) string {
       if new(MyFormatParser).CanParse(path) {
           return "myformat"
       }
       // ...
   }
   ```

3. Add tests with fixtures

### Update Data Model

1. Modify structs in `pkg/models/`:
   ```go
   // pkg/models/bookmark.go
   type Bookmark struct {
       // ... existing fields
       NewField string `yaml:"new_field" json:"new_field"`
   }
   ```

2. Update validation logic
3. Add migration function if needed (post-stabilization)
4. Update tests

## Troubleshooting

### Build Errors

**Error:** `cannot find package`
```bash
$ go mod tidy
$ go mod download
```

**Error:** `version conflict`
```bash
$ go clean -modcache
$ go mod download
```

### Test Failures

**Error:** `permission denied on ~/.pkit`
```bash
# Use test isolation
$ export PKIT_HOME=/tmp/pkit-test
$ go test ./...
```

**Error:** `index not found`
```bash
# Clean test state
$ rm -rf /tmp/pkit-test
$ go test ./...
```

### Runtime Issues

**Error:** `GitHub API rate limit`
```bash
# Set up auth token
$ export GITHUB_TOKEN=ghp_your_token_here
$ pkit config set github.use_auth true
```

**Error:** `git: command not found`
```bash
# Install git
$ brew install git  # macOS
$ apt-get install git  # Linux
```

## Performance Testing

### Benchmark Indexing

```go
// internal/index/index_bench_test.go
func BenchmarkIndexSource(b *testing.B) {
    source := setupTestSource(300) // 300 prompts

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        IndexSource(source)
    }
}
```

```bash
$ go test -bench=. -benchmem ./internal/index/
BenchmarkIndexSource-8    10    120ms/op    45MB/op
```

### Profile Memory

```bash
$ go test -memprofile=mem.prof ./internal/index/
$ go tool pprof mem.prof
(pprof) top10
(pprof) list IndexSource
```

### Profile CPU

```bash
$ go test -cpuprofile=cpu.prof ./internal/index/
$ go tool pprof cpu.prof
(pprof) top10
(pprof) web  # Opens browser with call graph
```

## CI/CD

### GitHub Actions Workflow

```yaml
# .github/workflows/test.yml
name: Test

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Install dependencies
        run: go mod download

      - name: Run tests
        run: go test -v -cover ./...

      - name: Run linters
        uses: golangci/golangci-lint-action@v6
        with:
          version: latest
```

## Release Process

### Version Tagging

```bash
# Create release tag
$ git tag -a v0.1.0 -m "Phase 1 MVP release"
$ git push origin v0.1.0
```

### GoReleaser

```yaml
# .goreleaser.yml
builds:
  - main: ./cmd/pkit
    binary: pkit
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64

archives:
  - format: tar.gz
    name_template: "pkit_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
```

```bash
$ goreleaser release --snapshot --clean
```

## Contributing

### Before Submitting PR

1. ✅ Tests pass: `go test ./...`
2. ✅ Linters pass: `golangci-lint run`
3. ✅ Code formatted: `go fmt ./...`
4. ✅ Documentation updated
5. ✅ Constitution compliance checked
6. ✅ Commit messages follow convention

### Commit Message Convention

```
<type>: <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `test`: Tests
- `refactor`: Code refactoring
- `perf`: Performance improvement
- `chore`: Build/tooling changes

**Example:**
```
feat: add interactive TUI browser

Implement pkit find command with Bubbletea TUI for real-time prompt
filtering and in-TUI bookmarking with tags.

Closes #42
```

## Resources

### Documentation

- [Constitution](../../.specify/memory/constitution.md) - Project principles
- [Specification](./spec.md) - Feature requirements
- [Plan](./plan.md) - Implementation plan
- [Data Model](./data-model.md) - Entity definitions
- [Contracts](./contracts/) - Command specifications

### External References

- [Go Documentation](https://go.dev/doc/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Cobra Guide](https://github.com/spf13/cobra)
- [Bubbletea Tutorial](https://github.com/charmbracelet/bubbletea/tree/master/tutorials)
- [Bleve Documentation](https://blevesearch.com/docs/)
- [go-git Examples](https://github.com/go-git/go-git/tree/master/_examples)

## Getting Help

- **Repository**: https://github.com/whisller/pkit
- **Issues**: https://github.com/whisller/pkit/issues
- **Discussions**: https://github.com/whisller/pkit/discussions

## License

See LICENSE file in repository root.
