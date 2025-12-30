package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/whisller/pkit/internal/alias"
	"github.com/whisller/pkit/internal/bookmark"
	"github.com/whisller/pkit/internal/index"
	"github.com/whisller/pkit/internal/tag"
	"github.com/whisller/pkit/pkg/models"
)

// PromptListItem represents a prompt in the list view.
type PromptListItem struct {
	Prompt     models.Prompt
	Bookmarked bool
	Tags       []string
}

// PromptDetail represents a prompt in detail view.
type PromptDetail struct {
	Prompt     models.Prompt
	Bookmarked bool
	Tags       []string
	Bookmark   *models.Bookmark
}

// FilterState represents active filters.
type FilterState struct {
	SearchQuery   string
	SourceFilters []string
	TagFilters    []string
	Bookmarked    bool
	Page          int
	PerPage       int
}

// PaginatedResult represents paginated list results.
type PaginatedResult struct {
	Items       []PromptListItem
	TotalItems  int
	TotalPages  int
	CurrentPage int
	HasPrev     bool
	HasNext     bool
	PrevPage    int
	NextPage    int
}

// templates holds parsed templates.
var templates *template.Template

// initTemplates initializes templates from embedded FS.
func initTemplates() error {
	var err error
	funcMap := template.FuncMap{
		"sub": func(a, b int) int {
			return a - b
		},
	}
	templates, err = template.New("").Funcs(funcMap).ParseFS(templateFS, "templates/*.html", "templates/components/*.html")
	if err != nil {
		return fmt.Errorf("failed to parse templates: %w", err)
	}
	return nil
}

// renderTemplate renders a template with data.
func (s *Server) renderTemplate(w http.ResponseWriter, name string, data interface{}) error {
	if templates == nil {
		if err := initTemplates(); err != nil {
			return err
		}
	}

	err := templates.ExecuteTemplate(w, name, data)
	if err != nil {
		// Template rendering errors are rare in production
		http.Error(w, "Template rendering error", http.StatusInternalServerError)
	}
	return err
}

// handleHealth handles GET /health.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{
		"status":         "ok",
		"version":        "1.0.0",
		"uptime_seconds": int(time.Since(s.startTime).Seconds()),
	}

	// JSON encoding to HTTP response rarely fails
	_ = json.NewEncoder(w).Encode(response)
}

// handleList handles GET / - list prompts with filters.
func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	filters := parseFilters(r.URL.Query())

	// Use indexer to search (reuses CLI search logic)
	// NOTE: Don't filter by tags in indexer - user tags are in tags.yml, not in the index
	searchOpts := index.SearchOptions{
		Query:      filters.SearchQuery,
		SourceID:   "", // We'll filter by sources manually if multiple selected
		MaxResults: 10000, // Get all results, we'll paginate in memory
	}

	// If only one source selected, use indexer's source filter for efficiency
	if len(filters.SourceFilters) == 1 {
		searchOpts.SourceID = filters.SourceFilters[0]
	}

	results, err := s.indexer.Search(searchOpts)
	if err != nil {
		http.Error(w, "Search failed", http.StatusInternalServerError)
		return
	}

	// Convert to list items and apply user-specific filters (bookmarks, tags, sources)
	s.cache.mu.RLock()
	items := []PromptListItem{}
	for _, result := range results {
		// Apply source filter if multiple sources selected
		if len(filters.SourceFilters) > 1 {
			matchesSource := false
			for _, sourceFilter := range filters.SourceFilters {
				if result.Prompt.SourceID == sourceFilter {
					matchesSource = true
					break
				}
			}
			if !matchesSource {
				continue
			}
		}

		// Apply bookmarked filter (user-specific, not in index)
		_, isBookmarked := s.cache.bookmarks[result.Prompt.ID]
		if filters.Bookmarked && !isBookmarked {
			continue
		}

		// Get tags for this prompt
		tags := s.cache.tags[result.Prompt.ID]
		if tags == nil {
			tags = []string{}
		} else {
			tagsCopy := make([]string, len(tags))
			copy(tagsCopy, tags)
			sort.Strings(tagsCopy)
			tags = tagsCopy
		}

		// Apply tag filters (user tags are not in index, filter manually)
		if len(filters.TagFilters) > 0 {
			hasAllTags := true
			for _, filterTag := range filters.TagFilters {
				found := false
				for _, tag := range tags {
					if tag == filterTag {
						found = true
						break
					}
				}
				if !found {
					hasAllTags = false
					break
				}
			}
			if !hasAllTags {
				continue
			}
		}

		items = append(items, PromptListItem{
			Prompt:     result.Prompt,
			Bookmarked: isBookmarked,
			Tags:       tags,
		})
	}
	s.cache.mu.RUnlock()

	// Paginate results
	paginated := paginate(items, filters.Page, filters.PerPage)

	// Get unique sources and tags for filter UI
	// Build from ALL results (not filtered items) so all options remain visible
	sources := make(map[string]bool)
	allTags := make(map[string]bool)

	s.cache.mu.RLock()
	for _, result := range results {
		sources[result.Prompt.SourceID] = true
		// Get tags for this prompt
		tags := s.cache.tags[result.Prompt.ID]
		for _, tag := range tags {
			allTags[tag] = true
		}
	}
	s.cache.mu.RUnlock()

	// Convert maps to sorted slices
	sourceList := make([]string, 0, len(sources))
	for source := range sources {
		sourceList = append(sourceList, source)
	}
	sort.Strings(sourceList) // Sort alphabetically

	tagList := make([]string, 0, len(allTags))
	for tag := range allTags {
		tagList = append(tagList, tag)
	}
	sort.Strings(tagList) // Sort alphabetically

	// Prepare template data
	data := map[string]interface{}{
		"Items":         paginated.Items,
		"TotalItems":    paginated.TotalItems,
		"TotalPages":    paginated.TotalPages,
		"CurrentPage":   paginated.CurrentPage,
		"HasPrev":       paginated.HasPrev,
		"HasNext":       paginated.HasNext,
		"PrevPage":      paginated.PrevPage,
		"NextPage":      paginated.NextPage,
		"Filters":       filters,
		"Sources":       sourceList,
		"Tags":          tagList,
	}

	// Render template (errors handled by renderTemplate)
	_ = s.renderTemplate(w, "list", data)
}

// handleDetail handles GET /prompts/:id - view prompt details.
func (s *Server) handleDetail(w http.ResponseWriter, r *http.Request) {
	// Extract prompt ID from URL path
	// Path format: /prompts/source:name
	promptID := strings.TrimPrefix(r.URL.Path, "/prompts/")

	if promptID == "" {
		http.Error(w, "Prompt ID required", http.StatusBadRequest)
		return
	}

	// URL decode the prompt ID (handles encoded colons, etc.)
	var err error
	promptID, err = url.QueryUnescape(promptID)
	if err != nil {
		http.Error(w, "Invalid prompt ID", http.StatusBadRequest)
		return
	}

	// Create resolver with server's indexer to load content (same as CLI)
	aliases, _ := alias.LoadAliases()
	bookmarks, _ := bookmark.LoadBookmarks()
	resolver := bookmark.NewResolver(s.indexer, aliases, bookmarks)

	// Use resolver's logic to load content
	prompt, err := resolver.Resolve(promptID)
	if err != nil {
		// Render 404 error page
		data := map[string]interface{}{
			"Error":    "Prompt Not Found",
			"Message":  fmt.Sprintf("The prompt '%s' does not exist or could not be loaded.", promptID),
			"BackLink": "/",
		}
		w.WriteHeader(http.StatusNotFound)
		// Template rendering error on error page - ignore
		_ = s.renderTemplate(w, "error", data)
		return
	}

	// Get bookmark status
	s.cache.mu.RLock()
	bookmark, isBookmarked := s.cache.bookmarks[promptID]
	tags := s.cache.tags[promptID]
	s.cache.mu.RUnlock()

	if tags == nil {
		tags = []string{}
	} else {
		// Sort tags alphabetically
		tagsCopy := make([]string, len(tags))
		copy(tagsCopy, tags)
		sort.Strings(tagsCopy)
		tags = tagsCopy
	}

	// Prepare detail data
	detail := PromptDetail{
		Prompt:     *prompt,
		Bookmarked: isBookmarked,
		Tags:       tags,
	}

	if isBookmarked {
		detail.Bookmark = &bookmark
	}

	// Prepare template data
	data := map[string]interface{}{
		"Detail": detail,
	}

	// Render template (errors handled by renderTemplate)
	_ = s.renderTemplate(w, "detail", data)
}

// handleBookmarkToggle handles POST /api/bookmarks - toggle bookmark.
func (s *Server) handleBookmarkToggle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		// JSON encoding error response rarely fails
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Method not allowed",
		})
		return
	}

	// Parse request body
	var req struct {
		PromptID string `json:"prompt_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// JSON encoding error response rarely fails
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	if req.PromptID == "" {
		// JSON encoding error response rarely fails
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "prompt_id is required",
		})
		return
	}

	// Check if already bookmarked
	_, err := s.bookmarkMgr.GetBookmark(req.PromptID)
	isBookmarked := err == nil

	var bookmarked bool
	var message string

	if isBookmarked {
		// Remove bookmark
		if err := s.bookmarkMgr.RemoveBookmark(req.PromptID); err != nil {
			// JSON encoding error response rarely fails
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   fmt.Sprintf("Failed to remove bookmark: %v", err),
			})
			return
		}
		bookmarked = false
		message = "Bookmark removed"
	} else {
		// Add bookmark (same as CLI)
		bm := models.Bookmark{
			PromptID: req.PromptID,
		}

		// Validate bookmark (same as CLI)
		if err := bookmark.ValidateBookmark(&bm); err != nil {
			// JSON encoding error response rarely fails
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   fmt.Sprintf("Invalid bookmark: %v", err),
			})
			return
		}

		if err := s.bookmarkMgr.AddBookmark(bm); err != nil {
			// JSON encoding error response rarely fails
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   fmt.Sprintf("Failed to add bookmark: %v", err),
			})
			return
		}
		bookmarked = true
		message = "Bookmarked"
	}

	// Reload cache (errors ignored - cache will be reloaded on next request)
	_ = s.loadCache()

	// Return success - JSON encoding to HTTP response rarely fails
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"bookmarked": bookmarked,
		"message":    message,
	})
}

// handleTagUpdate handles POST /api/tags - update tags.
func (s *Server) handleTagUpdate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		// JSON encoding error response rarely fails
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Method not allowed",
		})
		return
	}

	// Parse request body
	var req struct {
		PromptID string   `json:"prompt_id"`
		Tags     []string `json:"tags"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// JSON encoding error response rarely fails
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	if req.PromptID == "" {
		// JSON encoding error response rarely fails
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "prompt_id is required",
		})
		return
	}

	// Parse tags using CLI's tag.ParseTags() logic
	tagString := strings.Join(req.Tags, ",")
	uniqueTags := tag.ParseTags(tagString)

	// Remove all existing tags first
	existingTags, _ := s.tagMgr.GetTags(req.PromptID)
	if len(existingTags) > 0 {
		// RemoveTags with empty slice removes all tags, errors can be ignored (tags may not exist)
		_ = s.tagMgr.RemoveTags(req.PromptID, []string{})
	}

	// Add new tags if any
	var message string
	if len(uniqueTags) > 0 {
		if err := s.tagMgr.AddTags(req.PromptID, uniqueTags); err != nil {
			// JSON encoding error response rarely fails
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   fmt.Sprintf("Failed to update tags: %v", err),
			})
			return
		}
		message = "Tags updated"
	} else {
		message = "Tags cleared"
	}

	// Reload cache (errors ignored - cache will be reloaded on next request)
	_ = s.loadCache()

	// Return success - JSON encoding to HTTP response rarely fails
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"tags":    uniqueTags,
		"message": message,
	})
}

// handleSearch handles GET /api/search - real-time search.
func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	// Stub - optional feature for AJAX
	w.Header().Set("Content-Type", "application/json")
	// Write to HTTP response rarely fails
	_, _ = w.Write([]byte(`{"results": [], "total": 0}`))
}

// parseFilters extracts filter state from query parameters.
func parseFilters(query url.Values) FilterState {
	page, _ := strconv.Atoi(query.Get("page"))
	if page < 1 {
		page = 1
	}

	return FilterState{
		SearchQuery:   query.Get("search"),
		SourceFilters: query["sources"],
		TagFilters:    query["tags"],
		Bookmarked:    query.Get("bookmarked") == "true",
		Page:          page,
		PerPage:       50,
	}
}


// paginate paginates a list of items.
func paginate(items []PromptListItem, page int, perPage int) PaginatedResult {
	totalItems := len(items)
	totalPages := (totalItems + perPage - 1) / perPage

	if totalPages == 0 {
		totalPages = 1
	}

	if page > totalPages {
		page = totalPages
	}

	start := (page - 1) * perPage
	end := start + perPage

	if start > totalItems {
		start = totalItems
	}
	if end > totalItems {
		end = totalItems
	}

	pageItems := items[start:end]

	return PaginatedResult{
		Items:       pageItems,
		TotalItems:  totalItems,
		TotalPages:  totalPages,
		CurrentPage: page,
		HasPrev:     page > 1,
		HasNext:     page < totalPages,
		PrevPage:    page - 1,
		NextPage:    page + 1,
	}
}
