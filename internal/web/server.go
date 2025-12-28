package web

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/whisller/pkit/internal/bookmark"
	"github.com/whisller/pkit/internal/index"
	"github.com/whisller/pkit/internal/tag"
	"github.com/whisller/pkit/pkg/models"
)

// Server represents the HTTP server for the web interface.
type Server struct {
	mux            *http.ServeMux
	server         *http.Server
	indexer        *index.Indexer
	bookmarkMgr    *bookmark.Manager
	tagMgr         *tag.Manager
	cache          dataCache
	startTime      time.Time
}

// dataCache holds user-specific cached data for fast access.
// Prompts are not cached; they're accessed directly from the indexer.
type dataCache struct {
	bookmarks map[string]models.Bookmark
	tags      map[string][]string
	mu        sync.RWMutex
}

// NewServer creates a new web server instance.
func NewServer(port int, indexPath string) (*Server, error) {
	// Initialize indexer
	indexer, err := index.NewIndexer(indexPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize indexer: %w", err)
	}

	mux := http.NewServeMux()

	s := &Server{
		mux:         mux,
		server: &http.Server{
			Addr:    fmt.Sprintf("127.0.0.1:%d", port),
			Handler: mux,
		},
		indexer:     indexer,
		bookmarkMgr: bookmark.NewManager(),
		tagMgr:      tag.NewManager(),
		startTime:   time.Now(),
	}

	// Initialize cache
	s.cache.bookmarks = make(map[string]models.Bookmark)
	s.cache.tags = make(map[string][]string)

	// Register routes
	s.registerRoutes()

	// Load initial cache
	if err := s.loadCache(); err != nil {
		return nil, fmt.Errorf("failed to load initial cache: %w", err)
	}

	return s, nil
}

// Start starts the HTTP server.
func (s *Server) Start() error {
	fmt.Printf("pkit web server starting...\n")
	fmt.Printf("Listening on http://%s\n", s.server.Addr)
	fmt.Printf("Press Ctrl+C to stop\n\n")

	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	fmt.Println("\nShutting down server...")
	return s.server.Shutdown(ctx)
}

// loadCache loads user-specific data into memory cache.
// Prompts are accessed directly from the indexer, not cached.
func (s *Server) loadCache() error {
	s.cache.mu.Lock()
	defer s.cache.mu.Unlock()

	// Load bookmarks
	bookmarks, err := s.bookmarkMgr.ListBookmarks()
	if err != nil {
		// Don't fail if bookmarks file doesn't exist yet
		bookmarks = []models.Bookmark{}
	}

	s.cache.bookmarks = make(map[string]models.Bookmark)
	for _, bm := range bookmarks {
		s.cache.bookmarks[bm.PromptID] = bm
	}

	// Load tags
	allTags, err := s.tagMgr.ListAllTags()
	if err != nil {
		// Don't fail if tags file doesn't exist yet
		allTags = []models.PromptTags{}
	}

	s.cache.tags = make(map[string][]string)
	for _, pt := range allTags {
		s.cache.tags[pt.PromptID] = pt.Tags
	}

	return nil
}

// registerRoutes registers all HTTP routes.
func (s *Server) registerRoutes() {
	// HTML endpoints - register FIRST
	s.mux.HandleFunc("/", s.handleList)
	s.mux.HandleFunc("/prompts/", s.handleDetail)

	// JSON API endpoints
	s.mux.HandleFunc("/api/bookmarks", s.handleBookmarkToggle)
	s.mux.HandleFunc("/api/tags", s.handleTagUpdate)
	s.mux.HandleFunc("/api/search", s.handleSearch)

	// Static assets (embedded) with cache headers and correct MIME types
	s.mux.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		// Set correct Content-Type based on file extension
		path := r.URL.Path
		if strings.HasSuffix(path, ".css") {
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
		} else if strings.HasSuffix(path, ".js") {
			w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		}

		w.Header().Set("Cache-Control", "public, max-age=3600")

		// Get the static subdirectory from embedded FS
		staticSubFS, err := fs.Sub(staticFS, "static")
		if err != nil {
			http.Error(w, "Static files not found", http.StatusInternalServerError)
			return
		}

		// Serve the file
		staticHandler := http.StripPrefix("/static/", http.FileServer(http.FS(staticSubFS)))
		staticHandler.ServeHTTP(w, r)
	})

	// Health check
	s.mux.HandleFunc("/health", s.handleHealth)

	// Apply middleware AFTER all routes are registered
	// (order matters: security -> localhost -> logging -> handler)
	handler := securityHeadersMiddleware(loggingMiddleware(localhostOnlyMiddleware(s.mux)))
	s.server.Handler = handler
}
