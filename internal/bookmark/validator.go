package bookmark

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
	"github.com/whisller/pkit/pkg/models"
)

const (
	// BookmarksFileName is the name of the bookmarks file
	BookmarksFileName = "bookmarks.yml"
)

// BookmarksFile represents the structure of the bookmarks YAML file
type BookmarksFile struct {
	Bookmarks []models.Bookmark `yaml:"bookmarks"`
}

// GetBookmarksPath returns the full path to the bookmarks file.
// It uses ~/.pkit/bookmarks.yml by default.
func GetBookmarksPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	bookmarksPath := filepath.Join(homeDir, ".pkit", BookmarksFileName)
	return bookmarksPath, nil
}

// ValidateBookmarksFile validates the bookmarks file for integrity and correctness.
// This implements FR-024: fail-safe startup check.
// Returns an error if the file is corrupted or contains invalid data.
// Returns nil if the file doesn't exist (empty state is OK).
func ValidateBookmarksFile(path string) error {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil // Empty state is OK
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("cannot read bookmarks file: %w", err)
	}

	// Parse YAML
	var bookmarksFile BookmarksFile
	if err := yaml.Unmarshal(data, &bookmarksFile); err != nil {
		return fmt.Errorf("corrupted bookmarks file at %s: %w\n\nRecovery options:\n  1. Manually fix YAML syntax errors\n  2. Restore from backup: ~/.pkit/backups/\n  3. Reset to empty state: pkit init --force (WARNING: destroys all bookmarks)", path, err)
	}

	// Validate each bookmark
	promptIDs := make(map[string]bool)
	for i, bm := range bookmarksFile.Bookmarks {
		// Check for duplicate prompt IDs
		if promptIDs[bm.PromptID] {
			return fmt.Errorf("duplicate prompt_id %q at index %d in bookmarks file", bm.PromptID, i)
		}
		promptIDs[bm.PromptID] = true

		// Validate bookmark using model validation
		if err := bm.Validate(); err != nil {
			return fmt.Errorf("invalid bookmark at index %d: %w", i, err)
		}
	}

	return nil
}

// ValidateBookmark validates a single bookmark.
// This is a convenience wrapper around models.Bookmark.Validate().
func ValidateBookmark(bm *models.Bookmark) error {
	if bm == nil {
		return fmt.Errorf("bookmark cannot be nil")
	}

	return bm.Validate()
}

// CheckBookmarkIntegrity performs a comprehensive integrity check on the bookmarks file.
// Returns a list of warnings (non-fatal issues) and any fatal errors.
func CheckBookmarkIntegrity(path string) (warnings []string, err error) {
	warnings = []string{}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return warnings, nil // Empty state is OK
	}

	// Read and parse
	data, err := os.ReadFile(path)
	if err != nil {
		return warnings, fmt.Errorf("cannot read bookmarks file: %w", err)
	}

	var bookmarksFile BookmarksFile
	if err := yaml.Unmarshal(data, &bookmarksFile); err != nil {
		return warnings, fmt.Errorf("corrupted bookmarks file: %w", err)
	}

	// Check for warnings (non-fatal issues)
	promptIDs := make(map[string]bool)
	for i, bm := range bookmarksFile.Bookmarks {
		// Check for duplicate prompt IDs (fatal error)
		if promptIDs[bm.PromptID] {
			return warnings, fmt.Errorf("duplicate prompt_id %q at index %d", bm.PromptID, i)
		}
		promptIDs[bm.PromptID] = true

		// Validate bookmark (fatal if invalid)
		if err := bm.Validate(); err != nil {
			return warnings, fmt.Errorf("invalid bookmark at index %d: %w", i, err)
		}

		// Check for potential warnings
		if bm.UsageCount == 0 && bm.LastUsedAt == nil {
			warnings = append(warnings, fmt.Sprintf("bookmark for prompt %q has never been used", bm.PromptID))
		}

		// Check if prompt ID is empty
		if bm.PromptID == "" {
			warnings = append(warnings, fmt.Sprintf("bookmark at index %d has empty prompt_id", i))
		}
	}

	return warnings, nil
}

// LoadBookmarks loads bookmarks from the file.
// Returns an empty slice if the file doesn't exist.
func LoadBookmarks() ([]models.Bookmark, error) {
	path, err := GetBookmarksPath()
	if err != nil {
		return nil, err
	}

	// If file doesn't exist, return empty slice
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return []models.Bookmark{}, nil
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read bookmarks file: %w", err)
	}

	// Parse YAML
	var bookmarksFile BookmarksFile
	if err := yaml.Unmarshal(data, &bookmarksFile); err != nil {
		return nil, fmt.Errorf("failed to parse bookmarks file: %w", err)
	}

	return bookmarksFile.Bookmarks, nil
}

// SaveBookmarks saves bookmarks to the file using atomic write.
func SaveBookmarks(bookmarks []models.Bookmark) error {
	path, err := GetBookmarksPath()
	if err != nil {
		return err
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create bookmarks directory: %w", err)
	}

	// Create bookmarks file structure
	bookmarksFile := BookmarksFile{
		Bookmarks: bookmarks,
	}

	// Marshal to YAML
	data, err := yaml.MarshalWithOptions(bookmarksFile, yaml.Indent(2))
	if err != nil {
		return fmt.Errorf("failed to marshal bookmarks: %w", err)
	}

	// Write to temporary file first (atomic write pattern)
	tmpFile := path + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write temporary bookmarks file: %w", err)
	}

	// Rename temporary file to actual bookmarks file (atomic operation)
	if err := os.Rename(tmpFile, path); err != nil {
		// Clean up temporary file on error (best effort)
		_ = os.Remove(tmpFile)
		return fmt.Errorf("failed to save bookmarks file: %w", err)
	}

	return nil
}

// ValidatePromptIDNotBookmarked checks if a prompt ID is not already bookmarked.
func ValidatePromptIDNotBookmarked(promptID string) error {
	bookmarks, err := LoadBookmarks()
	if err != nil {
		return fmt.Errorf("failed to load bookmarks: %w", err)
	}

	for _, bookmark := range bookmarks {
		if bookmark.PromptID == promptID {
			return fmt.Errorf("prompt '%s' is already bookmarked", promptID)
		}
	}

	return nil
}
