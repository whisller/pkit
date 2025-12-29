package bookmark

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/whisller/pkit/internal/alias"
	"github.com/whisller/pkit/internal/config"
	"github.com/whisller/pkit/internal/index"
	"github.com/whisller/pkit/internal/source"
	"github.com/whisller/pkit/pkg/models"
)

// Resolver handles resolving aliases and prompt IDs to Prompt instances.
type Resolver struct {
	indexer   *index.Indexer
	aliases   []models.Alias
	bookmarks []models.Bookmark
}

// NewResolver creates a new prompt resolver.
func NewResolver(indexer *index.Indexer, aliases []models.Alias, bookmarks []models.Bookmark) *Resolver {
	return &Resolver{
		indexer:   indexer,
		aliases:   aliases,
		bookmarks: bookmarks,
	}
}

// Resolve resolves an alias or prompt ID to a Prompt.
// Resolution order:
// 1. Check if it's an alias
// 2. Check if it's a prompt ID (source:name format)
// 3. Return error if not found
func (r *Resolver) Resolve(identifier string) (*models.Prompt, error) {
	var promptID string

	// Try to resolve as alias first
	if r.aliases != nil {
		for _, a := range r.aliases {
			if a.Name == identifier {
				promptID = a.PromptID
				break
			}
		}
	}

	// If not an alias, check if it's a prompt ID (source:name format)
	if promptID == "" {
		if strings.Contains(identifier, ":") {
			promptID = identifier
		} else {
			return nil, fmt.Errorf("no alias or prompt found for: %s", identifier)
		}
	}

	// Fetch the prompt from index
	prompt, err := r.indexer.GetPromptByID(promptID)
	if err != nil {
		return nil, fmt.Errorf("prompt not found: %s", promptID)
	}

	// Load content from source file
	if err := source.LoadPromptContent(prompt); err != nil {
		return nil, fmt.Errorf("failed to load prompt content: %w", err)
	}

	// Track bookmark usage if this prompt is bookmarked
	if r.bookmarks != nil {
		for _, bookmark := range r.bookmarks {
			if bookmark.PromptID == promptID {
				manager := NewManager()
				_ = manager.IncrementUsage(promptID)
				break
			}
		}
	}

	return prompt, nil
}

// ResolveWithContext creates a resolver with loaded aliases, bookmarks, and index.
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
	defer func() {
		if closeErr := indexer.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close index: %w", closeErr)
		}
	}()

	// Load aliases
	aliases, err := alias.LoadAliases()
	if err != nil {
		// If aliases don't exist yet, that's okay
		aliases = []models.Alias{}
	}

	// Load bookmarks
	bookmarks, err := LoadBookmarks()
	if err != nil {
		// If bookmarks don't exist yet, that's okay
		bookmarks = []models.Bookmark{}
	}

	// Create resolver and resolve
	resolver := NewResolver(indexer, aliases, bookmarks)
	return resolver.Resolve(identifier)
}
