# Technical Research: Local Web Server Implementation for pkit

**Date**: 2025-12-28
**Feature**: 003-web-interface
**Status**: Complete

This document contains detailed research and recommendations for implementing a local web server in Go for the pkit project. Each research question includes: Decision (what to use), Rationale (why chosen), and Alternatives considered.

---

## R001: Go HTTP Server Patterns for Single Binary Distribution

**Question**: What's the best practice for embedding and serving static assets (HTML/CSS/JS) in a Go binary for cross-platform distribution?

### Decision

Use Go's built-in `embed` package (Go 1.16+) with `embed.FS` for production builds, combined with a conditional file system that switches to disk-based loading during development for hot reloading.

### Implementation Pattern

```go
package server

import (
    "embed"
    "io/fs"
    "net/http"
    "os"
)

//go:embed static/*
var embeddedAssets embed.FS

//go:embed templates/*
var embeddedTemplates embed.FS

// GetAssetFS returns the appropriate filesystem for assets
// based on the environment (development vs production)
func GetAssetFS(devMode bool) (http.FileSystem, error) {
    if devMode {
        // Development: serve from disk for hot reloading
        return http.Dir("./internal/web/static"), nil
    }

    // Production: serve from embedded files
    staticFS, err := fs.Sub(embeddedAssets, "static")
    if err != nil {
        return nil, err
    }
    return http.FS(staticFS), nil
}

// GetTemplateFS returns the appropriate filesystem for templates
func GetTemplateFS(devMode bool) (fs.FS, error) {
    if devMode {
        // Development: load from disk
        return os.DirFS("./internal/web/templates"), nil
    }

    // Production: use embedded templates
    return fs.Sub(embeddedTemplates, "templates")
}
```

### Usage with HTTP Server

```go
package server

import (
    "html/template"
    "net/http"
)

type Server struct {
    templates *template.Template
    devMode   bool
}

func NewServer(devMode bool) (*Server, error) {
    s := &Server{devMode: devMode}

    // Load templates
    templateFS, err := GetTemplateFS(devMode)
    if err != nil {
        return nil, err
    }

    // Parse all templates with inheritance support
    s.templates, err = template.ParseFS(templateFS,
        "base.html",
        "partials/*.html",
        "pages/*.html",
    )
    if err != nil {
        return nil, err
    }

    return s, nil
}

func (s *Server) RegisterRoutes(mux *http.ServeMux) error {
    // Serve static assets
    assetFS, err := GetAssetFS(s.devMode)
    if err != nil {
        return err
    }

    mux.Handle("/static/", http.StripPrefix("/static/",
        http.FileServer(assetFS)))

    // Register page handlers
    mux.HandleFunc("/", s.handleHome)
    mux.HandleFunc("/prompt/", s.handlePromptDetail)

    return nil
}
```

### Template Structure

Organize templates in a hierarchical structure:

```
internal/web/templates/
├── base.html          # Base layout with common structure
├── partials/
│   ├── header.html    # Reusable header
│   ├── footer.html    # Reusable footer
│   ├── prompt-card.html   # Prompt list item component
│   └── filters.html   # Filter sidebar component
└── pages/
    ├── home.html      # Main prompt listing page
    └── detail.html    # Prompt detail page
```

### Asset Caching Strategy

```go
package server

import (
    "net/http"
    "time"
)

// CacheMiddleware adds appropriate cache headers
func CacheMiddleware(next http.Handler, maxAge time.Duration) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // For static assets, use aggressive caching
        w.Header().Set("Cache-Control",
            fmt.Sprintf("public, max-age=%d", int(maxAge.Seconds())))
        next.ServeHTTP(w, r)
    })
}

// Setup in RegisterRoutes
func (s *Server) RegisterRoutes(mux *http.ServeMux) error {
    assetFS, err := GetAssetFS(s.devMode)
    if err != nil {
        return err
    }

    // In production, cache static assets for 1 year
    cacheMaxAge := 365 * 24 * time.Hour
    if s.devMode {
        cacheMaxAge = 0 // No caching in dev mode
    }

    staticHandler := http.StripPrefix("/static/",
        http.FileServer(assetFS))
    mux.Handle("/static/",
        CacheMiddleware(staticHandler, cacheMaxAge))

    return nil
}
```

### Template Reload During Development

For development mode, reload templates on each request:

```go
func (s *Server) executeTemplate(w http.ResponseWriter, name string, data interface{}) error {
    if s.devMode {
        // Reload templates on each request in dev mode
        templateFS, err := GetTemplateFS(true)
        if err != nil {
            return err
        }

        tmpls, err := template.ParseFS(templateFS,
            "base.html",
            "partials/*.html",
            "pages/*.html",
        )
        if err != nil {
            return err
        }

        return tmpls.ExecuteTemplate(w, name, data)
    }

    // Production: use pre-parsed templates
    return s.templates.ExecuteTemplate(w, name, data)
}
```

### Rationale

1. **Native Solution**: `embed` package is built into Go 1.16+, eliminating external dependencies
2. **Single Binary**: All assets are compiled into the binary, making distribution trivial
3. **Development Flexibility**: Conditional file system allows hot reloading during development without affecting production builds
4. **Performance**: Embedded assets are loaded into memory once at startup, providing fast access
5. **Compatibility**: Works seamlessly with `html/template`, `http.FileServer`, and standard library packages
6. **Zero Runtime Dependencies**: No need for external file serving or template engines

### Alternatives Considered

#### 1. Third-party Embedding Libraries (packr, statik, go-bindata)

**Pros**:
- More features and tooling in some cases
- Some offer compression and preprocessing

**Cons**:
- External dependencies (violates lightweight requirement)
- Deprecated or archived (go-bindata, packr)
- Unnecessary complexity when `embed` provides native solution
- Additional build steps required

**Why Not Chosen**: The native `embed` package provides everything needed without external dependencies, which aligns with the project's minimal dependency constraint.

#### 2. Docker Volume Mounts / External Assets

**Pros**:
- Easy asset updates without recompilation
- Familiar to Docker users

**Cons**:
- Requires Docker for distribution (adds complexity)
- Multi-file deployment increases complexity
- Doesn't meet single-binary distribution requirement
- Breaks cross-platform portability guarantees

**Why Not Chosen**: Violates the core requirement of single binary distribution for cross-platform deployment.

#### 3. Asset Generation with go:generate

**Pros**:
- Generates Go code from assets at build time
- Can integrate preprocessing pipelines

**Cons**:
- More complex build process
- Generated Go files add to repository size
- Harder to maintain and debug
- Unnecessary when `embed` provides cleaner solution

**Why Not Chosen**: `embed` achieves the same result with cleaner code and no generated files in the repository.

---

## R002: Server-Side Rendering Architecture

**Question**: How should we structure HTML templates to minimize JavaScript while maintaining interactive filtering and search?

### Decision

Use Go's `html/template` with a block/inheritance pattern for layouts, combined with progressive enhancement: HTML forms with standard GET parameters for filters (works without JS), enhanced with minimal JavaScript for real-time filtering.

### Template Inheritance Pattern

```go
// base.html - The foundation layout
{{define "base"}}
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{block "title" .}}pkit - Prompt Library{{end}}</title>
    <link rel="stylesheet" href="/static/css/style.css">
</head>
<body>
    <header>
        {{template "header" .}}
    </header>

    <main>
        {{block "content" .}}
            <!-- Page-specific content goes here -->
        {{end}}
    </main>

    <footer>
        {{template "footer" .}}
    </footer>

    {{block "scripts" .}}
        <!-- Optional page-specific scripts -->
    {{end}}
</body>
</html>
{{end}}

// partials/header.html
{{define "header"}}
<nav class="navbar">
    <h1>pkit - Prompt Library</h1>
    <div class="stats">
        <span>{{.PromptCount}} prompts</span>
    </div>
</nav>
{{end}}
```

```go
// pages/home.html - Main listing page
{{template "base" .}}

{{define "title"}}Prompts - pkit{{end}}

{{define "content"}}
<div class="container">
    <aside class="sidebar">
        <form method="GET" action="/" id="filter-form">
            <div class="filter-section">
                <label for="search">Search:</label>
                <input
                    type="text"
                    id="search"
                    name="q"
                    value="{{.Query}}"
                    placeholder="Search prompts..."
                    autocomplete="off"
                >
            </div>

            <div class="filter-section">
                <label>Source:</label>
                <select name="source" id="source-filter">
                    <option value="">All Sources</option>
                    {{range .Sources}}
                    <option value="{{.ID}}" {{if eq $.SelectedSource .ID}}selected{{end}}>
                        {{.Name}} ({{.Count}})
                    </option>
                    {{end}}
                </select>
            </div>

            <div class="filter-section">
                <label>Tags:</label>
                {{range .Tags}}
                <label class="tag-checkbox">
                    <input
                        type="checkbox"
                        name="tags"
                        value="{{.Name}}"
                        {{if .Selected}}checked{{end}}
                    >
                    {{.Name}} ({{.Count}})
                </label>
                {{end}}
            </div>

            <div class="filter-section">
                <label class="tag-checkbox">
                    <input
                        type="checkbox"
                        name="bookmarked"
                        value="true"
                        {{if .ShowBookmarked}}checked{{end}}
                    >
                    Bookmarked Only
                </label>
            </div>

            <button type="submit">Apply Filters</button>
        </form>
    </aside>

    <section class="main-content">
        <div class="prompt-list" id="prompt-list">
            {{if .Prompts}}
                {{range .Prompts}}
                    {{template "prompt-card" .}}
                {{end}}
            {{else}}
                <div class="empty-state">
                    <p>No prompts found matching your filters.</p>
                </div>
            {{end}}
        </div>

        {{if .Pagination}}
        <nav class="pagination">
            {{if .Pagination.HasPrev}}
            <a href="?{{.Pagination.PrevURL}}">← Previous</a>
            {{end}}

            <span>Page {{.Pagination.CurrentPage}} of {{.Pagination.TotalPages}}</span>

            {{if .Pagination.HasNext}}
            <a href="?{{.Pagination.NextURL}}">Next →</a>
            {{end}}
        </nav>
        {{end}}
    </section>
</div>
{{end}}

{{define "scripts"}}
<script src="/static/js/progressive-enhancement.js"></script>
{{end}}
```

```go
// partials/prompt-card.html
{{define "prompt-card"}}
<article class="prompt-card" data-prompt-id="{{.ID}}">
    <div class="prompt-header">
        <h3>
            <a href="/prompt/{{.ID}}">{{.Name}}</a>
            {{if .IsBookmarked}}
                <span class="bookmark-indicator" title="Bookmarked">★</span>
            {{end}}
        </h3>
        <span class="source-badge">{{.SourceID}}</span>
    </div>

    <p class="description">{{.Description}}</p>

    {{if .Tags}}
    <div class="tags">
        {{range .Tags}}
        <span class="tag">{{.}}</span>
        {{end}}
    </div>
    {{end}}

    <div class="actions">
        <button
            class="bookmark-btn"
            data-prompt-id="{{.ID}}"
            data-bookmarked="{{.IsBookmarked}}"
            aria-label="Toggle bookmark"
        >
            {{if .IsBookmarked}}★{{else}}☆{{end}}
        </button>
    </div>
</article>
{{end}}
```

### Progressive Enhancement JavaScript

Minimal JavaScript that enhances the form-based filtering:

```javascript
// static/js/progressive-enhancement.js
(function() {
    'use strict';

    // Feature detection
    if (!('URLSearchParams' in window)) {
        console.log('URLSearchParams not supported, using basic forms');
        return; // Gracefully degrade to form submission
    }

    const filterForm = document.getElementById('filter-form');
    const promptList = document.getElementById('prompt-list');
    const searchInput = document.getElementById('search');

    let debounceTimer;

    // Progressive enhancement: real-time search
    if (searchInput && filterForm) {
        searchInput.addEventListener('input', function(e) {
            clearTimeout(debounceTimer);
            debounceTimer = setTimeout(() => {
                submitFilters();
            }, 300); // 300ms debounce
        });

        // Enhance filter changes
        filterForm.addEventListener('change', function(e) {
            if (e.target.type !== 'text') {
                submitFilters();
            }
        });

        // Prevent default form submission, use AJAX instead
        filterForm.addEventListener('submit', function(e) {
            e.preventDefault();
            submitFilters();
        });
    }

    function submitFilters() {
        const formData = new FormData(filterForm);
        const params = new URLSearchParams(formData);

        // Add indicator that this is an AJAX request
        params.set('ajax', '1');

        // Update URL without reload
        const newURL = '/?' + params.toString();
        history.replaceState(null, '', newURL);

        // Fetch filtered results
        fetch(newURL, {
            headers: {
                'X-Requested-With': 'XMLHttpRequest'
            }
        })
        .then(response => response.text())
        .then(html => {
            // Update only the prompt list
            promptList.innerHTML = html;
        })
        .catch(error => {
            console.error('Filter request failed:', error);
            // Fallback: do traditional form submission
            filterForm.submit();
        });
    }

    // Bookmark functionality with optimistic updates
    document.addEventListener('click', function(e) {
        if (e.target.matches('.bookmark-btn')) {
            e.preventDefault();
            const btn = e.target;
            const promptID = btn.dataset.promptId;
            const isBookmarked = btn.dataset.bookmarked === 'true';

            // Optimistic update
            btn.textContent = isBookmarked ? '☆' : '★';
            btn.dataset.bookmarked = !isBookmarked;

            // Send request
            fetch('/api/bookmark/' + promptID, {
                method: isBookmarked ? 'DELETE' : 'POST',
                headers: {
                    'Content-Type': 'application/json'
                }
            })
            .catch(error => {
                console.error('Bookmark failed:', error);
                // Revert on failure
                btn.textContent = isBookmarked ? '★' : '☆';
                btn.dataset.bookmarked = isBookmarked;
                alert('Failed to update bookmark. Please try again.');
            });
        }
    });
})();
```

### Go Handler Implementation

```go
package server

import (
    "net/http"
    "net/url"
    "strconv"
    "github.com/whisller/pkit/internal/index"
    "github.com/whisller/pkit/internal/bookmark"
)

type HomePageData struct {
    Query          string
    PromptCount    int
    Sources        []SourceFilter
    Tags           []TagFilter
    SelectedSource string
    ShowBookmarked bool
    Prompts        []PromptView
    Pagination     *PaginationData
}

type PromptView struct {
    ID           string
    Name         string
    Description  string
    SourceID     string
    Tags         []string
    IsBookmarked bool
}

type SourceFilter struct {
    ID    string
    Name  string
    Count int
}

type TagFilter struct {
    Name     string
    Count    int
    Selected bool
}

type PaginationData struct {
    CurrentPage int
    TotalPages  int
    HasPrev     bool
    HasNext     bool
    PrevURL     string
    NextURL     string
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
    // Parse query parameters
    query := r.URL.Query()
    searchQuery := query.Get("q")
    sourceFilter := query.Get("source")
    tagFilters := query["tags"]
    bookmarkedOnly := query.Get("bookmarked") == "true"
    page, _ := strconv.Atoi(query.Get("page"))
    if page < 1 {
        page = 1
    }

    const pageSize = 50

    // Search using Bleve index
    searchOpts := index.SearchOptions{
        Query:      searchQuery,
        SourceID:   sourceFilter,
        Tags:       tagFilters,
        MaxResults: pageSize * page, // Get all results up to current page
    }

    results, err := s.indexer.Search(searchOpts)
    if err != nil {
        http.Error(w, "Search failed", http.StatusInternalServerError)
        return
    }

    // Paginate results
    start := (page - 1) * pageSize
    end := start + pageSize
    if end > len(results) {
        end = len(results)
    }

    var pageResults []index.SearchResult
    if start < len(results) {
        pageResults = results[start:end]
    }

    // Load bookmarks
    bookmarkMgr := bookmark.NewManager()
    bookmarks, _ := bookmarkMgr.ListBookmarks()
    bookmarkMap := make(map[string]bool)
    for _, b := range bookmarks {
        bookmarkMap[b.PromptID] = true
    }

    // Convert to view models
    prompts := make([]PromptView, 0, len(pageResults))
    for _, result := range pageResults {
        prompts = append(prompts, PromptView{
            ID:           result.Prompt.ID,
            Name:         result.Prompt.Name,
            Description:  result.Prompt.Description,
            SourceID:     result.Prompt.SourceID,
            Tags:         result.Prompt.Tags,
            IsBookmarked: bookmarkMap[result.Prompt.ID],
        })
    }

    // Build pagination
    totalPages := (len(results) + pageSize - 1) / pageSize
    pagination := &PaginationData{
        CurrentPage: page,
        TotalPages:  totalPages,
        HasPrev:     page > 1,
        HasNext:     page < totalPages,
    }

    if pagination.HasPrev {
        params := url.Values{}
        params.Set("page", strconv.Itoa(page-1))
        if searchQuery != "" {
            params.Set("q", searchQuery)
        }
        pagination.PrevURL = params.Encode()
    }

    if pagination.HasNext {
        params := url.Values{}
        params.Set("page", strconv.Itoa(page+1))
        if searchQuery != "" {
            params.Set("q", searchQuery)
        }
        pagination.NextURL = params.Encode()
    }

    data := HomePageData{
        Query:          searchQuery,
        PromptCount:    len(results),
        SelectedSource: sourceFilter,
        ShowBookmarked: bookmarkedOnly,
        Prompts:        prompts,
        Pagination:     pagination,
    }

    // Check if this is an AJAX request
    isAJAX := r.Header.Get("X-Requested-With") == "XMLHttpRequest" ||
              query.Get("ajax") == "1"

    if isAJAX {
        // Return only the prompt list partial
        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        for _, prompt := range prompts {
            s.executeTemplate(w, "prompt-card", prompt)
        }
        return
    }

    // Return full page
    s.executeTemplate(w, "home.html", data)
}
```

### Rationale

1. **Progressive Enhancement Philosophy**: The application works perfectly without JavaScript (standard HTML forms), but provides a better experience when JavaScript is available
2. **SEO & Accessibility**: Server-rendered HTML ensures content is immediately available to search engines and screen readers
3. **Bookmarkable State**: Filter state is stored in URL parameters, making all searches bookmarkable and shareable
4. **Performance**: Server-side filtering is fast with Bleve index; minimal JavaScript payload reduces load time
5. **Reliability**: Form fallback ensures the app works even when JavaScript fails to load or is disabled
6. **Standards Compliance**: Uses standard HTML forms and GET parameters following web best practices

### Alternatives Considered

#### 1. Client-Side SPA (React, Vue, Svelte)

**Pros**:
- Rich interactivity without page reloads
- Modern development experience
- Large ecosystem of components

**Cons**:
- Requires Node.js build pipeline (violates no-build-toolchain constraint)
- Large JavaScript bundle sizes
- SEO complexity (requires SSR setup)
- Breaks without JavaScript
- Significant external dependencies

**Why Not Chosen**: Violates multiple constraints (no build toolchain, lightweight, minimal dependencies). Over-engineered for the requirements.

#### 2. HTMX with Server-Side Rendering

**Pros**:
- Declarative AJAX interactions
- Server-side rendering with progressive enhancement
- Minimal JavaScript

**Cons**:
- External dependency (htmx.js, ~14KB)
- Learning curve for custom attributes
- Not necessary for simple filtering

**Why Not Chosen**: While HTMX is excellent, the progressive enhancement approach with vanilla JavaScript achieves similar results with zero dependencies, aligning better with the "minimal dependencies" constraint.

#### 3. Full JavaScript with JSON API

**Pros**:
- Complete separation of concerns
- API could be reused for other clients

**Cons**:
- Requires two implementations (API + client)
- No functionality without JavaScript
- More complex state management
- SEO challenges

**Why Not Chosen**: Doesn't meet progressive enhancement requirement; unnecessarily complex for a single-user local tool.

#### 4. WebAssembly for Client-Side Filtering

**Pros**:
- Could reuse Go code for filtering
- Near-native performance

**Cons**:
- Large initial payload (~2MB+ for Go WASM)
- Complex build process
- Poor browser compatibility in older versions
- Significant runtime initialization time

**Why Not Chosen**: Massive overhead for minimal benefit; violates lightweight constraint.

---

## R003: Localhost Security & Port Management

**Question**: Best practices for ensuring server only binds to localhost and handling port conflicts gracefully?

### Decision

Bind explicitly to `127.0.0.1` (not `localhost`) to avoid IPv6 complications, implement port conflict detection with user-friendly error messages, and provide graceful shutdown on SIGINT/SIGTERM.

### Implementation

```go
package server

import (
    "context"
    "errors"
    "fmt"
    "log"
    "net"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
)

type Config struct {
    Port       int
    Host       string // Should always be 127.0.0.1 for security
    DevMode    bool
    AutoLaunch bool // Automatically open browser
}

type Server struct {
    httpServer *http.Server
    config     Config
    indexer    *index.Indexer
    templates  *template.Template
    devMode    bool
}

func NewServer(cfg Config) (*Server, error) {
    // Force localhost binding for security
    if cfg.Host == "" {
        cfg.Host = "127.0.0.1"
    }

    // Validate that host is localhost/loopback
    if cfg.Host != "127.0.0.1" && cfg.Host != "localhost" && cfg.Host != "::1" {
        return nil, fmt.Errorf("server must bind to localhost only, got: %s", cfg.Host)
    }

    // Default port
    if cfg.Port == 0 {
        cfg.Port = 8080
    }

    s := &Server{
        config:  cfg,
        devMode: cfg.DevMode,
    }

    // Load templates
    templateFS, err := GetTemplateFS(cfg.DevMode)
    if err != nil {
        return nil, fmt.Errorf("failed to load templates: %w", err)
    }

    s.templates, err = template.ParseFS(templateFS,
        "base.html",
        "partials/*.html",
        "pages/*.html",
    )
    if err != nil {
        return nil, fmt.Errorf("failed to parse templates: %w", err)
    }

    // Initialize indexer
    indexPath, err := config.GetIndexPath()
    if err != nil {
        return nil, err
    }

    s.indexer, err = index.Open(indexPath)
    if err != nil {
        return nil, fmt.Errorf("failed to open search index: %w", err)
    }

    return s, nil
}

// CheckPortAvailable checks if the port is available before starting
func (s *Server) CheckPortAvailable() error {
    addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

    // Try to listen on the port briefly
    listener, err := net.Listen("tcp", addr)
    if err != nil {
        if isAddressInUseError(err) {
            return fmt.Errorf(
                "port %d is already in use.\n\n"+
                "Possible solutions:\n"+
                "  1. Stop the process using port %d\n"+
                "  2. Use a different port with --port flag: pkit serve --port 8081\n"+
                "  3. Find the process: lsof -i :%d (Unix) or netstat -ano | findstr :%d (Windows)",
                s.config.Port, s.config.Port, s.config.Port, s.config.Port,
            )
        }
        return fmt.Errorf("failed to bind to %s: %w", addr, err)
    }

    // Close immediately since we're just checking
    listener.Close()
    return nil
}

// isAddressInUseError checks if the error is due to port already in use
func isAddressInUseError(err error) bool {
    var opErr *net.OpError
    if errors.As(err, &opErr) {
        var syscallErr *os.SyscallError
        if errors.As(opErr.Err, &syscallErr) {
            return syscallErr.Err == syscall.EADDRINUSE
        }
    }
    return false
}

// Start starts the HTTP server with graceful shutdown support
func (s *Server) Start() error {
    // Check port availability first
    if err := s.CheckPortAvailable(); err != nil {
        return err
    }

    // Setup routes
    mux := http.NewServeMux()
    if err := s.RegisterRoutes(mux); err != nil {
        return fmt.Errorf("failed to register routes: %w", err)
    }

    // Create HTTP server
    addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
    s.httpServer = &http.Server{
        Addr:           addr,
        Handler:        mux,
        ReadTimeout:    10 * time.Second,
        WriteTimeout:   10 * time.Second,
        IdleTimeout:    120 * time.Second,
        MaxHeaderBytes: 1 << 20, // 1 MB
    }

    // Setup graceful shutdown
    ctx, stop := signal.NotifyContext(
        context.Background(),
        os.Interrupt,        // SIGINT (Ctrl+C)
        syscall.SIGTERM,     // SIGTERM (docker stop, systemd)
        syscall.SIGHUP,      // SIGHUP (terminal closed)
    )
    defer stop()

    // Start server in goroutine
    serverErr := make(chan error, 1)
    go func() {
        log.Printf("Starting pkit web server on http://%s", addr)
        log.Printf("Press Ctrl+C to stop")

        if err := s.httpServer.ListenAndServe(); err != nil &&
           err != http.ErrServerClosed {
            serverErr <- fmt.Errorf("server error: %w", err)
        }
    }()

    // Auto-launch browser if requested
    if s.config.AutoLaunch {
        go func() {
            time.Sleep(500 * time.Millisecond)
            openBrowser(fmt.Sprintf("http://%s", addr))
        }()
    }

    // Wait for interrupt signal or server error
    select {
    case <-ctx.Done():
        log.Println("\nShutdown signal received, stopping server...")

        // Create shutdown context with timeout
        shutdownCtx, cancel := context.WithTimeout(
            context.Background(),
            10*time.Second,
        )
        defer cancel()

        // Attempt graceful shutdown
        if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
            log.Printf("Server shutdown error: %v", err)
            return err
        }

        log.Println("Server stopped gracefully")
        return nil

    case err := <-serverErr:
        return err
    }
}

// Close closes the server and cleans up resources
func (s *Server) Close() error {
    if s.indexer != nil {
        return s.indexer.Close()
    }
    return nil
}

// openBrowser attempts to open the default browser
func openBrowser(url string) {
    var cmd string
    var args []string

    switch runtime.GOOS {
    case "darwin":
        cmd = "open"
        args = []string{url}
    case "windows":
        cmd = "cmd"
        args = []string{"/c", "start", url}
    default: // Linux and others
        cmd = "xdg-open"
        args = []string{url}
    }

    if err := exec.Command(cmd, args...).Start(); err != nil {
        log.Printf("Failed to open browser: %v", err)
        log.Printf("Please open manually: %s", url)
    }
}
```

### CLI Command Implementation

```go
package main

import (
    "fmt"
    "log"

    "github.com/spf13/cobra"
    "github.com/whisller/pkit/internal/web/server"
)

var serveCmd = &cobra.Command{
    Use:   "serve",
    Short: "Start the local web interface",
    Long: `Start a local web server for browsing and managing prompts.

The server only binds to localhost (127.0.0.1) for security and is not
accessible from other machines on your network.

Examples:
  pkit serve                    # Start on default port 8080
  pkit serve --port 3000        # Use custom port
  pkit serve --open             # Auto-open browser
  pkit serve --dev              # Enable development mode`,
    RunE: runServe,
}

var (
    servePort   int
    serveOpen   bool
    serveDevMode bool
)

func init() {
    rootCmd.AddCommand(serveCmd)

    serveCmd.Flags().IntVarP(&servePort, "port", "p", 8080,
        "Port to listen on")
    serveCmd.Flags().BoolVar(&serveOpen, "open", false,
        "Automatically open browser")
    serveCmd.Flags().BoolVar(&serveDevMode, "dev", false,
        "Enable development mode (hot reload templates)")
}

func runServe(cmd *cobra.Command, args []string) error {
    // Create server
    srv, err := server.NewServer(server.Config{
        Port:       servePort,
        Host:       "127.0.0.1",
        DevMode:    serveDevMode,
        AutoLaunch: serveOpen,
    })
    if err != nil {
        return fmt.Errorf("failed to create server: %w", err)
    }
    defer srv.Close()

    // Start server (blocks until shutdown)
    if err := srv.Start(); err != nil {
        return err
    }

    return nil
}
```

### Rationale

1. **Explicit IPv4 Binding**: Using `127.0.0.1` instead of `localhost` avoids IPv6 binding complications (where `localhost` might resolve to `::1`)
2. **Security First**: Explicitly validating that the host is a loopback address prevents accidental network exposure
3. **Port Conflict Detection**: Pre-checking port availability provides immediate, actionable error messages
4. **Graceful Shutdown**: Using `signal.NotifyContext` with multiple signals ensures clean shutdown in various environments (terminal, Docker, systemd)
5. **User-Friendly Errors**: Detailed error messages with solutions help users quickly resolve port conflicts
6. **Timeout Protection**: 10-second shutdown timeout prevents indefinite hanging

### Alternatives Considered

#### 1. Bind to `localhost` hostname

**Pros**:
- More intuitive for users
- Works with both IPv4 and IPv6

**Cons**:
- May bind to both `127.0.0.1` and `::1` depending on OS
- Unpredictable behavior across platforms
- Potential IPv6 firewall prompts

**Why Not Chosen**: `127.0.0.1` provides predictable, consistent behavior across all platforms.

#### 2. Bind to `0.0.0.0` (all interfaces)

**Pros**:
- Allows network access if desired
- More flexible

**Cons**:
- Major security risk for local-only tool
- Exposes application to network attacks
- Triggers firewall warnings
- Violates localhost-only requirement

**Why Not Chosen**: Completely unacceptable for security reasons.

#### 3. Random/Dynamic Port Selection

**Pros**:
- Eliminates port conflicts
- Used by some development tools

**Cons**:
- User can't bookmark or remember the URL
- Complicates browser auto-launch
- Unexpected behavior for users
- Makes it harder to configure firewall rules if needed

**Why Not Chosen**: Fixed default port with override flag provides better user experience while still allowing flexibility.

#### 4. Port Scanning to Find Available Port

**Pros**:
- Automatically avoids conflicts
- Seamless startup

**Cons**:
- Adds complexity
- May find unexpected ports
- Users lose control over port selection
- Security auditing becomes harder

**Why Not Chosen**: Explicit port specification with clear error messages is more transparent and predictable.

---

## R004: State Management Without Database

**Question**: How to efficiently manage filter/search state and pagination using existing YAML storage?

### Decision

Use in-memory caching of the Bleve search index with full load at startup, implement URL-based state management for filters and pagination, and use file-locking for concurrent-safe bookmark/tag writes.

### Architecture

```go
package server

import (
    "sync"
    "time"

    "github.com/whisller/pkit/internal/index"
    "github.com/whisller/pkit/internal/bookmark"
    "github.com/whisller/pkit/pkg/models"
)

// PromptCache provides fast access to prompt data
type PromptCache struct {
    mu           sync.RWMutex
    indexer      *index.Indexer
    bookmarks    map[string]models.Bookmark
    tags         map[string][]string // promptID -> tags
    lastReload   time.Time
    reloadTicker *time.Ticker
}

func NewPromptCache(indexer *index.Indexer) *PromptCache {
    pc := &PromptCache{
        indexer:   indexer,
        bookmarks: make(map[string]models.Bookmark),
        tags:      make(map[string][]string),
    }

    // Initial load
    pc.reload()

    // Auto-reload every 30 seconds to pick up CLI changes
    pc.reloadTicker = time.NewTicker(30 * time.Second)
    go pc.autoReload()

    return pc
}

func (pc *PromptCache) reload() error {
    pc.mu.Lock()
    defer pc.mu.Unlock()

    // Load bookmarks
    mgr := bookmark.NewManager()
    bookmarks, err := mgr.ListBookmarks()
    if err != nil {
        return err
    }

    newBookmarks := make(map[string]models.Bookmark)
    newTags := make(map[string][]string)

    for _, bm := range bookmarks {
        newBookmarks[bm.PromptID] = bm
        if len(bm.Tags) > 0 {
            newTags[bm.PromptID] = bm.Tags
        }
    }

    pc.bookmarks = newBookmarks
    pc.tags = newTags
    pc.lastReload = time.Now()

    return nil
}

func (pc *PromptCache) autoReload() {
    for range pc.reloadTicker.C {
        if err := pc.reload(); err != nil {
            log.Printf("Cache reload failed: %v", err)
        }
    }
}

func (pc *PromptCache) Close() {
    if pc.reloadTicker != nil {
        pc.reloadTicker.Stop()
    }
}

// IsBookmarked checks if a prompt is bookmarked
func (pc *PromptCache) IsBookmarked(promptID string) bool {
    pc.mu.RLock()
    defer pc.mu.RUnlock()

    _, exists := pc.bookmarks[promptID]
    return exists
}

// GetTags gets user tags for a prompt
func (pc *PromptCache) GetTags(promptID string) []string {
    pc.mu.RLock()
    defer pc.mu.RUnlock()

    tags, exists := pc.tags[promptID]
    if !exists {
        return []string{}
    }

    // Return copy to prevent external modification
    result := make([]string, len(tags))
    copy(result, tags)
    return result
}

// Search performs a search with caching
func (pc *PromptCache) Search(opts index.SearchOptions) ([]index.SearchResult, error) {
    // Bleve index already provides efficient in-memory search
    results, err := pc.indexer.Search(opts)
    if err != nil {
        return nil, err
    }

    // Enrich with bookmark status (from cache)
    pc.mu.RLock()
    defer pc.mu.RUnlock()

    for i := range results {
        // Note: Bookmark status should be added in view layer
        // This is just for reference
        _ = pc.bookmarks[results[i].Prompt.ID]
    }

    return results, nil
}
```

### Concurrent-Safe Bookmark Operations

```go
package bookmark

import (
    "fmt"
    "os"
    "path/filepath"
    "syscall"

    "github.com/goccy/go-yaml"
    "github.com/whisller/pkit/pkg/models"
    "github.com/whisller/pkit/internal/config"
)

// GetBookmarksPath returns the path to bookmarks.yml
func GetBookmarksPath() (string, error) {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return "", err
    }
    return filepath.Join(homeDir, ".pkit", "bookmarks.yml"), nil
}

// LoadBookmarks loads bookmarks from YAML with file locking
func LoadBookmarks() ([]models.Bookmark, error) {
    path, err := GetBookmarksPath()
    if err != nil {
        return nil, err
    }

    // Check if file exists
    if _, err := os.Stat(path); os.IsNotExist(err) {
        return []models.Bookmark{}, nil
    }

    // Open file with read lock
    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    // Acquire shared lock (multiple readers allowed)
    if err := syscall.Flock(int(file.Fd()), syscall.LOCK_SH); err != nil {
        return nil, fmt.Errorf("failed to acquire read lock: %w", err)
    }
    defer syscall.Flock(int(file.Fd()), syscall.LOCK_UN)

    // Read and parse
    var bookmarks []models.Bookmark
    decoder := yaml.NewDecoder(file)
    if err := decoder.Decode(&bookmarks); err != nil {
        return nil, err
    }

    return bookmarks, nil
}

// SaveBookmarks saves bookmarks to YAML with atomic write and file locking
func SaveBookmarks(bookmarks []models.Bookmark) error {
    path, err := GetBookmarksPath()
    if err != nil {
        return err
    }

    // Ensure directory exists
    if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
        return err
    }

    // Marshal to YAML first (fail fast if marshaling fails)
    data, err := yaml.Marshal(bookmarks)
    if err != nil {
        return fmt.Errorf("failed to marshal bookmarks: %w", err)
    }

    // Write to temporary file
    tmpPath := path + ".tmp"
    tmpFile, err := os.Create(tmpPath)
    if err != nil {
        return err
    }

    // Acquire exclusive lock on temp file
    if err := syscall.Flock(int(tmpFile.Fd()), syscall.LOCK_EX); err != nil {
        tmpFile.Close()
        os.Remove(tmpPath)
        return fmt.Errorf("failed to acquire write lock: %w", err)
    }

    // Write data
    if _, err := tmpFile.Write(data); err != nil {
        syscall.Flock(int(tmpFile.Fd()), syscall.LOCK_UN)
        tmpFile.Close()
        os.Remove(tmpPath)
        return err
    }

    // Sync to disk
    if err := tmpFile.Sync(); err != nil {
        syscall.Flock(int(tmpFile.Fd()), syscall.LOCK_UN)
        tmpFile.Close()
        os.Remove(tmpPath)
        return err
    }

    // Release lock and close
    syscall.Flock(int(tmpFile.Fd()), syscall.LOCK_UN)
    tmpFile.Close()

    // Atomic rename
    if err := os.Rename(tmpPath, path); err != nil {
        os.Remove(tmpPath)
        return fmt.Errorf("failed to save bookmarks: %w", err)
    }

    return nil
}
```

### URL State Management Pattern

All filter/search state is stored in URL parameters:

```
http://127.0.0.1:8080/?q=kubernetes&source=fabric&tags=devops,automation&bookmarked=true&page=2
```

Query parameters:
- `q`: Search query
- `source`: Source filter
- `tags`: Comma-separated tag filters
- `bookmarked`: Show only bookmarked prompts (true/false)
- `page`: Current page number

This approach provides:
- **Bookmarkability**: Users can bookmark specific searches
- **Shareability**: URLs can be copied (though only useful locally)
- **Browser history**: Back/forward buttons work naturally
- **No state management**: Server is stateless; all state is in URL
- **Cache-friendly**: Same URL always produces same results

### Pagination Implementation

```go
package server

import (
    "net/url"
    "strconv"
)

type Paginator struct {
    CurrentPage  int
    PageSize     int
    TotalResults int
}

func NewPaginator(page, pageSize, totalResults int) *Paginator {
    if page < 1 {
        page = 1
    }
    if pageSize < 1 {
        pageSize = 50
    }

    return &Paginator{
        CurrentPage:  page,
        PageSize:     pageSize,
        TotalResults: totalResults,
    }
}

func (p *Paginator) TotalPages() int {
    if p.TotalResults == 0 {
        return 1
    }
    return (p.TotalResults + p.PageSize - 1) / p.PageSize
}

func (p *Paginator) HasPrevious() bool {
    return p.CurrentPage > 1
}

func (p *Paginator) HasNext() bool {
    return p.CurrentPage < p.TotalPages()
}

func (p *Paginator) Start() int {
    return (p.CurrentPage - 1) * p.PageSize
}

func (p *Paginator) End() int {
    end := p.Start() + p.PageSize
    if end > p.TotalResults {
        end = p.TotalResults
    }
    return end
}

func (p *Paginator) BuildURL(baseParams url.Values, page int) string {
    params := url.Values{}

    // Copy existing parameters
    for k, v := range baseParams {
        if k != "page" { // Don't copy page param
            params[k] = v
        }
    }

    // Add page parameter
    params.Set("page", strconv.Itoa(page))

    return params.Encode()
}

func (p *Paginator) PreviousURL(baseParams url.Values) string {
    if !p.HasPrevious() {
        return ""
    }
    return p.BuildURL(baseParams, p.CurrentPage-1)
}

func (p *Paginator) NextURL(baseParams url.Values) string {
    if !p.HasNext() {
        return ""
    }
    return p.BuildURL(baseParams, p.CurrentPage+1)
}
```

### Rationale

1. **Bleve In-Memory Performance**: The existing Bleve index provides fast in-memory search (50-100ms for 10M+ documents), eliminating the need for a separate cache layer
2. **Lightweight Caching**: Only bookmarks and user tags are cached in memory, as these are small datasets that change infrequently
3. **File Locking Safety**: `syscall.Flock` provides OS-level file locking, preventing race conditions between web and CLI writes
4. **Atomic Writes**: Temp file + rename pattern ensures YAML files are never corrupted, even with concurrent access
5. **URL State Management**: Stateless server design with URL-based state provides bookmarkability and simplicity
6. **Periodic Reload**: 30-second reload interval picks up CLI changes without manual refresh while avoiding constant file system polling
7. **Read-Write Lock Pattern**: RWMutex allows multiple concurrent reads while ensuring write safety

### Alternatives Considered

#### 1. SQLite Database

**Pros**:
- Structured query support
- Built-in ACID guarantees
- Better concurrency handling
- Faster complex queries

**Cons**:
- Requires schema migrations
- Additional dependency (though cgo-free versions exist)
- Duplicate storage (YAML + SQLite)
- Complicates CLI compatibility
- Adds significant complexity

**Why Not Chosen**: Bleve already provides fast search; bookmarks/tags are small datasets that don't require a database. YAML storage aligns with existing CLI design.

#### 2. Redis for Caching

**Pros**:
- Very fast in-memory caching
- Built-in pub/sub for cache invalidation
- Rich data structures

**Cons**:
- External service dependency (requires Redis server)
- Massive overkill for local single-user tool
- Violates lightweight constraint
- Adds deployment complexity

**Why Not Chosen**: Completely inappropriate for a local, single-user CLI tool. In-memory maps provide equivalent performance.

#### 3. Session Cookies for State

**Pros**:
- Standard web pattern
- Doesn't pollute URLs
- Can store more data

**Cons**:
- Not bookmarkable (major UX regression)
- Breaks browser history
- Requires session management
- Complicates sharing/debugging

**Why Not Chosen**: URL-based state provides better UX (bookmarkable, shareable URLs) for a read-heavy application.

#### 4. File System Watching (fsnotify)

**Pros**:
- Immediate detection of YAML changes
- No polling overhead

**Cons**:
- Additional dependency
- Platform-specific behavior
- Complexity (buffering, debouncing)
- Overkill for infrequent updates

**Why Not Chosen**: 30-second periodic reload is simple and sufficient; bookmarks/tags change infrequently enough that instant updates aren't necessary.

---

## R005: Client-Side Clipboard API

**Question**: What's the most reliable cross-browser approach for copying prompt content to clipboard?

### Decision

Use the modern Clipboard API (`navigator.clipboard.writeText()`) with a fallback to `document.execCommand('copy')` for older browsers, combined with proper error handling and user feedback.

### Implementation

```javascript
// static/js/clipboard.js
(function() {
    'use strict';

    /**
     * Copy text to clipboard with modern API and fallback
     * @param {string} text - The text to copy
     * @returns {Promise<boolean>} - True if successful, false otherwise
     */
    async function copyToClipboard(text) {
        // Method 1: Modern Clipboard API (preferred)
        if (navigator.clipboard && window.isSecureContext) {
            try {
                await navigator.clipboard.writeText(text);
                return true;
            } catch (err) {
                console.warn('Clipboard API failed, trying fallback:', err);
                // Continue to fallback
            }
        }

        // Method 2: Fallback using execCommand (deprecated but widely supported)
        return copyToClipboardFallback(text);
    }

    /**
     * Fallback clipboard copy using execCommand
     * @param {string} text - The text to copy
     * @returns {boolean} - True if successful, false otherwise
     */
    function copyToClipboardFallback(text) {
        // Create temporary textarea
        const textarea = document.createElement('textarea');

        // Style it to be invisible and non-interactive
        textarea.value = text;
        textarea.style.position = 'fixed';
        textarea.style.left = '-999999px';
        textarea.style.top = '-999999px';
        textarea.style.opacity = '0';
        textarea.setAttribute('readonly', '');
        textarea.setAttribute('aria-hidden', 'true');

        document.body.appendChild(textarea);

        // Select the text
        textarea.focus();
        textarea.select();

        // For iOS compatibility
        textarea.setSelectionRange(0, text.length);

        let success = false;
        try {
            // Execute copy command
            success = document.execCommand('copy');
        } catch (err) {
            console.error('Fallback copy failed:', err);
        }

        // Clean up
        document.body.removeChild(textarea);

        return success;
    }

    /**
     * Show temporary feedback message
     * @param {HTMLElement} element - Element near which to show message
     * @param {string} message - Message to display
     * @param {boolean} isSuccess - Whether this is a success message
     */
    function showFeedback(element, message, isSuccess) {
        // Remove any existing feedback
        const existingFeedback = document.querySelector('.copy-feedback');
        if (existingFeedback) {
            existingFeedback.remove();
        }

        // Create feedback element
        const feedback = document.createElement('div');
        feedback.className = 'copy-feedback ' + (isSuccess ? 'success' : 'error');
        feedback.textContent = message;
        feedback.setAttribute('role', 'status');
        feedback.setAttribute('aria-live', 'polite');

        // Position near the button
        const rect = element.getBoundingClientRect();
        feedback.style.position = 'fixed';
        feedback.style.left = rect.left + 'px';
        feedback.style.top = (rect.bottom + 5) + 'px';

        document.body.appendChild(feedback);

        // Fade in
        setTimeout(() => feedback.classList.add('visible'), 10);

        // Remove after 2 seconds
        setTimeout(() => {
            feedback.classList.remove('visible');
            setTimeout(() => feedback.remove(), 300);
        }, 2000);
    }

    /**
     * Setup clipboard copy for all copy buttons
     */
    function setupClipboardButtons() {
        document.addEventListener('click', async function(e) {
            const copyBtn = e.target.closest('[data-copy-text]');
            if (!copyBtn) return;

            e.preventDefault();

            const text = copyBtn.getAttribute('data-copy-text');
            if (!text) {
                showFeedback(copyBtn, 'Nothing to copy', false);
                return;
            }

            // Disable button during copy
            const originalText = copyBtn.textContent;
            copyBtn.disabled = true;
            copyBtn.textContent = 'Copying...';

            try {
                const success = await copyToClipboard(text);

                if (success) {
                    showFeedback(copyBtn, 'Copied to clipboard!', true);
                    copyBtn.textContent = 'Copied!';

                    // Reset button text after delay
                    setTimeout(() => {
                        copyBtn.textContent = originalText;
                        copyBtn.disabled = false;
                    }, 2000);
                } else {
                    throw new Error('Copy failed');
                }
            } catch (err) {
                console.error('Copy failed:', err);
                showFeedback(copyBtn, 'Failed to copy', false);
                copyBtn.textContent = originalText;
                copyBtn.disabled = false;
            }
        });
    }

    /**
     * Check if clipboard functionality is available
     * @returns {boolean}
     */
    function isClipboardAvailable() {
        return !!(navigator.clipboard || document.queryCommandSupported('copy'));
    }

    // Initialize when DOM is ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', setupClipboardButtons);
    } else {
        setupClipboardButtons();
    }

    // Hide copy buttons if clipboard is not available at all
    if (!isClipboardAvailable()) {
        const style = document.createElement('style');
        style.textContent = '[data-copy-text] { display: none !important; }';
        document.head.appendChild(style);
    }

    // Export for potential testing
    window.pkitClipboard = {
        copy: copyToClipboard,
        isAvailable: isClipboardAvailable
    };
})();
```

### CSS for Feedback

```css
/* static/css/clipboard.css */

.copy-feedback {
    padding: 8px 12px;
    border-radius: 4px;
    font-size: 14px;
    font-weight: 500;
    z-index: 1000;
    opacity: 0;
    transition: opacity 0.3s ease;
    pointer-events: none;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

.copy-feedback.visible {
    opacity: 1;
}

.copy-feedback.success {
    background-color: #10b981;
    color: white;
}

.copy-feedback.error {
    background-color: #ef4444;
    color: white;
}

button[data-copy-text] {
    cursor: pointer;
    transition: opacity 0.2s;
}

button[data-copy-text]:disabled {
    opacity: 0.6;
    cursor: not-allowed;
}
```

### HTML Usage in Templates

```html
<!-- In prompt detail page -->
<article class="prompt-detail">
    <header>
        <h1>{{.Prompt.Name}}</h1>
        <button
            class="btn btn-copy"
            data-copy-text="{{.Prompt.Content}}"
            aria-label="Copy prompt to clipboard"
        >
            Copy to Clipboard
        </button>
    </header>

    <div class="prompt-content">
        <pre>{{.Prompt.Content}}</pre>
    </div>
</article>
```

### Alternative: Full Content from Server

For very large prompts, fetch content on demand:

```javascript
// Enhanced version with server fetch
document.addEventListener('click', async function(e) {
    const copyBtn = e.target.closest('[data-copy-prompt-id]');
    if (!copyBtn) return;

    e.preventDefault();

    const promptID = copyBtn.getAttribute('data-copy-prompt-id');
    copyBtn.disabled = true;
    copyBtn.textContent = 'Loading...';

    try {
        // Fetch full content from server
        const response = await fetch(`/api/prompt/${promptID}/content`);
        if (!response.ok) throw new Error('Failed to fetch');

        const data = await response.json();
        const success = await copyToClipboard(data.content);

        if (success) {
            showFeedback(copyBtn, 'Copied to clipboard!', true);
            copyBtn.textContent = 'Copied!';
        } else {
            throw new Error('Copy failed');
        }
    } catch (err) {
        showFeedback(copyBtn, 'Failed to copy', false);
    } finally {
        setTimeout(() => {
            copyBtn.textContent = 'Copy';
            copyBtn.disabled = false;
        }, 2000);
    }
});
```

### Server Endpoint for Content

```go
func (s *Server) handlePromptContent(w http.ResponseWriter, r *http.Request) {
    // Extract prompt ID from URL
    promptID := strings.TrimPrefix(r.URL.Path, "/api/prompt/")
    promptID = strings.TrimSuffix(promptID, "/content")

    // Get prompt with full content
    prompt, err := s.indexer.GetPromptByID(promptID)
    if err != nil {
        http.Error(w, "Prompt not found", http.StatusNotFound)
        return
    }

    // Load full content from source file if needed
    if prompt.Content == "" {
        // Content not in index, load from file
        content, err := s.loadPromptContent(prompt)
        if err != nil {
            http.Error(w, "Failed to load content", http.StatusInternalServerError)
            return
        }
        prompt.Content = content
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "content": prompt.Content,
    })
}
```

### Rationale

1. **Modern First**: Clipboard API is the recommended approach, supported in all modern browsers as of 2024-2025
2. **Progressive Enhancement**: Fallback to execCommand ensures compatibility with older browsers
3. **Security Aware**: Checks for secure context (HTTPS/localhost) required by Clipboard API
4. **User Feedback**: Visual confirmation of copy success/failure improves UX
5. **Accessibility**: Includes ARIA attributes and proper semantic elements
6. **Error Handling**: Gracefully handles failures without breaking the page
7. **No Dependencies**: Pure vanilla JavaScript, no external libraries needed

### Browser Compatibility

| Method | Chrome | Firefox | Safari | Edge |
|--------|--------|---------|--------|------|
| navigator.clipboard | 63+ | 53+ | 13.1+ | 79+ |
| execCommand fallback | All | All | All | All |

Since this is a localhost application (secure context), the Clipboard API will work in all modern browsers.

### Alternatives Considered

#### 1. Third-Party Libraries (clipboard.js, copy-to-clipboard)

**Pros**:
- Battle-tested
- Handles edge cases
- Simpler API

**Cons**:
- External dependency (~5KB)
- Overkill for simple text copying
- Not necessary with modern browsers

**Why Not Chosen**: Modern Clipboard API + simple fallback provides equivalent functionality without dependencies.

#### 2. Flash-Based Clipboard (ZeroClipboard)

**Pros**:
- Worked everywhere in the past

**Cons**:
- Flash is dead (EOL December 2020)
- Not supported in modern browsers
- Security issues

**Why Not Chosen**: Completely obsolete technology.

#### 3. Native Electron/Desktop App Clipboard

**Pros**:
- Direct system clipboard access
- No browser limitations

**Cons**:
- Requires Electron or similar framework
- Much larger binary size
- More complex distribution
- Violates lightweight web interface requirement

**Why Not Chosen**: Web-based solution is simpler and meets all requirements.

#### 4. Server-Side Clipboard Integration

**Pros**:
- Works without browser support

**Cons**:
- Requires OS-level integrations
- Platform-specific code
- Security/permission issues
- Doesn't work over web interface

**Why Not Chosen**: Not applicable for web interface; clipboard access should be client-side for security.

---

## Summary & Recommendations

### Implementation Priority

1. **R001 (Embed)**: Implement first - foundation for serving assets
2. **R003 (Localhost Security)**: Implement second - critical for security
3. **R002 (Templates)**: Implement third - core UI functionality
4. **R004 (State Management)**: Implement fourth - data layer
5. **R005 (Clipboard)**: Implement last - enhancement feature

### Key Dependencies

- **Go 1.16+**: For `embed` package support
- **Existing pkit dependencies**: Bleve, goccy/go-yaml (already in project)
- **Zero new dependencies**: All solutions use standard library or existing deps

### Risk Mitigation

1. **Port Conflicts**: Detailed error messages with solutions guide users
2. **Concurrent Access**: File locking prevents corruption when CLI and web run simultaneously
3. **Browser Compatibility**: Progressive enhancement ensures basic functionality works everywhere
4. **Template Errors**: Development mode with hot reload catches issues early
5. **Memory Usage**: Periodic cache refresh prevents unbounded growth

### Performance Expectations

- **Startup time**: < 1 second (index already open)
- **Search response**: < 500ms for 1000+ prompts
- **Page render**: < 100ms (server-side templating is fast)
- **Memory usage**: ~50-80MB (Bleve index + cache)
- **Binary size**: +2-3MB (embedded assets)

### Next Steps

1. Create the directory structure for web server code:
   ```
   internal/web/
   ├── server/
   │   ├── server.go
   │   ├── routes.go
   │   ├── handlers.go
   │   └── middleware.go
   ├── static/
   │   ├── css/
   │   │   ├── style.css
   │   │   └── clipboard.css
   │   └── js/
   │       ├── progressive-enhancement.js
   │       └── clipboard.js
   └── templates/
       ├── base.html
       ├── partials/
       │   ├── header.html
       │   ├── footer.html
       │   ├── prompt-card.html
       │   └── filters.html
       └── pages/
           ├── home.html
           └── detail.html
   ```

2. Implement server initialization and startup
3. Create base templates with inheritance
4. Implement core routes (home, detail, API endpoints)
5. Add progressive enhancement JavaScript
6. Test cross-browser compatibility
7. Verify concurrent CLI/web access safety

---

## Sources

### R001: Go HTTP Server Patterns
- [embed package - Go Packages](https://pkg.go.dev/embed)
- [Mastering Embed in Go 1.16: Bundle Out Static Content](https://lakefs.io/blog/working-with-embed-in-go/)
- [Go Static Assets Embedding vs. Traditional Serving | Leapcell](https://leapcell.io/blog/go-static-assets-embedding-vs-traditional-serving)
- [Embedding Frontend Assets in Go Binaries with Embed Package | Leapcell](https://leapcell.io/blog/embedding-frontend-assets-in-go-binaries-with-embed-package)
- [Go Embedding Series — 03. The Ultimate Guide to go:embed Syntax and Usage](https://ehewen.com/en/blog/go-embed/)

### R002: Server-Side Rendering Architecture
- [HTML templating and inheritance — Let's Go](https://lets-go.alexedwards.net/sample/02.08-html-templating-and-inheritance.html)
- [template package - html/template - Go Packages](https://pkg.go.dev/html/template)
- [Golang Templates, Part 1: Concepts and Composition | francoposa.io](https://francoposa.io/resources/golang/golang-templates-1/)
- [Advanced Go Template Rendering for Robust Server-Side Applications | Leapcell](https://leapcell.io/blog/advanced-go-template-rendering-for-robust-server-side-applications)
- [Clean UI with Go's HTML Templates: Base, Partials, and FuncMaps | Medium](https://medium.com/@uygaroztcyln/clean-ui-with-gos-html-templates-base-partials-and-funcmaps-4915296c9097)
- [Server-Side-Rendering Renaissance](https://daily.dev/blog/server-side-rendering-renaissance)
- [Server Actions vs Client Rendering in Next.js: The 2025 Guide](https://asepalazhari.com/blog/server-actions-vs-client-rendering-nextjs-guide)
- [The Great Rendering Battle: Server-Side vs Client-Side Rendering in 2025](https://dev.to/adamgolan/the-great-rendering-battle-server-side-vs-client-side-rendering-in-2025-3cm5)

### R003: Localhost Security & Port Management
- [Building Local HTTP Servers in Go: Development to Production](https://www.ksred.com/building-local-http-servers-in-go-from-development-to-production-patterns/)
- [How to get rid of the annoying firewall prompt on 'go run'](https://aarol.dev/posts/go-windows-firewall/)
- [How To Make an HTTP Server in Go | DigitalOcean](https://www.digitalocean.com/community/tutorials/how-to-make-an-http-server-in-go)
- [Golang http ListenAndServe Function](https://www.javaguides.net/2025/01/golang-http-listenandserve-function.html)
- [Graceful Shutdown in Go: Practical Patterns](https://victoriametrics.com/blog/go-graceful-shutdown/)
- [Graceful shutdown in Go – Luís Simas](https://luissimas.github.io/blog/graceful-shutdown-go/)
- [How to handle signals with Go to graceful shutdown HTTP server](https://rafallorenz.com/go/handle-signals-to-graceful-shutdown-http-server/)
- [How to Build an HTTP Server in Go: Part 2 — Graceful Shutdown and Timeouts](https://blogs.learningdevops.com/how-to-build-an-http-server-in-go-part-2-graceful-shutdown-and-timeouts-008d90e42da2)

### R004: State Management Without Database
- [Bleve — Modern text indexing library for go](https://blevesearch.com/)
- [GitHub - blevesearch/bleve](https://github.com/blevesearch/bleve)
- [Build a fast search engine in Golang](https://kevincoder.co.za/bleve-how-to-build-a-rocket-fast-search-engine)
- [Concurrency Control in Go: Mutexes, Read-Write Locks, and Atomic Operations](https://clouddevs.com/go/concurrency-control/)
- [Concurrent operations on a file with Go | Medium](https://medium.com/@abdelrahmanmustafa/concurrent-operations-on-a-file-with-go-346123e5c84f)
- [Atomic Operations and Synchronization Primitives - Go Optimization Guide](https://goperf.dev/01-common-patterns/atomic-ops/)
- [Dev vs Ops - Go embed | Simon's Blog](https://www.simonam.dev/dev-vs-ops-using-go-embed/)
- [Hot Reload for Golang with Go Air](https://www.bytesizego.com/blog/golang-air)

### R005: Client-Side Clipboard API
- [Clipboard API - Web APIs | MDN](https://developer.mozilla.org/en-US/docs/Web/API/Clipboard_API)
- [Interact with the clipboard - Mozilla | MDN](https://developer.mozilla.org/en-US/docs/Mozilla/Add-ons/WebExtensions/Interact_with_the_clipboard)
- [JavaScript Clipboard API with fallback – SiteLint](https://www.sitelint.com/blog/javascript-clipboard-api-with-fallback)
- [navigator.clipboard - The New Asynchronous Clipboard API](https://www.trevorlasn.com/blog/javascript-navigator-clipboard)
- [Clipboard API fallback - DEV Community](https://dev.to/phuocng/clipboard-api-fallback-nh7)
- [How do I copy to the clipboard in JavaScript? | Sentry](https://sentry.io/answers/how-do-i-copy-to-the-clipboard-in-javascript/)
