package bookmark

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/whisller/pkit/internal/config"
	"github.com/whisller/pkit/internal/index"
	"github.com/whisller/pkit/internal/source"
	"github.com/whisller/pkit/pkg/models"
)

// Resolver handles resolving aliases and prompt IDs to Prompt instances.
type Resolver struct {
	indexer   *index.Indexer
	bookmarks []models.Bookmark
}

// NewResolver creates a new prompt resolver.
func NewResolver(indexer *index.Indexer, bookmarks []models.Bookmark) *Resolver {
	return &Resolver{
		indexer:   indexer,
		bookmarks: bookmarks,
	}
}

// Resolve resolves an alias or prompt ID to a Prompt.
// Resolution order:
// 1. Check if it's a bookmark alias
// 2. Check if it's a prompt ID (source:name format)
// 3. Return error if not found
func (r *Resolver) Resolve(identifier string) (*models.Prompt, error) {
	// Try to resolve as bookmark alias first
	if r.bookmarks != nil {
		for _, bookmark := range r.bookmarks {
			if bookmark.Alias == identifier {
				// Found bookmark - fetch the actual prompt
				prompt, err := r.indexer.GetPromptByID(bookmark.PromptID)
				if err != nil {
					return nil, fmt.Errorf("bookmark '%s' points to non-existent prompt '%s': %w", identifier, bookmark.PromptID, err)
				}
				// Load content from source file
				if err := source.LoadPromptContent(prompt); err != nil {
					return nil, fmt.Errorf("failed to load prompt content: %w", err)
				}
				return prompt, nil
			}
		}
	}

	// Try to resolve as prompt ID (source:name format)
	if strings.Contains(identifier, ":") {
		prompt, err := r.indexer.GetPromptByID(identifier)
		if err != nil {
			return nil, fmt.Errorf("prompt not found: %s", identifier)
		}
		// Load content from source file
		if err := source.LoadPromptContent(prompt); err != nil {
			return nil, fmt.Errorf("failed to load prompt content: %w", err)
		}
		return prompt, nil
	}

	// Not found anywhere
	return nil, fmt.Errorf("no prompt or bookmark found for: %s", identifier)
}

// ResolveWithContext creates a resolver with loaded bookmarks and index.
func ResolveWithContext(identifier string) (*models.Prompt, error) {
	// Get index path
	indexBasePath, err := config.GetIndexPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get index path: %w", err)
	}

	// Open index
	indexPath := filepath.Join(indexBasePath, "prompts.bleve")
	indexer, err := index.NewIndexer(indexPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open index: %w", err)
	}
	defer indexer.Close()

	// Load bookmarks
	bookmarks, err := LoadBookmarks()
	if err != nil {
		// If bookmarks don't exist yet, that's okay - we can still resolve prompt IDs
		bookmarks = []models.Bookmark{}
	}

	// Create resolver and resolve
	resolver := NewResolver(indexer, bookmarks)
	return resolver.Resolve(identifier)
}
