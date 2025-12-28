# HTTP API Contract: Local Web Interface

**Feature**: 003-web-interface
**Date**: 2025-12-28
**Status**: Design

## Overview

This document defines all HTTP endpoints for the pkit web interface. The API follows a hybrid approach:
- **Server-side rendering** for core navigation (HTML responses)
- **JSON APIs** for interactive features (bookmark, tag operations)
- **Progressive enhancement**: Works without JavaScript, enhanced with JS

## Design Principles

1. **RESTful where applicable**: Use standard HTTP methods and status codes
2. **Progressive enhancement**: HTML forms work without JS, AJAX enhances UX
3. **Stateless**: All state in URL query parameters (bookmarkable URLs)
4. **Idempotent reads**: GET requests don't modify data
5. **CORS not needed**: Single-origin (localhost only)

## Base Configuration

**Protocol**: HTTP (localhost only, no HTTPS needed)
**Host**: `127.0.0.1` (explicit IPv4 binding)
**Port**: Configurable (default: 8080)
**Base URL**: `http://127.0.0.1:8080`

## Endpoints

### HTML Endpoints (Server-Side Rendering)

#### GET /

**Purpose**: List prompts with filters and pagination

**Request**:
```
GET /?search=code&source=fabric&tags=dev&tags=work&bookmarked=true&page=2 HTTP/1.1
Host: 127.0.0.1:8080
```

**Query Parameters**:
| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `search` | string | No | Search query (matches ID, name, description) | `code review` |
| `source` | string | No | Filter by source ID | `fabric` |
| `tags` | string[] | No | Filter by tags (can repeat) | `?tags=dev&tags=work` |
| `bookmarked` | boolean | No | Show only bookmarked prompts | `true` or `false` |
| `page` | integer | No | Page number (1-indexed, default: 1) | `2` |

**Response (200 OK)**:
```html
Content-Type: text/html; charset=utf-8

<!DOCTYPE html>
<html>
<head>
    <title>pkit - Prompt Library</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <!-- Rendered list.html template -->
    <!-- Includes: filters, prompt cards, pagination -->
</body>
</html>
```

**Response (500 Internal Server Error)**:
```html
Content-Type: text/html; charset=utf-8

<!DOCTYPE html>
<html>
<head><title>Error</title></head>
<body>
    <h1>Error loading prompts</h1>
    <p>Unable to load prompt library. Please try again.</p>
</body>
</html>
```

**Behavior**:
- Invalid page numbers → redirect to page 1
- Empty results → show "No prompts found" message
- Filters preserved in pagination links

---

#### GET /prompts/:id

**Purpose**: View full prompt details

**Request**:
```
GET /prompts/fabric:summarize HTTP/1.1
Host: 127.0.0.1:8080
```

**Path Parameters**:
| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| `id` | string | Prompt ID (format: `source:name`) | `fabric:summarize` |

**Response (200 OK)**:
```html
Content-Type: text/html; charset=utf-8

<!DOCTYPE html>
<html>
<head>
    <title>fabric:summarize - pkit</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <!-- Rendered detail.html template -->
    <!-- Includes: full prompt content, metadata, tags, bookmark button -->
</body>
</html>
```

**Response (404 Not Found)**:
```html
Content-Type: text/html; charset=utf-8

<!DOCTYPE html>
<html>
<head><title>Prompt Not Found</title></head>
<body>
    <h1>Prompt Not Found</h1>
    <p>The prompt "fabric:summarize" does not exist.</p>
    <a href="/">← Back to prompt list</a>
</body>
</html>
```

**Behavior**:
- URL decoding for prompt IDs with special characters
- Increment usage count if bookmarked (fire-and-forget)
- Back navigation preserves previous filter state (via Referer header or history)

---

### JSON API Endpoints

#### POST /api/bookmarks

**Purpose**: Toggle bookmark status for a prompt

**Request**:
```http
POST /api/bookmarks HTTP/1.1
Host: 127.0.0.1:8080
Content-Type: application/json

{
  "prompt_id": "fabric:summarize"
}
```

**Request Body**:
```json
{
  "prompt_id": "string (required)"
}
```

**Response (200 OK) - Bookmark Added**:
```json
{
  "success": true,
  "bookmarked": true,
  "message": "Bookmarked"
}
```

**Response (200 OK) - Bookmark Removed**:
```json
{
  "success": true,
  "bookmarked": false,
  "message": "Bookmark removed"
}
```

**Response (400 Bad Request) - Invalid Input**:
```json
{
  "success": false,
  "error": "Invalid prompt_id"
}
```

**Response (404 Not Found) - Prompt Not Found**:
```json
{
  "success": false,
  "error": "Prompt not found"
}
```

**Response (500 Internal Server Error) - Save Failed**:
```json
{
  "success": false,
  "error": "Failed to save bookmark"
}
```

**Behavior**:
- Idempotent: Calling multiple times has same effect
- Creates bookmark with current timestamp if not exists
- Removes bookmark if already exists
- Updates LastUsedAt timestamp

---

#### POST /api/tags

**Purpose**: Update tags for a prompt

**Request**:
```http
POST /api/tags HTTP/1.1
Host: 127.0.0.1:8080
Content-Type: application/json

{
  "prompt_id": "fabric:summarize",
  "tags": ["dev", "documentation", "work"]
}
```

**Request Body**:
```json
{
  "prompt_id": "string (required)",
  "tags": ["string"] (required, can be empty array to clear all tags)
}
```

**Response (200 OK) - Tags Updated**:
```json
{
  "success": true,
  "tags": ["dev", "documentation", "work"],
  "message": "Tags updated"
}
```

**Response (200 OK) - Tags Cleared**:
```json
{
  "success": true,
  "tags": [],
  "message": "Tags cleared"
}
```

**Response (400 Bad Request) - Invalid Input**:
```json
{
  "success": false,
  "error": "Invalid prompt_id or tags"
}
```

**Response (404 Not Found) - Prompt Not Found**:
```json
{
  "success": false,
  "error": "Prompt not found"
}
```

**Response (500 Internal Server Error) - Save Failed**:
```json
{
  "success": false,
  "error": "Failed to save tags"
}
```

**Behavior**:
- **Replaces all tags** (not incremental)
- Validates tag format: lowercase, alphanumeric + hyphens only
- Empty array clears all tags for the prompt
- Deduplicates tags automatically
- Unused tags are automatically cleaned up

---

#### GET /api/search

**Purpose**: Real-time search for AJAX enhancement (optional)

**Request**:
```
GET /api/search?q=code+review HTTP/1.1
Host: 127.0.0.1:8080
```

**Query Parameters**:
| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `q` | string | Yes | Search query | `code review` |
| `limit` | integer | No | Max results (default: 20) | `10` |

**Response (200 OK)**:
```json
{
  "query": "code review",
  "results": [
    {
      "id": "fabric:code-review",
      "name": "code-review",
      "description": "Review code for quality and security",
      "source": "fabric",
      "bookmarked": true,
      "tags": ["dev", "security"]
    },
    {
      "id": "awesome:review-code",
      "name": "review-code",
      "description": "Comprehensive code review prompt",
      "source": "awesome-chatgpt-prompts",
      "bookmarked": false,
      "tags": []
    }
  ],
  "total": 2
}
```

**Response (400 Bad Request) - Missing Query**:
```json
{
  "error": "Missing search query"
}
```

**Behavior**:
- Uses Bleve search index for fast results
- Returns summary data (no full content)
- Respects filter state if provided (source, tags, bookmarked)
- Lightweight for autocomplete/live search

---

### Static Asset Endpoints

#### GET /static/:file

**Purpose**: Serve embedded static assets (CSS, JS)

**Request**:
```
GET /static/style.css HTTP/1.1
Host: 127.0.0.1:8080
```

**Response (200 OK) - CSS**:
```css
Content-Type: text/css; charset=utf-8
Cache-Control: public, max-age=3600

/* CSS content */
body { font-family: sans-serif; }
```

**Response (200 OK) - JavaScript**:
```javascript
Content-Type: application/javascript; charset=utf-8
Cache-Control: public, max-age=3600

// JS content
document.addEventListener('DOMContentLoaded', () => { ... });
```

**Response (404 Not Found)**:
```html
404 Not Found
```

**Behavior**:
- Assets embedded in binary via `go:embed`
- Cache headers for browser caching (1 hour)
- No dynamic generation
- Limited to: `style.css`, `app.js`

---

### Health Check Endpoint

#### GET /health

**Purpose**: Health check for monitoring

**Request**:
```
GET /health HTTP/1.1
Host: 127.0.0.1:8080
```

**Response (200 OK)**:
```json
{
  "status": "ok",
  "version": "1.0.0",
  "uptime_seconds": 3600
}
```

**Behavior**:
- Always returns 200 if server is running
- Used for readiness checks

---

## HTTP Methods & Status Codes

### Supported Methods

| Method | Usage |
|--------|-------|
| GET | Retrieve resources (HTML pages, JSON data) |
| POST | Modify resources (bookmarks, tags) |
| OPTIONS | CORS preflight (not needed for localhost) |

### Status Codes

| Code | Meaning | When Used |
|------|---------|-----------|
| 200 OK | Success | All successful requests |
| 400 Bad Request | Invalid input | Malformed JSON, invalid params |
| 404 Not Found | Resource not found | Prompt ID doesn't exist |
| 500 Internal Server Error | Server error | File I/O failure, unexpected error |

### Error Response Format

**HTML Errors** (for HTML endpoints):
```html
<!DOCTYPE html>
<html>
<head><title>Error</title></head>
<body>
    <h1>Error</h1>
    <p>Error message here</p>
    <a href="/">← Back home</a>
</body>
</html>
```

**JSON Errors** (for JSON endpoints):
```json
{
  "success": false,
  "error": "Error message here"
}
```

---

## Content Types

| Endpoint Pattern | Content-Type | Notes |
|------------------|--------------|-------|
| `/`, `/prompts/*` | `text/html; charset=utf-8` | Server-rendered HTML |
| `/api/*` | `application/json; charset=utf-8` | JSON API responses |
| `/static/*.css` | `text/css; charset=utf-8` | Embedded CSS |
| `/static/*.js` | `application/javascript; charset=utf-8` | Embedded JS |

---

## CORS Policy

**Not Required**: All requests originate from `http://127.0.0.1:8080` (same origin)

**Headers**: No CORS headers needed

---

## Security Headers

```http
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Content-Security-Policy: default-src 'self'; style-src 'self' 'unsafe-inline'; script-src 'self'
```

**Rationale**:
- Localhost-only, but defense-in-depth
- Prevent MIME sniffing
- Prevent clickjacking
- Restrict content sources
- Allow inline styles (for lipgloss-style rendering if needed)

---

## Rate Limiting

**Not Implemented**: Single-user local server, rate limiting not necessary

---

## Caching Strategy

### Browser Caching

| Resource Type | Cache-Control | Rationale |
|---------------|---------------|-----------|
| Static assets (CSS/JS) | `public, max-age=3600` | Assets rarely change, 1-hour cache |
| HTML pages | `no-cache` | Always fetch fresh data |
| JSON API responses | `no-cache` | Data changes frequently |

### Server-Side Caching

- All prompts cached in memory at startup
- Cache reloaded after write operations (bookmarks, tags)
- No HTTP cache (ETag, If-Modified-Since) - not needed for local use

---

## URL Structure Examples

```
# Home page (all prompts)
http://127.0.0.1:8080/

# Filtered view
http://127.0.0.1:8080/?source=fabric&tags=dev

# Search results
http://127.0.0.1:8080/?search=code+review

# Bookmarked prompts only
http://127.0.0.1:8080/?bookmarked=true

# Page 2 of search results
http://127.0.0.1:8080/?search=test&page=2

# Prompt detail
http://127.0.0.1:8080/prompts/fabric:summarize

# Combined filters
http://127.0.0.1:8080/?search=code&source=fabric&tags=dev&tags=security&bookmarked=true&page=1
```

---

## Progressive Enhancement Examples

### Without JavaScript

**Bookmark toggle**:
```html
<form method="POST" action="/api/bookmarks">
    <input type="hidden" name="prompt_id" value="fabric:summarize">
    <button type="submit">Bookmark</button>
</form>
```

**Search**:
```html
<form method="GET" action="/">
    <input type="text" name="search" placeholder="Search prompts...">
    <button type="submit">Search</button>
</form>
```

### With JavaScript Enhancement

**Bookmark toggle** (AJAX):
```javascript
fetch('/api/bookmarks', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({prompt_id: 'fabric:summarize'})
})
.then(res => res.json())
.then(data => {
    // Update UI without page reload
    updateBookmarkButton(data.bookmarked);
});
```

**Live search** (AJAX):
```javascript
let searchTimeout;
searchInput.addEventListener('input', (e) => {
    clearTimeout(searchTimeout);
    searchTimeout = setTimeout(() => {
        fetch(`/api/search?q=${encodeURIComponent(e.target.value)}`)
            .then(res => res.json())
            .then(data => updateSearchResults(data.results));
    }, 300); // Debounce 300ms
});
```

---

## Testing Endpoints

### Manual Testing

```bash
# Start server
pkit serve --port 8080

# Test list page
curl http://127.0.0.1:8080/

# Test detail page
curl http://127.0.0.1:8080/prompts/fabric:summarize

# Test bookmark toggle
curl -X POST http://127.0.0.1:8080/api/bookmarks \
  -H "Content-Type: application/json" \
  -d '{"prompt_id":"fabric:summarize"}'

# Test tag update
curl -X POST http://127.0.0.1:8080/api/tags \
  -H "Content-Type: application/json" \
  -d '{"prompt_id":"fabric:summarize","tags":["dev","work"]}'

# Test search
curl http://127.0.0.1:8080/api/search?q=code

# Test health check
curl http://127.0.0.1:8080/health
```

### Integration Test Scenarios

1. **Bookmark workflow**: List → Detail → Bookmark → Verify in list (shows [*])
2. **Tag workflow**: Detail → Edit tags → Save → Verify in list (shows tags)
3. **Search workflow**: Search → View result → Back → Search preserved
4. **Filter workflow**: Apply filters → Navigate pages → Filters preserved
5. **Error handling**: Invalid prompt ID → 404 page with back link

---

## Implementation Notes

### Handler Organization

```go
// internal/web/handlers.go
func (s *Server) registerRoutes() {
    // HTML endpoints
    s.mux.HandleFunc("/", s.handleList)
    s.mux.HandleFunc("/prompts/", s.handleDetail)

    // JSON API endpoints
    s.mux.HandleFunc("/api/bookmarks", s.handleBookmarkToggle)
    s.mux.HandleFunc("/api/tags", s.handleTagUpdate)
    s.mux.HandleFunc("/api/search", s.handleSearch)

    // Static assets
    s.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

    // Health check
    s.mux.HandleFunc("/health", s.handleHealth)
}
```

### Middleware Chain

```go
// Request logging → Localhost enforcement → CORS (no-op) → Handler
handler = loggingMiddleware(localhostOnlyMiddleware(handler))
```

---

## Future Considerations (Out of Scope for v1)

- WebSocket for real-time updates (not needed, page refresh sufficient)
- GraphQL API (REST is simpler for this use case)
- Authentication (single-user local server, not needed)
- HTTPS (localhost HTTP is sufficient)
- API versioning (v1 endpoint prefix, not needed initially)
