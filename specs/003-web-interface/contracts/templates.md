# HTML Template Structure: Local Web Interface

**Feature**: 003-web-interface
**Date**: 2025-12-28
**Status**: Design

## Overview

This document defines the HTML template hierarchy and data contracts for server-side rendering using Go's `html/template` package. Templates follow a component-based structure with inheritance for consistency and maintainability.

## Template System

**Engine**: Go `html/template` (standard library)
**Location**: `internal/web/templates/` (embedded via `go:embed`)
**Syntax**: Go template syntax (`{{`, `}}`, `.`, `range`, `if`, etc.)

## Template Hierarchy

```
layout.html (base template - defines overall structure)
├── list.html (prompt list page - extends layout)
│   ├── components/filters.html (filter sidebar)
│   ├── components/prompt-card.html (individual prompt card)
│   └── components/pagination.html (page navigation)
├── detail.html (prompt detail page - extends layout)
│   ├── components/prompt-header.html (metadata header)
│   ├── components/prompt-content.html (full prompt text)
│   └── components/tag-editor.html (tag management UI)
└── error.html (error page - extends layout)
```

## Base Layout Template

### layout.html

**Purpose**: Provides HTML structure, navigation, and common elements

**Template**:
```html
{{define "layout"}}
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{block "title" .}}pkit - Prompt Library{{end}}</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <header class="site-header">
        <div class="container">
            <h1><a href="/">pkit</a></h1>
            <nav>
                <a href="/">Library</a>
                <a href="/?bookmarked=true">Bookmarks</a>
            </nav>
        </div>
    </header>

    <main class="site-main">
        {{block "content" .}}{{end}}
    </main>

    <footer class="site-footer">
        <div class="container">
            <p>pkit - Prompt Library Manager</p>
        </div>
    </footer>

    <script src="/static/app.js"></script>
</body>
</html>
{{end}}
```

**Data Contract**:
```go
// No data required at layout level
// Child templates provide data via their own structs
```

---

## Page Templates

### list.html

**Purpose**: Display paginated, filtered list of prompts

**Template**:
```html
{{define "title"}}Prompts ({{.TotalItems}}) - pkit{{end}}

{{define "content"}}
<div class="list-container">
    <!-- Filters Sidebar -->
    <aside class="filters-sidebar">
        {{template "filters" .}}
    </aside>

    <!-- Prompts List -->
    <section class="prompts-list">
        <header class="list-header">
            <h2>Prompts ({{.TotalItems}})</h2>
            {{if .FilterState.SearchQuery}}
                <p class="search-info">
                    Search results for "{{.FilterState.SearchQuery}}"
                    <a href="/">Clear search</a>
                </p>
            {{end}}
        </header>

        {{if .Items}}
            <div class="prompt-cards">
                {{range .Items}}
                    {{template "prompt-card" .}}
                {{end}}
            </div>

            {{template "pagination" .}}
        {{else}}
            <div class="empty-state">
                <p>No prompts found</p>
                {{if .FilterState.SearchQuery}}
                    <a href="/">Clear search and view all prompts</a>
                {{end}}
            </div>
        {{end}}
    </section>
</div>
{{end}}
```

**Data Contract**:
```go
type ListPageData struct {
    Items       []PromptListItem // Current page prompts
    TotalItems  int              // Total matching prompts
    TotalPages  int              // Total pages
    CurrentPage int              // Current page (1-indexed)
    HasPrev     bool             // Has previous page?
    HasNext     bool             // Has next page?
    PrevPage    int              // Previous page number
    NextPage    int              // Next page number
    FilterState FilterState      // Active filters
    Sources     []string         // Available sources (for filter options)
    AllTags     []string         // All available tags (for filter options)
}
```

---

### detail.html

**Purpose**: Display full prompt with metadata and actions

**Template**:
```html
{{define "title"}}{{.Prompt.ID}} - pkit{{end}}

{{define "content"}}
<div class="detail-container">
    <nav class="breadcrumb">
        <a href="/">← Back to prompts</a>
    </nav>

    {{template "prompt-header" .}}

    <div class="prompt-body">
        {{template "prompt-content" .}}

        <aside class="prompt-sidebar">
            <div class="actions-panel">
                <h3>Actions</h3>
                <button
                    class="btn btn-bookmark {{if .Bookmarked}}bookmarked{{end}}"
                    data-prompt-id="{{.Prompt.ID}}"
                    data-bookmarked="{{.Bookmarked}}">
                    {{if .Bookmarked}}★ Bookmarked{{else}}☆ Bookmark{{end}}
                </button>
                <button
                    class="btn btn-copy"
                    data-prompt-id="{{.Prompt.ID}}">
                    📋 Copy to Clipboard
                </button>
            </div>

            {{template "tag-editor" .}}

            <div class="metadata-panel">
                <h3>Info</h3>
                <dl>
                    <dt>Source</dt>
                    <dd><span class="badge badge-source">{{.Source.ID}}</span></dd>

                    {{if .Bookmarked}}
                        <dt>Bookmarked</dt>
                        <dd>{{.Bookmark.CreatedAt.Format "2006-01-02 15:04"}}</dd>

                        {{if .Bookmark.LastUsedAt}}
                            <dt>Last Used</dt>
                            <dd>{{.Bookmark.LastUsedAt.Format "2006-01-02 15:04"}}</dd>
                        {{end}}

                        <dt>Usage Count</dt>
                        <dd>{{.Bookmark.UsageCount}} times</dd>
                    {{end}}
                </dl>
            </div>
        </aside>
    </div>
</div>
{{end}}
```

**Data Contract**:
```go
type DetailPageData struct {
    Prompt     models.Prompt    // Full prompt data
    Bookmarked bool             // Is bookmarked?
    Tags       []string         // User tags
    Source     models.Source    // Source metadata
    Bookmark   *models.Bookmark // Bookmark data (if bookmarked)
    AllTags    []string         // All available tags (for autocomplete)
}
```

---

### error.html

**Purpose**: Display error messages with helpful actions

**Template**:
```html
{{define "title"}}Error - pkit{{end}}

{{define "content"}}
<div class="error-container">
    <div class="error-card">
        <h1>{{.Title}}</h1>
        <p class="error-message">{{.Message}}</p>

        {{if .Details}}
            <details class="error-details">
                <summary>Technical details</summary>
                <pre>{{.Details}}</pre>
            </details>
        {{end}}

        <div class="error-actions">
            <a href="/" class="btn">← Back to prompts</a>
            {{if .RetryAction}}
                <a href="{{.RetryAction}}" class="btn">Try again</a>
            {{end}}
        </div>
    </div>
</div>
{{end}}
```

**Data Contract**:
```go
type ErrorPageData struct {
    Title       string  // Error title (e.g., "Prompt Not Found")
    Message     string  // User-friendly message
    Details     string  // Technical details (optional)
    RetryAction string  // Retry URL (optional)
}
```

---

## Component Templates

### components/filters.html

**Purpose**: Filter sidebar with source, tag, and bookmark filters

**Template**:
```html
{{define "filters"}}
<div class="filters-panel">
    <h3>Filters</h3>

    <!-- Search -->
    <form method="GET" action="/" class="filter-search">
        <input
            type="text"
            name="search"
            placeholder="Search prompts..."
            value="{{.FilterState.SearchQuery}}"
            class="search-input">
        <button type="submit" class="btn-search">Search</button>
        {{if .FilterState.SearchQuery}}
            <a href="/" class="btn-clear">Clear</a>
        {{end}}
    </form>

    <!-- Source Filter -->
    <div class="filter-group">
        <h4>Sources</h4>
        <ul class="filter-list">
            <li>
                <a href="/" class="filter-item {{if not .FilterState.SourceFilter}}active{{end}}">
                    All sources
                </a>
            </li>
            {{range .Sources}}
                <li>
                    <a href="/?source={{.}}" class="filter-item {{if eq $.FilterState.SourceFilter .}}active{{end}}">
                        {{.}}
                    </a>
                </li>
            {{end}}
        </ul>
    </div>

    <!-- Tag Filter -->
    <div class="filter-group">
        <h4>Tags</h4>
        {{if .AllTags}}
            <ul class="filter-list">
                {{range .AllTags}}
                    <li>
                        <label class="filter-checkbox">
                            <input
                                type="checkbox"
                                name="tags"
                                value="{{.}}"
                                {{if contains $.FilterState.TagFilters .}}checked{{end}}
                                onchange="this.form.submit()">
                            {{.}}
                        </label>
                    </li>
                {{end}}
            </ul>
        {{else}}
            <p class="empty-note">No tags yet</p>
        {{end}}
    </div>

    <!-- Bookmark Filter -->
    <div class="filter-group">
        <label class="filter-checkbox">
            <input
                type="checkbox"
                name="bookmarked"
                value="true"
                {{if .FilterState.Bookmarked}}checked{{end}}
                onchange="this.form.submit()">
            Show only bookmarked
        </label>
    </div>
</div>
{{end}}
```

**Data Contract**: (same as ListPageData)

---

### components/prompt-card.html

**Purpose**: Individual prompt card in list view

**Template**:
```html
{{define "prompt-card"}}
<article class="prompt-card">
    <a href="/prompts/{{.Prompt.ID}}" class="card-link">
        <header class="card-header">
            <h3 class="card-title">
                {{if .Bookmarked}}
                    <span class="bookmark-indicator">[*]</span>
                {{end}}
                {{.Prompt.ID}}
            </h3>
            <span class="badge badge-source">{{.Prompt.Source}}</span>
        </header>

        <p class="card-description">{{.Prompt.Description}}</p>

        {{if .Tags}}
            <footer class="card-footer">
                <div class="card-tags">
                    {{range .Tags}}
                        <span class="tag">{{.}}</span>
                    {{end}}
                </div>
            </footer>
        {{end}}
    </a>
</article>
{{end}}
```

**Data Contract**:
```go
type PromptListItem struct {
    Prompt     models.Prompt // Full prompt data
    Bookmarked bool          // Is bookmarked?
    Tags       []string      // User tags
}
```

---

### components/pagination.html

**Purpose**: Page navigation controls

**Template**:
```html
{{define "pagination"}}
<nav class="pagination" aria-label="Page navigation">
    <div class="pagination-info">
        Showing {{.StartItem}}-{{.EndItem}} of {{.TotalItems}} prompts
    </div>

    {{if gt .TotalPages 1}}
        <div class="pagination-controls">
            {{if .HasPrev}}
                <a href="{{.PrevURL}}" class="btn btn-page">← Previous</a>
            {{else}}
                <span class="btn btn-page disabled">← Previous</span>
            {{end}}

            <span class="page-numbers">
                Page {{.CurrentPage}} of {{.TotalPages}}
            </span>

            {{if .HasNext}}
                <a href="{{.NextURL}}" class="btn btn-page">Next →</a>
            {{else}}
                <span class="btn btn-page disabled">Next →</span>
            {{end}}
        </div>
    {{end}}
</nav>
{{end}}
```

**Data Contract**:
```go
type PaginationData struct {
    TotalItems  int    // Total items
    TotalPages  int    // Total pages
    CurrentPage int    // Current page (1-indexed)
    HasPrev     bool   // Has previous page?
    HasNext     bool   // Has next page?
    PrevURL     string // Previous page URL (with filters)
    NextURL     string // Next page URL (with filters)
    StartItem   int    // First item number on page
    EndItem     int    // Last item number on page
}
```

---

### components/prompt-header.html

**Purpose**: Prompt metadata header in detail view

**Template**:
```html
{{define "prompt-header"}}
<header class="prompt-header">
    <h1 class="prompt-title">{{.Prompt.ID}}</h1>

    <div class="prompt-meta">
        <span class="badge badge-source">{{.Source.ID}}</span>

        {{if .Bookmarked}}
            <span class="badge badge-bookmark">★ Bookmarked</span>
        {{end}}

        {{range .Tags}}
            <span class="tag">{{.}}</span>
        {{end}}
    </div>

    {{if .Prompt.Description}}
        <p class="prompt-description">{{.Prompt.Description}}</p>
    {{end}}
</header>
{{end}}
```

**Data Contract**: (same as DetailPageData)

---

### components/prompt-content.html

**Purpose**: Display full prompt content

**Template**:
```html
{{define "prompt-content"}}
<div class="prompt-content">
    <pre><code>{{.Prompt.Content}}</code></pre>
</div>
{{end}}
```

**Data Contract**: (same as DetailPageData)

---

### components/tag-editor.html

**Purpose**: Tag management UI

**Template**:
```html
{{define "tag-editor"}}
<div class="tag-editor-panel">
    <h3>Tags</h3>

    {{if .Tags}}
        <div class="current-tags">
            {{range .Tags}}
                <span class="tag tag-editable">
                    {{.}}
                    <button class="tag-remove" data-tag="{{.}}">×</button>
                </span>
            {{end}}
        </div>
    {{else}}
        <p class="empty-note">No tags assigned</p>
    {{end}}

    <form class="tag-form" data-prompt-id="{{.Prompt.ID}}">
        <input
            type="text"
            class="tag-input"
            placeholder="Add tags (comma-separated)"
            value="{{join .Tags ", "}}"
            list="tag-suggestions">

        <datalist id="tag-suggestions">
            {{range .AllTags}}
                <option value="{{.}}">
            {{end}}
        </datalist>

        <button type="submit" class="btn btn-sm">Save Tags</button>
    </form>

    <p class="help-text">
        Separate tags with commas. Lowercase and hyphens only.
    </p>
</div>
{{end}}
```

**Data Contract**: (same as DetailPageData)

---

## Template Functions

Custom functions registered with `template.FuncMap`:

```go
var templateFuncs = template.FuncMap{
    // Join strings with separator
    "join": func(arr []string, sep string) string {
        return strings.Join(arr, sep)
    },

    // Check if slice contains value
    "contains": func(slice []string, item string) bool {
        for _, s := range slice {
            if s == item {
                return true
            }
        }
        return false
    },

    // Format time as relative (e.g., "2 hours ago")
    "timeAgo": func(t time.Time) string {
        duration := time.Since(t)
        if duration < time.Minute {
            return "just now"
        } else if duration < time.Hour {
            mins := int(duration.Minutes())
            return fmt.Sprintf("%d minutes ago", mins)
        } else if duration < 24*time.Hour {
            hours := int(duration.Hours())
            return fmt.Sprintf("%d hours ago", hours)
        } else {
            days := int(duration.Hours() / 24)
            return fmt.Sprintf("%d days ago", days)
        }
    },

    // Truncate string to max length
    "truncate": func(s string, maxLen int) string {
        if len(s) <= maxLen {
            return s
        }
        return s[:maxLen-3] + "..."
    },

    // Build URL with query params
    "buildURL": func(path string, params map[string]string) string {
        u, _ := url.Parse(path)
        q := u.Query()
        for k, v := range params {
            q.Set(k, v)
        }
        u.RawQuery = q.Encode()
        return u.String()
    },
}
```

---

## Template Loading

### Development Mode

```go
// Load templates from disk (hot reload)
func loadTemplates() (*template.Template, error) {
    return template.New("").
        Funcs(templateFuncs).
        ParseGlob("internal/web/templates/*.html")
}
```

### Production Mode (Embedded)

```go
//go:embed templates/*.html templates/**/*.html
var templateFS embed.FS

func loadTemplates() (*template.Template, error) {
    return template.New("").
        Funcs(templateFuncs).
        ParseFS(templateFS, "templates/*.html", "templates/**/*.html")
}
```

---

## Rendering Pipeline

```go
func (s *Server) renderTemplate(w http.ResponseWriter, name string, data interface{}) error {
    // Set content type
    w.Header().Set("Content-Type", "text/html; charset=utf-8")

    // Execute template
    err := s.templates.ExecuteTemplate(w, name, data)
    if err != nil {
        log.Printf("Template error: %v", err)
        http.Error(w, "Internal Server Error", 500)
        return err
    }

    return nil
}

// Usage in handler
func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
    data := s.loadListData(r)
    s.renderTemplate(w, "list.html", data)
}
```

---

## CSS Classes Reference

### Layout
- `.site-header`, `.site-main`, `.site-footer`: Page structure
- `.container`: Centered content wrapper

### List View
- `.list-container`: Main container (flexbox: sidebar + content)
- `.filters-sidebar`: Left sidebar with filters
- `.prompts-list`: Right content area

### Cards
- `.prompt-card`: Individual prompt card
- `.card-header`, `.card-title`, `.card-description`, `.card-footer`: Card sections
- `.card-tags`: Tag container in card footer

### Badges & Tags
- `.badge`: General badge (source, bookmark)
- `.badge-source`: Source badge
- `.badge-bookmark`: Bookmark badge
- `.tag`: Tag pill
- `.tag-editable`: Editable tag with remove button

### Forms & Inputs
- `.filter-search`: Search form
- `.search-input`: Search text input
- `.tag-form`: Tag editing form
- `.tag-input`: Tag text input
- `.filter-checkbox`: Checkbox with label

### Buttons
- `.btn`: Base button class
- `.btn-primary`, `.btn-secondary`: Button variants
- `.btn-bookmark`: Bookmark toggle button
- `.btn-copy`: Copy to clipboard button
- `.btn-page`: Pagination button

### State
- `.active`: Active filter/navigation item
- `.disabled`: Disabled button/link
- `.bookmarked`: Bookmarked state
- `.empty-state`: Empty results message

---

## Accessibility

### ARIA Labels
```html
<nav aria-label="Page navigation">
<button aria-label="Bookmark this prompt">
<form role="search">
```

### Semantic HTML
- Use `<header>`, `<main>`, `<footer>`, `<nav>`, `<article>`
- Proper heading hierarchy (`<h1>` → `<h2>` → `<h3>`)
- Form labels associated with inputs

### Keyboard Navigation
- All interactive elements focusable
- Tab order follows visual flow
- Focus styles visible

---

## Testing Templates

```go
func TestTemplateRendering(t *testing.T) {
    tmpl, err := loadTemplates()
    if err != nil {
        t.Fatalf("Failed to load templates: %v", err)
    }

    // Test list template
    data := ListPageData{
        Items:      []PromptListItem{{...}},
        TotalItems: 1,
        // ...
    }

    var buf bytes.Buffer
    err = tmpl.ExecuteTemplate(&buf, "list.html", data)
    if err != nil {
        t.Fatalf("Failed to render list template: %v", err)
    }

    html := buf.String()
    if !strings.Contains(html, "Prompts (1)") {
        t.Error("Expected prompt count in output")
    }
}
```

---

## Browser Compatibility

**Target**: Modern browsers (last 2 versions)
- Chrome/Edge (Chromium)
- Firefox
- Safari

**No polyfills needed**: HTML5/CSS3 features used are widely supported

**Progressive enhancement**: Core functionality works without JavaScript, enhanced with JS
