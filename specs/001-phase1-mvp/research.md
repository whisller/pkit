# Research Findings: pkit Phase 1 MVP

**Date**: 2025-12-25
**Phase**: 0 - Technical Research
**Purpose**: Document technology decisions, patterns, and implementation strategies

---

## Topic 1: Git Operations with go-git

**Decision**: Use go-git v5 for all Git operations without requiring system Git installation

**Rationale**:
- Single binary distribution without external dependencies
- Cross-platform consistency (no git version differences)
- Programmatic control over authentication and progress
- Well-maintained library used by many Go projects
- Supports authentication via HTTPS with tokens

**Alternatives Considered**:
- **Shell out to system git**: Rejected - requires git installation, harder to control authentication, inconsistent across platforms
- **libgit2 bindings**: Rejected - requires C dependencies, complicates cross-compilation

**Implementation Notes**:
```go
// Clone with authentication
import (
    "github.com/go-git/go-git/v5"
    "github.com/go-git/go-git/v5/plumbing/transport/http"
)

// Clone repository
_, err := git.PlainClone(localPath, false, &git.CloneOptions{
    URL: repoURL,
    Auth: &http.BasicAuth{
        Username: "x-access-token", // Can be anything for GitHub
        Password: token,             // GitHub token from keyring
    },
    Progress: os.Stderr,             // Show progress to stderr
})

// Check for updates
repo, _ := git.PlainOpen(localPath)
remote, _ := repo.Remote("origin")
refs, _ := remote.List(&git.ListOptions{Auth: auth})
// Compare with local refs to detect updates
```

**Security Considerations**:
- Never log or display tokens
- Use context.Context for timeout/cancellation
- Handle auth errors gracefully (don't expose token in error messages)

---

## Topic 2: Bleve Search Indexing

**Decision**: Use bleve v2 for full-text search with persistent disk-based index

**Rationale**:
- Pure Go implementation (no C dependencies)
- Supports fuzzy matching and highlighting
- Persistent index with incremental updates
- Fast search (<1s across 10k+ documents)
- Built-in analyzers for text processing

**Alternatives Considered**:
- **In-memory search**: Rejected - doesn't scale, slow startup on large prompt sets
- **SQLite FTS**: Rejected - adds SQL complexity, less Go-idiomatic
- **External search (Elasticsearch)**: Rejected - violates single-binary constraint

**Implementation Notes**:
```go
import (
    "github.com/blevesearch/bleve/v2"
)

// Index structure
type PromptDocument struct {
    ID          string   // source:name
    Name        string
    Description string
    Content     string
    Source      string
    Tags        []string
    Type        string   // Document type for bleve
}

// Create index
mapping := bleve.NewIndexMapping()
index, err := bleve.New(indexPath, mapping)

// Index a prompt
doc := PromptDocument{
    ID:          "fabric:code-review",
    Name:        "code-review",
    Description: "Review code for best practices",
    Content:     promptText,
    Source:      "fabric",
    Tags:        []string{"code", "review"},
    Type:        "prompt",
}
index.Index(doc.ID, doc)

// Search
query := bleve.NewMatchQuery("code review")
searchRequest := bleve.NewSearchRequest(query)
searchResult, _ := index.Search(searchRequest)
```

**Performance Considerations**:
- Batch indexing during subscribe (index multiple prompts in transaction)
- Incremental updates on upgrade (only re-index changed prompts)
- Store index in ~/.pkit/cache/index.bleve
- Use keyword analyzer for prompt IDs, text analyzer for content

---

## Topic 3: GitHub API Rate Limiting Strategy

**Decision**: Hybrid approach - use Git clone for content, GitHub API only for metadata when needed

**Rationale**:
- Git clone operations don't count against API rate limit
- API only needed for repository metadata (stars, description) - optional for Phase 1
- 60 requests/hour sufficient for checking rate limit status itself
- Token-based auth gives 5000 req/h if user configures it

**Alternatives Considered**:
- **API-only approach**: Rejected - hits rate limit quickly with multiple sources
- **No API at all**: Rejected - can't track rate limit status or show warnings

**Implementation Notes**:
```go
// Rate limit tracking structure
type RateLimitState struct {
    Remaining int
    Limit     int
    ResetAt   time.Time
}

// Check rate limit (minimal API usage)
func (g *GitHubClient) CheckRateLimit() (*RateLimitState, error) {
    // Only call API when needed (e.g., subscribe/update operations)
    // Cache result for 5 minutes to avoid frequent checks
}

// Warn at 80% consumption
if state.Remaining < int(float64(state.Limit)*0.2) {
    fmt.Fprintf(os.Stderr, "WARNING: GitHub rate limit low (%d/%d remaining)\n", state.Remaining, state.Limit)
    fmt.Fprintf(os.Stderr, "Configure token with: export GITHUB_TOKEN=<token>\n")
}
```

**Rate Limit Handling**:
- Track rate limit in ~/.pkit/cache/ratelimit.json
- Display warning at 80% consumption
- Clear instructions to add GitHub token for higher limits
- Never block operations due to rate limit (git clone still works)

---

## Topic 4: Secure Token Storage with zalando/go-keyring

**Decision**: Use zalando/go-keyring with environment variable fallback

**Rationale**:
- Cross-platform support (macOS Keychain, Windows Credential Manager, Linux Secret Service)
- Simple API: `keyring.Set(service, user, password)` / `keyring.Get(service, user)`
- Fallback to GITHUB_TOKEN environment variable for CI/containers
- Well-maintained library

**Alternatives Considered**:
- **Platform-specific implementations**: Rejected - increases complexity, hard to maintain
- **Plain text config**: Rejected - security risk
- **go-keyring (different library)**: Rejected - zalando/go-keyring more actively maintained

**Implementation Notes**:
```go
import "github.com/zalando/go-keyring"

const (
    serviceName = "pkit"
    accountName = "github-token"
)

// Store token
func StoreGitHubToken(token string) error {
    return keyring.Set(serviceName, accountName, token)
}

// Retrieve token
func GetGitHubToken() (string, error) {
    // Try keyring first
    token, err := keyring.Get(serviceName, accountName)
    if err == nil {
        return token, nil
    }

    // Fallback to environment variable
    if envToken := os.Getenv("GITHUB_TOKEN"); envToken != "" {
        return envToken, nil
    }

    return "", errors.New("no GitHub token found (keyring or GITHUB_TOKEN env)")
}

// Delete token
func DeleteGitHubToken() error {
    return keyring.Delete(serviceName, accountName)
}
```

**Error Handling**:
- Gracefully handle keyring unavailable (e.g., headless Linux)
- Clear error messages pointing to environment variable fallback
- Token optional for unauthenticated access (60 req/h)

---

## Topic 5: Prompt Parser Security (Sandboxing)

**Decision**: Use goccy/go-yaml with strict unmarshaling, no code execution paths

**Rationale**:
- goccy/go-yaml is pure Go, no C dependencies
- Supports strict unmarshaling to prevent unexpected fields
- No eval/exec paths in parser code
- Markdown parsing uses text-only, no HTML rendering
- CSV parsing uses standard library encoding/csv

**Alternatives Considered**:
- **gopkg.in/yaml.v3**: Rejected - slower, less feature-complete than goccy
- **Allow HTML in Markdown**: Rejected - security risk (XSS if ever displayed in web UI)

**Implementation Notes**:
```go
import (
    "github.com/goccy/go-yaml"
)

// Fabric format (Markdown with YAML frontmatter)
type FabricPrompt struct {
    Title       string   `yaml:"title"`
    Description string   `yaml:"description"`
    Tags        []string `yaml:"tags"`
    Content     string   // Everything after frontmatter
}

// Safe parsing
func ParseFabricPrompt(data []byte) (*FabricPrompt, error) {
    // Split frontmatter and content
    parts := bytes.SplitN(data, []byte("---"), 3)
    if len(parts) < 3 {
        return nil, errors.New("invalid frontmatter")
    }

    var prompt FabricPrompt
    if err := yaml.UnmarshalWithOptions(parts[1], &prompt, yaml.Strict()); err != nil {
        return nil, fmt.Errorf("invalid frontmatter: %w", err)
    }

    prompt.Content = string(bytes.TrimSpace(parts[2]))
    return &prompt, nil
}
```

**Security Rules**:
- NEVER execute code from prompts
- NEVER use `eval`, `exec`, or dynamic imports
- Only parse recognized formats (Markdown, YAML, CSV)
- Skip files with executable extensions (.sh, .py, .js, etc.)
- Warn if repository contains git hooks or executable scripts
- Validate all input before indexing

**File Type Whitelist**:
- `.md` - Markdown
- `.yaml`, `.yml` - YAML
- `.csv` - CSV
- `.txt` - Plain text
- All others: skip with debug log

---

## Topic 6: Bubbletea Interactive Finder UX

**Decision**: Use Bubbletea with bubbles/list component and real-time filtering

**Rationale**:
- Bubbletea is the standard Go TUI framework (used by gh, glow, soft-serve)
- Bubbles provides pre-built list, textinput, viewport components
- Real-time updates via Elm architecture (Model-Update-View)
- TTY detection with isatty prevents activation when piped
- Clean fallback to traditional output when not in TTY

**Alternatives Considered**:
- **termui**: Rejected - less actively maintained, more complex
- **tview**: Rejected - different paradigm, less composable
- **Custom TUI**: Rejected - reinventing the wheel

**Implementation Notes**:
```go
import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/bubbles/list"
    "github.com/charmbracelet/bubbles/textinput"
    "github.com/mattn/go-isatty"
)

// Model
type finderModel struct {
    list       list.Model
    textInput  textinput.Model
    prompts    []Prompt
    filtered   []Prompt
    selected   *Prompt
}

// Update (handle input)
func (m finderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "enter":
            m.selected = m.filtered[m.list.Index()]
            return m, tea.Quit
        case "esc", "ctrl+c":
            return m, tea.Quit
        }
    }
    // Update textinput, filter list based on input
    // ...
}

// View (render)
func (m finderModel) View() string {
    return lipgloss.JoinVertical(
        lipgloss.Top,
        m.textInput.View(),
        m.list.View(),
    )
}

// Launch interactive finder
func RunInteractiveFinder(prompts []Prompt) (*Prompt, error) {
    // Check if TTY
    if !isatty.IsTerminal(os.Stdout.Fd()) {
        // Fallback to traditional output
        return nil, errors.New("not a TTY")
    }

    // Run Bubbletea program
    model := initialModel(prompts)
    p := tea.NewProgram(model)
    finalModel, err := p.Run()
    return finalModel.(finderModel).selected, err
}
```

**Keybindings**:
- `↑/↓` or `j/k`: Navigate
- `Enter`: Select
- `Ctrl-G`: Select and execute get
- `Ctrl-P`: Toggle preview
- `Esc` or `Ctrl-C`: Exit
- `/`: Focus search input

**Performance**:
- Filter prompts on keystroke (<50ms target)
- Use fuzzy matching algorithm (simple substring or levenshtein)
- Limit displayed results to 50 items (pagination)
- Lazy load preview pane content

---

## Topic 7: Parallel Operations with Goroutines

**Decision**: Use goroutine pool with errgroup for parallel source operations

**Rationale**:
- Go's goroutines are perfect for I/O-bound git operations
- errgroup provides error aggregation and context cancellation
- Channel-based progress reporting to display per-source status
- Limit concurrency to avoid overwhelming system (10 concurrent max)

**Alternatives Considered**:
- **Sequential operations**: Rejected - too slow for multiple sources
- **Unlimited goroutines**: Rejected - can overwhelm system with many sources
- **Worker pool library**: Rejected - errgroup sufficient for our needs

**Implementation Notes**:
```go
import (
    "golang.org/x/sync/errgroup"
    "context"
)

// Parallel subscribe
func SubscribeParallel(sources []string) error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    g, ctx := errgroup.WithContext(ctx)
    g.SetLimit(10) // Max 10 concurrent operations

    progressCh := make(chan SourceProgress, len(sources))

    // Launch goroutine per source
    for _, source := range sources {
        source := source // Capture loop variable
        g.Go(func() error {
            return subscribeSource(ctx, source, progressCh)
        })
    }

    // Progress reporter in separate goroutine
    go func() {
        for progress := range progressCh {
            fmt.Fprintf(os.Stderr, "[%s] %s\n", progress.Source, progress.Message)
        }
    }()

    // Wait for all to complete
    err := g.Wait()
    close(progressCh)
    return err
}

type SourceProgress struct {
    Source  string
    Message string
    Done    bool
}
```

**Error Handling**:
- First error cancels all remaining operations (errgroup behavior)
- Collect all errors and report which sources failed
- Allow partial success (some sources indexed, others failed)
- Retry logic for transient network errors

**Progress Reporting**:
- Per-source progress: "Cloning...", "Indexing...", "Complete"
- Update stderr without interfering with stdout
- Use ANSI codes for in-place updates (optional, fallback to line-by-line)

---

## Topic 8: Cross-Platform Binary Distribution

**Decision**: Use GoReleaser for automated builds and releases

**Rationale**:
- Industry standard for Go binary releases
- Handles cross-compilation for macOS/Linux/Windows
- Generates checksums and signatures
- Creates GitHub releases automatically
- Supports Homebrew tap generation

**Alternatives Considered**:
- **Manual builds**: Rejected - error-prone, time-consuming
- **GitHub Actions only**: Rejected - reinventing what GoReleaser does
- **Docker**: Rejected - not suitable for CLI tool distribution

**Implementation Notes**:

`.goreleaser.yml`:
```yaml
project_name: pkit

builds:
  - env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    binary: pkit
    ldflags:
      - -s -w -X main.version={{.Version}}

archives:
  - format: tar.gz
    name_template: >-
      {{ .ProjectName }}_
      {{- .Version }}_
      {{- .Os }}_
      {{- .Arch }}
    format_overrides:
      - goos: windows
        format: zip

checksum:
  name_template: 'checksums.txt'

changelog:
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'

brews:
  - name: pkit
    repository:
      owner: yourorg
      name: homebrew-tap
    folder: Formula
    description: "Bookmark manager for AI prompts"
    homepage: "https://github.com/yourorg/pkit"
    install: |
      bin.install "pkit"
```

**Build Process**:
1. Tag release: `git tag v0.1.0`
2. Push tag: `git push origin v0.1.0`
3. GitHub Actions runs GoReleaser
4. Binaries published to GitHub Releases
5. Homebrew formula updated automatically

**Binary Size Optimization**:
- Use `-ldflags="-s -w"` to strip debug info
- CGO_ENABLED=0 for static binaries
- Target <20MB (easily achievable for Go CLIs)

---

## Topic 9: Shorthand Command Resolution Pattern

**Decision**: Use Cobra's `SilenceErrors` with custom `RunE` on root command

**Rationale**:
- Cobra allows custom handling of unknown commands
- Check if arg matches bookmark alias or prompt ID before showing error
- Preserve all explicit command behavior
- Fast lookup using hash map (<10ms)

**Alternatives Considered**:
- **Cobra aliases**: Rejected - can't dynamically generate all aliases
- **Custom command parsing**: Rejected - loses Cobra's features
- **Preprocess args**: Rejected - breaks Cobra's help system

**Implementation Notes**:
```go
// Root command with custom error handling
var rootCmd = &cobra.Command{
    Use:          "pkit",
    SilenceErrors: true,
    RunE: func(cmd *cobra.Command, args []string) error {
        if len(args) == 0 {
            return cmd.Help()
        }

        // Try to resolve as bookmark/prompt
        promptID := args[0]
        if exists, err := bookmarkExists(promptID); err == nil && exists {
            // Execute get command
            return getCommand.RunE(cmd, args)
        }

        // Unknown command
        return fmt.Errorf("unknown command or prompt: %q\nRun 'pkit --help' for usage", args[0])
    },
}

// Fast bookmark lookup
func bookmarkExists(alias string) (bool, error) {
    // Load bookmarks (cached in memory)
    bookmarks, err := loadBookmarks()
    if err != nil {
        return false, err
    }

    // Hash map lookup O(1)
    _, exists := bookmarks[alias]
    return exists, nil
}
```

**Performance**:
- Cache bookmarks.yml in memory (reload only when modified)
- Hash map lookup is O(1)
- File mtime check to detect changes
- Total resolution time <10ms

**Edge Cases**:
- Bookmark name conflicts with command: Explicit command wins
- Empty args: Show help
- Invalid bookmark: Clear error message
- Multiple args: Pass through to get command

---

## Summary

All technical decisions documented and ready for implementation. Key technologies:
- **go-git v5** - Git operations
- **bleve v2** - Search indexing
- **goccy/go-yaml** - YAML parsing
- **zalando/go-keyring** - Token storage
- **Bubbletea** - Interactive TUI
- **errgroup** - Parallel operations
- **GoReleaser** - Binary distribution

All choices prioritize:
1. **Single binary** distribution
2. **Cross-platform** compatibility
3. **Security** (no code execution, secure token storage)
4. **Performance** (<1s search, <100ms get, parallel ops)
5. **Go best practices** (idiomatic patterns, proper error handling)

Ready to proceed to Phase 1: Design & Contracts.
