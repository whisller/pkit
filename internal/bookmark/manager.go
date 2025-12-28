package bookmark

import (
	"fmt"
	"strings"
	"time"

	"github.com/whisller/pkit/pkg/models"
)

// Manager handles CRUD operations on bookmarks.
type Manager struct{}

// NewManager creates a new bookmark manager.
func NewManager() *Manager {
	return &Manager{}
}

// AddBookmark adds a new bookmark to the collection.
// Returns error if prompt_id already bookmarked.
func (m *Manager) AddBookmark(bookmark models.Bookmark) error {
	// Load existing bookmarks
	bookmarks, err := LoadBookmarks()
	if err != nil {
		return fmt.Errorf("failed to load bookmarks: %w", err)
	}

	// Check if prompt_id already bookmarked
	for _, existing := range bookmarks {
		if existing.PromptID == bookmark.PromptID {
			return fmt.Errorf("prompt '%s' is already bookmarked", bookmark.PromptID)
		}
	}

	// Set creation time
	now := time.Now()
	bookmark.CreatedAt = now
	bookmark.UpdatedAt = now

	// Add bookmark
	bookmarks = append(bookmarks, bookmark)

	// Save
	if err := SaveBookmarks(bookmarks); err != nil {
		return fmt.Errorf("failed to save bookmarks: %w", err)
	}

	return nil
}

// UpdateBookmark updates an existing bookmark by prompt ID.
// Returns error if bookmark not found.
func (m *Manager) UpdateBookmark(promptID string, updater func(*models.Bookmark) error) error {
	// Load existing bookmarks
	bookmarks, err := LoadBookmarks()
	if err != nil {
		return fmt.Errorf("failed to load bookmarks: %w", err)
	}

	// Find and update bookmark
	found := false
	for i := range bookmarks {
		if bookmarks[i].PromptID == promptID {
			// Apply update
			if err := updater(&bookmarks[i]); err != nil {
				return fmt.Errorf("failed to update bookmark: %w", err)
			}
			bookmarks[i].UpdatedAt = time.Now()
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("bookmark for prompt '%s' not found", promptID)
	}

	// Save
	if err := SaveBookmarks(bookmarks); err != nil {
		return fmt.Errorf("failed to save bookmarks: %w", err)
	}

	return nil
}

// RemoveBookmark removes a bookmark by prompt ID.
// Returns error if bookmark not found.
func (m *Manager) RemoveBookmark(promptID string) error {
	// Load existing bookmarks
	bookmarks, err := LoadBookmarks()
	if err != nil {
		return fmt.Errorf("failed to load bookmarks: %w", err)
	}

	// Find and remove bookmark
	found := false
	newBookmarks := make([]models.Bookmark, 0, len(bookmarks))
	for _, bookmark := range bookmarks {
		if bookmark.PromptID == promptID {
			found = true
			continue
		}
		newBookmarks = append(newBookmarks, bookmark)
	}

	if !found {
		return fmt.Errorf("bookmark for prompt '%s' not found", promptID)
	}

	// Save
	if err := SaveBookmarks(newBookmarks); err != nil {
		return fmt.Errorf("failed to save bookmarks: %w", err)
	}

	return nil
}

// GetBookmark retrieves a bookmark by prompt ID.
// Returns error if bookmark not found.
func (m *Manager) GetBookmark(promptID string) (*models.Bookmark, error) {
	// Load existing bookmarks
	bookmarks, err := LoadBookmarks()
	if err != nil {
		return nil, fmt.Errorf("failed to load bookmarks: %w", err)
	}

	// Find bookmark
	for _, bookmark := range bookmarks {
		if bookmark.PromptID == promptID {
			return &bookmark, nil
		}
	}

	return nil, fmt.Errorf("bookmark for prompt '%s' not found", promptID)
}

// ListBookmarks returns all bookmarks.
func (m *Manager) ListBookmarks() ([]models.Bookmark, error) {
	// Load existing bookmarks
	bookmarks, err := LoadBookmarks()
	if err != nil {
		return nil, fmt.Errorf("failed to load bookmarks: %w", err)
	}

	return bookmarks, nil
}

// ParseTags parses a comma-separated string of tags.
// Tags are deduplicated, lowercased, and trimmed.
func ParseTags(tagString string) []string {
	if tagString == "" {
		return []string{}
	}

	// Split by comma
	parts := strings.Split(tagString, ",")

	// Use map to deduplicate
	tagMap := make(map[string]bool)
	for _, part := range parts {
		tag := strings.TrimSpace(strings.ToLower(part))
		if tag != "" {
			tagMap[tag] = true
		}
	}

	// Convert back to slice
	tags := make([]string, 0, len(tagMap))
	for tag := range tagMap {
		tags = append(tags, tag)
	}

	return tags
}

// IncrementUsage increments usage count and updates last used time for a bookmark.
func (m *Manager) IncrementUsage(promptID string) error {
	return m.UpdateBookmark(promptID, func(bookmark *models.Bookmark) error {
		bookmark.UsageCount++
		now := time.Now()
		bookmark.LastUsedAt = &now
		return nil
	})
}
