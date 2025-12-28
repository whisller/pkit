# Development Quickstart: Web Interface

**Feature**: 003-web-interface
**Date**: 2025-12-28
**Audience**: Developers implementing the web interface

## Prerequisites

- Go 1.25.4+ installed
- pkit repository cloned
- Existing pkit CLI functional with subscribed sources
- Sample data in `~/.pkit/` (bookmarks, tags, sources)

## Quick Start

### 1. Run Web Server (Development Mode)

```bash
# From repo root
go run cmd/pkit/main.go serve --port 8080

# Server output:
# pkit web server starting...
# Listening on http://127.0.0.1:8080
# Press Ctrl+C to stop
```

### 2. Access Web Interface

Open browser to: **http://127.0.0.1:8080**

### 3. Test Basic Functionality

- **List view**: See all prompts
- **Search**: Type in search box, press Enter
- **Filters**: Click source/tag filters
- **Detail**: Click any prompt card
- **Bookmark**: Click bookmark button on detail page
- **Tags**: Edit tags in sidebar
- **Copy**: Click "Copy to Clipboard" button

## Project Structure

```
pkit/
├── cmd/
│   └── pkit/
│       └── main.go              # Add 'serve' command here
├── internal/
│   └── web/                     # NEW: Web server package
│       ├── server.go            # HTTP server setup
│       ├── handlers.go          # Request handlers
│       ├── middleware.go        # Middleware (logging, localhost check)
│       ├── templates/           # HTML templates
│       │   ├── layout.html
│       │   ├── list.html
│       │   ├── detail.html
│       │   ├── error.html
│       │   └── components/
│       │       ├── filters.html
│       │       ├── prompt-card.html
│       │       ├── pagination.html
│       │       ├── prompt-header.html
│       │       ├── prompt-content.html
│       │       └── tag-editor.html
│       ├── static/              # Static assets
│       │   ├── style.css
│       │   └── app.js
│       └── embed.go             # Go embed directives
└── tests/
    └── web/                     # Web tests
        ├── handlers_test.go
        └── integration_test.go
```

## Development Workflow

### Step 1: Implement Core Server

**File**: `internal/web/server.go`

```go
package web

import (
    "context"
    "fmt"
    "net/http"
    "time"
)

type Server struct {
    mux    *http.ServeMux
    server *http.Server
    // Add dependencies (bookmark manager, tag manager, index, etc.)
}

func NewServer(port int) *Server {
    mux := http.NewServeMux()

    s := &Server{
        mux: mux,
        server: &http.Server{
            Addr:    fmt.Sprintf("127.0.0.1:%d", port),
            Handler: mux,
        },
    }

    s.registerRoutes()
    return s
}

func (s *Server) Start() error {
    fmt.Printf("pkit web server starting...\n")
    fmt.Printf("Listening on http://%s\n", s.server.Addr)
    return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
    return s.server.Shutdown(ctx)
}
```

### Step 2: Register Routes

**File**: `internal/web/handlers.go`

```go
func (s *Server) registerRoutes() {
    // HTML endpoints
    s.mux.HandleFunc("/", s.handleList)
    s.mux.HandleFunc("/prompts/", s.handleDetail)

    // JSON API endpoints
    s.mux.HandleFunc("/api/bookmarks", s.handleBookmarkToggle)
    s.mux.HandleFunc("/api/tags", s.handleTagUpdate)
    s.mux.HandleFunc("/api/search", s.handleSearch)

    // Static assets (embedded)
    s.mux.Handle("/static/", http.StripPrefix("/static/",
        http.FileServer(http.FS(staticFS))))

    // Health check
    s.mux.HandleFunc("/health", s.handleHealth)
}
```

### Step 3: Implement List Handler

```go
func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
    // 1. Parse query parameters (search, source, tags, bookmarked, page)
    filters := parseFilters(r.URL.Query())

    // 2. Load prompts from index
    prompts := s.index.GetAllPrompts()

    // 3. Apply filters
    filtered := applyFilters(prompts, filters)

    // 4. Paginate (50 per page)
    paginated := paginate(filtered, filters.Page, 50)

    // 5. Load bookmarks and tags for each prompt
    for i := range paginated.Items {
        paginated.Items[i].Bookmarked = s.bookmarkMgr.IsBookmarked(paginated.Items[i].Prompt.ID)
        paginated.Items[i].Tags = s.tagMgr.GetTags(paginated.Items[i].Prompt.ID)
    }

    // 6. Render template
    data := ListPageData{
        Items:       paginated.Items,
        TotalItems:  paginated.TotalItems,
        // ... more fields
    }

    s.renderTemplate(w, "list.html", data)
}
```

### Step 4: Embed Templates and Static Assets

**File**: `internal/web/embed.go`

```go
package web

import "embed"

//go:embed templates/*.html templates/**/*.html
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS
```

### Step 5: Add CLI Command

**File**: `cmd/pkit/main.go`

```go
import (
    "github.com/whisller/pkit/internal/web"
    "github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
    Use:   "serve",
    Short: "Start local web server",
    Long:  "Start a local web server to browse prompts in your browser",
    RunE: func(cmd *cobra.Command, args []string) error {
        port, _ := cmd.Flags().GetInt("port")

        server := web.NewServer(port)

        // Handle graceful shutdown
        ctx, cancel := context.WithCancel(context.Background())
        defer cancel()

        go func() {
            sigChan := make(chan os.Signal, 1)
            signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
            <-sigChan
            fmt.Println("\nShutting down...")
            server.Shutdown(ctx)
        }()

        return server.Start()
    },
}

func init() {
    rootCmd.AddCommand(serveCmd)
    serveCmd.Flags().IntP("port", "p", 8080, "Port to listen on")
}
```

## Testing Sample Data

### Create Test Prompts

```bash
# Subscribe to sources (if not already)
pkit subscribe fabric/patterns
pkit subscribe awesome-chatgpt-prompts

# Create some bookmarks
pkit bookmark fabric:summarize
pkit bookmark awesome:code-review

# Add tags
pkit tag fabric:summarize dev,documentation
pkit tag awesome:code-review dev,security
```

### Verify Data Files

```bash
# Check bookmarks
cat ~/.pkit/bookmarks.yml

# Check tags
cat ~/.pkit/tags.yml

# Check sources
ls ~/.pkit/sources/
```

## Testing

### Unit Tests

**File**: `tests/web/handlers_test.go`

```go
package web_test

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/whisller/pkit/internal/web"
)

func TestListHandler(t *testing.T) {
    server := web.NewServer(8080)

    req := httptest.NewRequest("GET", "/", nil)
    w := httptest.NewRecorder()

    server.ServeHTTP(w, req)

    if w.Code != http.StatusOK {
        t.Errorf("Expected 200, got %d", w.Code)
    }

    body := w.Body.String()
    if !strings.Contains(body, "Prompts") {
        t.Error("Expected 'Prompts' in response")
    }
}
```

### Integration Tests

```bash
# Start server in background
go run cmd/pkit/main.go serve --port 8080 &
SERVER_PID=$!

# Test endpoints
curl -s http://127.0.0.1:8080/ | grep "Prompts"
curl -s http://127.0.0.1:8080/health | grep "ok"

# Cleanup
kill $SERVER_PID
```

### Manual Browser Testing

1. Start server: `go run cmd/pkit/main.go serve`
2. Open: http://127.0.0.1:8080
3. Test checklist:
   - [ ] List view loads with prompts
   - [ ] Search works (type and press Enter)
   - [ ] Source filter works
   - [ ] Tag filter works
   - [ ] Bookmarked filter works
   - [ ] Pagination works (if 50+ prompts)
   - [ ] Prompt detail page loads
   - [ ] Bookmark button toggles
   - [ ] Tag editing saves
   - [ ] Copy to clipboard works
   - [ ] Back navigation preserves filters

## Modifying Templates

### Development Mode (Hot Reload)

**Option 1**: Load templates from disk (not embedded)

```go
// In server.go (development flag)
func (s *Server) loadTemplates() error {
    if os.Getenv("DEV_MODE") == "1" {
        // Load from disk for hot reload
        return template.ParseGlob("internal/web/templates/*.html")
    }
    // Load embedded (production)
    return template.ParseFS(templateFS, "templates/*.html")
}
```

**Usage**:
```bash
DEV_MODE=1 go run cmd/pkit/main.go serve
# Edit templates, refresh browser (no rebuild needed)
```

**Option 2**: Use `air` for auto-rebuild

```bash
# Install air
go install github.com/cosmtrek/air@latest

# Run with air (watches Go files)
air -- serve --port 8080
```

### Modifying Static Assets

CSS and JS are embedded, so rebuild after changes:

```bash
# Edit internal/web/static/style.css
# Rebuild
go build -o bin/pkit cmd/pkit/main.go

# Or use air for auto-rebuild
air -- serve
```

## Debugging

### Enable Request Logging

```go
// In middleware.go
func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        log.Printf("%s %s", r.Method, r.URL.Path)
        next.ServeHTTP(w, r)
        log.Printf("Completed in %v", time.Since(start))
    })
}
```

### Check Template Errors

```go
func (s *Server) renderTemplate(w http.ResponseWriter, name string, data interface{}) error {
    err := s.templates.ExecuteTemplate(w, name, data)
    if err != nil {
        log.Printf("Template error: %v", err)
        log.Printf("Data: %+v", data)
        http.Error(w, "Template error", 500)
    }
    return err
}
```

### Test with curl

```bash
# Test list endpoint
curl -v http://127.0.0.1:8080/

# Test detail endpoint
curl -v http://127.0.0.1:8080/prompts/fabric:summarize

# Test bookmark toggle
curl -X POST http://127.0.0.1:8080/api/bookmarks \
  -H "Content-Type: application/json" \
  -d '{"prompt_id":"fabric:summarize"}'

# Test tag update
curl -X POST http://127.0.0.1:8080/api/tags \
  -H "Content-Type: application/json" \
  -d '{"prompt_id":"fabric:summarize","tags":["dev","work"]}'
```

## Common Issues

### Issue: Port Already in Use

```
Error: listen tcp 127.0.0.1:8080: bind: address already in use
```

**Solution**:
```bash
# Find process using port
lsof -i :8080

# Kill process or use different port
go run cmd/pkit/main.go serve --port 8081
```

### Issue: Templates Not Found

```
Error: template: pattern matches no files: `templates/*.html`
```

**Solution**:
- Check `embed.go` has correct `//go:embed` directive
- Verify template files exist in `internal/web/templates/`
- Build tags correct (not in `//+build` ignored file)

### Issue: Static Assets 404

```
GET /static/style.css 404 Not Found
```

**Solution**:
- Check `embed.go` includes `//go:embed static/*`
- Verify files exist in `internal/web/static/`
- Check route registration: `http.StripPrefix("/static/", ...)`

### Issue: Bookmark/Tag Changes Not Persisting

**Solution**:
- Check file permissions on `~/.pkit/bookmarks.yml` and `tags.yml`
- Verify manager is using correct file paths
- Check for errors in server logs

## Performance Tips

### 1. Cache Prompts at Startup

```go
type Server struct {
    // ...
    cache struct {
        prompts   []models.Prompt
        bookmarks map[string]models.Bookmark
        tags      map[string][]string
        mu        sync.RWMutex
    }
}

func (s *Server) loadCache() error {
    s.cache.mu.Lock()
    defer s.cache.mu.Unlock()

    // Load all prompts
    s.cache.prompts = s.index.GetAllPrompts()

    // Load bookmarks
    bookmarks := s.bookmarkMgr.ListBookmarks()
    s.cache.bookmarks = make(map[string]models.Bookmark)
    for _, bm := range bookmarks {
        s.cache.bookmarks[bm.PromptID] = bm
    }

    // Load tags
    s.cache.tags = s.tagMgr.GetAllTags()

    return nil
}
```

### 2. Use Read Locks for Queries

```go
func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
    s.cache.mu.RLock()
    prompts := s.cache.prompts
    s.cache.mu.RUnlock()

    // ... filter and render
}
```

### 3. Reload Cache After Writes

```go
func (s *Server) handleBookmarkToggle(w http.ResponseWriter, r *http.Request) {
    // ... toggle bookmark in file

    // Reload cache
    s.loadCache()

    // ... return response
}
```

## Next Steps

1. Implement core server (server.go, handlers.go)
2. Create HTML templates (layout, list, detail)
3. Add static assets (CSS, minimal JS)
4. Implement JSON APIs (bookmark, tags)
5. Add middleware (logging, localhost check)
6. Write tests (unit, integration)
7. Test in browser
8. Add graceful shutdown
9. Performance optimize (caching)
10. Document deployment (see quickstart-deploy.md)

## Resources

- **Go html/template docs**: https://pkg.go.dev/html/template
- **Go embed docs**: https://pkg.go.dev/embed
- **Go http docs**: https://pkg.go.dev/net/http
- **Existing TUI reference**: `internal/tui/finder.go`
- **Spec**: `specs/003-web-interface/spec.md`
- **API contract**: `specs/003-web-interface/contracts/http-api.md`
- **Template contract**: `specs/003-web-interface/contracts/templates.md`
