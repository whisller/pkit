# Data Model: Local Web Interface

**Feature**: 003-web-interface
**Date**: 2025-12-28
**Status**: Design

## Overview

The web interface reuses all existing data models from the CLI. No new storage formats or database schemas are introduced. This document defines view models for rendering and describes how existing models are used in the web context.

## Existing Models (Reused)

These models are defined in `pkg/models` and used as-is:

### Prompt

**Source**: `pkg/models/prompt.go`

**Purpose**: Represents a prompt from any subscribed source

**Fields**:
```go
type Prompt struct {
    ID          string   // Format: "<source>:<name>" (e.g., "fabric:summarize")
    Name        string   // Prompt name within source
    Description string   // Short description
    Content     string   // Full prompt text
    Source      string   // Source ID (e.g., "fabric", "awesome-chatgpt-prompts")
    Tags        []string // User-assigned tags (from tags.yml)
}
```

**Usage in Web UI**:
- List view: Display ID, Name, Description, Source, Tags
- Detail view: Display all fields including full Content
- Search: Match against ID, Name, Description

---

### Bookmark

**Source**: `pkg/models/bookmark.go`

**Purpose**: Represents a user's saved prompt reference

**Fields**:
```go
type Bookmark struct {
    PromptID   string     // Reference to prompt: "<source>:<name>"
    Notes      string     // Optional user notes
    CreatedAt  time.Time  // When bookmarked
    UpdatedAt  time.Time  // Last modification
    UsageCount int        // Number of times accessed
    LastUsedAt *time.Time // Last access timestamp
}
```

**Storage**: `~/.pkit/bookmarks.yml`

**Usage in Web UI**:
- Check if prompt is bookmarked: lookup by PromptID
- Display bookmark indicator ([*]) in list view
- Toggle bookmark: add/remove from bookmarks.yml
- Increment UsageCount when prompt is viewed

---

### Source

**Source**: `pkg/models/source.go`

**Purpose**: Represents a subscribed prompt source (GitHub repo)

**Fields**:
```go
type Source struct {
    ID         string    // Unique identifier (e.g., "fabric")
    URL        string    // GitHub URL
    Type       string    // Source type (e.g., "fabric", "awesome-chatgpt-prompts")
    Path       string    // Local path in ~/.pkit/sources/
    LastUpdate time.Time // Last sync timestamp
}
```

**Storage**: `~/.pkit/config.yml` (subscribed sources list)

**Usage in Web UI**:
- Display source filter options
- Show source badge/tag on prompts
- Filter prompts by source

---

### Tag Association

**Storage**: `~/.pkit/tags.yml` (managed by `internal/tag` package)

**Format**:
```yaml
fabric:summarize:
  - documentation
  - work
awesome:code-review:
  - development
  - security
```

**Usage in Web UI**:
- Display tags on prompt cards
- Filter by tags (multi-select)
- Add/remove tags via tag editor dialog
- Auto-complete tag suggestions

## View Models (Web-Specific)

These models are defined in `internal/web` for rendering purposes only. They aggregate existing models with computed properties.

### PromptListItem

**Purpose**: Enriched prompt data for list view rendering

**Structure**:
```go
type PromptListItem struct {
    Prompt      models.Prompt  // Full prompt model
    Bookmarked  bool           // Is this prompt bookmarked?
    Tags        []string       // User tags for this prompt
}
```

**Construction**:
```go
func NewPromptListItem(prompt models.Prompt, bookmarks []models.Bookmark, tagManager *tag.Manager) PromptListItem {
    // Check if bookmarked
    bookmarked := false
    for _, bm := range bookmarks {
        if bm.PromptID == prompt.ID {
            bookmarked = true
            break
        }
    }

    // Get tags
    tags, _ := tagManager.GetTags(prompt.ID)

    return PromptListItem{
        Prompt:     prompt,
        Bookmarked: bookmarked,
        Tags:       tags,
    }
}
```

**Usage**:
- Rendered in list.html template
- Displayed as prompt cards with bookmark indicator and tags
- Used for paginated lists (50 items per page)

---

### PromptDetail

**Purpose**: Full prompt data for detail view

**Structure**:
```go
type PromptDetail struct {
    Prompt      models.Prompt  // Full prompt model
    Bookmarked  bool           // Is this prompt bookmarked?
    Tags        []string       // User tags
    Source      models.Source  // Source metadata
    Bookmark    *models.Bookmark // Full bookmark data (if bookmarked)
}
```

**Construction**:
```go
func NewPromptDetail(prompt models.Prompt, bookmark *models.Bookmark, tags []string, source models.Source) PromptDetail {
    return PromptDetail{
        Prompt:     prompt,
        Bookmarked: bookmark != nil,
        Tags:       tags,
        Source:     source,
        Bookmark:   bookmark,
    }
}
```

**Usage**:
- Rendered in detail.html template
- Shows full prompt content with metadata
- Displays bookmark notes if bookmarked
- Shows all tags with edit capability

---

### FilterState

**Purpose**: Represents active filter/search state

**Structure**:
```go
type FilterState struct {
    SearchQuery  string   // Search text
    SourceFilter string   // Selected source ID (empty = all)
    TagFilters   []string // Selected tags
    Bookmarked   bool     // Show only bookmarked?
    Page         int      // Current page (1-indexed)
    PerPage      int      // Items per page (default: 50)
}
```

**Construction**:
```go
func NewFilterStateFromQuery(query url.Values) FilterState {
    return FilterState{
        SearchQuery:  query.Get("search"),
        SourceFilter: query.Get("source"),
        TagFilters:   query["tags"],  // ?tags=tag1&tags=tag2
        Bookmarked:   query.Get("bookmarked") == "true",
        Page:         parseIntOrDefault(query.Get("page"), 1),
        PerPage:      50,
    }
}
```

**Usage**:
- Constructed from URL query parameters
- Used to filter prompts in handler
- Preserved in pagination links
- Used to maintain filter state across navigation

**Example URLs**:
```
/?search=code&source=fabric&tags=dev&bookmarked=true&page=2
```

---

### PaginatedResult

**Purpose**: Wrapper for paginated prompt lists

**Structure**:
```go
type PaginatedResult struct {
    Items      []PromptListItem // Current page items
    TotalItems int              // Total matching items
    TotalPages int              // Total pages
    CurrentPage int             // Current page (1-indexed)
    HasPrev    bool             // Has previous page?
    HasNext    bool             // Has next page?
    PrevPage   int              // Previous page number
    NextPage   int              // Next page number
}
```

**Construction**:
```go
func NewPaginatedResult(allItems []PromptListItem, page int, perPage int) PaginatedResult {
    totalItems := len(allItems)
    totalPages := (totalItems + perPage - 1) / perPage

    start := (page - 1) * perPage
    end := start + perPage
    if end > totalItems {
        end = totalItems
    }

    items := allItems[start:end]

    return PaginatedResult{
        Items:       items,
        TotalItems:  totalItems,
        TotalPages:  totalPages,
        CurrentPage: page,
        HasPrev:     page > 1,
        HasNext:     page < totalPages,
        PrevPage:    page - 1,
        NextPage:    page + 1,
    }
}
```

**Usage**:
- Rendered in list.html with pagination controls
- Shows "Showing X-Y of Z results"
- Generates prev/next page links with preserved filters

## Data Flow

### Read Operations

**List View (GET /)**:
```
1. Load all prompts from index
2. Load all bookmarks from ~/.pkit/bookmarks.yml
3. Load all tag associations from ~/.pkit/tags.yml
4. Apply filters (search, source, tags, bookmarked)
5. Create PromptListItem for each filtered prompt
6. Paginate results (50 per page)
7. Render list.html with PaginatedResult
```

**Detail View (GET /prompts/:id)**:
```
1. Load prompt from index by ID
2. Load bookmark (if exists) from bookmarks.yml
3. Load tags from tags.yml
4. Load source metadata
5. Create PromptDetail
6. Render detail.html
```

### Write Operations

**Bookmark Toggle (POST /api/bookmarks)**:
```
1. Parse request: {"prompt_id": "fabric:summarize"}
2. Check if already bookmarked
3. If bookmarked: Remove from bookmarks.yml
4. If not bookmarked: Add to bookmarks.yml with timestamps
5. Return: {"bookmarked": true/false}
```

**Tag Update (POST /api/tags)**:
```
1. Parse request: {"prompt_id": "fabric:summarize", "tags": ["dev", "work"]}
2. Remove existing tags for prompt_id
3. Add new tags to tags.yml
4. Return: {"success": true}
```

## Storage Access Patterns

### In-Memory Caching

**Strategy**: Load all data at server start, cache in memory, reload on write operations

**Cache Structure**:
```go
type DataCache struct {
    Prompts    []models.Prompt          // All prompts from index
    Bookmarks  map[string]models.Bookmark // PromptID -> Bookmark
    Tags       map[string][]string       // PromptID -> Tags
    Sources    map[string]models.Source  // SourceID -> Source
    LastLoaded time.Time
    mu         sync.RWMutex
}
```

**Cache Operations**:
- **Read**: Use RLock, access cached data
- **Write**: Use Lock, update YAML file, reload cache
- **Refresh**: Reload on startup, after writes, or on explicit refresh

**Concurrency**:
- Read-heavy workload: Multiple readers allowed
- Writes serialize: One writer at a time
- No race conditions between web UI and CLI (CLI modifies files, web reloads cache)

## Validation Rules

### Prompt ID Format

- Pattern: `^[a-z0-9-]+:[a-z0-9-]+$`
- Example: `fabric:summarize`, `awesome:code-review`
- Validated on all API endpoints accepting prompt_id

### Tag Format

- Pattern: `^[a-z0-9-]+$`
- Lowercase only
- No spaces
- Hyphens allowed
- Empty tags rejected

### Page Number

- Minimum: 1
- Maximum: totalPages (computed)
- Invalid page numbers redirect to page 1

## Error Handling

### Missing Prompt

- Scenario: Prompt ID not found in index
- Response: 404 Not Found with error page
- Message: "Prompt not found"

### Invalid Filter State

- Scenario: Malformed query parameters
- Response: Ignore invalid params, use defaults
- Example: `?page=invalid` → use page 1

### Write Failures

- Scenario: Cannot write to YAML file (permissions, disk full)
- Response: 500 Internal Server Error
- JSON: `{"error": "Failed to save changes"}`

## Performance Considerations

### Caching Strategy

- **Cache all prompts at startup**: 1000 prompts = ~50MB memory (acceptable)
- **Index search**: Use existing Bleve index for fast search (<500ms)
- **No database**: Avoid database overhead, YAML files sufficient for scale

### Optimization Targets

- Server start: < 2 seconds (load and cache all data)
- List view render: < 100ms (cached data, simple template)
- Search: < 500ms (Bleve index)
- Bookmark/tag operations: < 100ms (YAML write + cache reload)

## Migration Notes

**No migrations needed** - web interface uses existing storage:
- Existing `~/.pkit/bookmarks.yml` works as-is
- Existing `~/.pkit/tags.yml` works as-is
- Existing Bleve index in `~/.pkit/cache/` reused
- No schema changes required
