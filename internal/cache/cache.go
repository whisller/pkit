package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

// GetCachePath returns the base cache directory path.
// Default: ~/.pkit/cache/
func GetCachePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	cachePath := filepath.Join(homeDir, ".pkit", "cache")
	return cachePath, nil
}

// GetSourceCachePath returns the cache directory for a specific source.
// Example: ~/.pkit/cache/awesome/
func GetSourceCachePath(sourceID string) (string, error) {
	basePath, err := GetCachePath()
	if err != nil {
		return "", err
	}

	sourceCachePath := filepath.Join(basePath, sourceID)
	return sourceCachePath, nil
}

// EnsureSourceCacheDir creates the cache directory for a source if it doesn't exist.
func EnsureSourceCacheDir(sourceID string) error {
	cachePath, err := GetSourceCachePath(sourceID)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(cachePath, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	return nil
}

// WritePromptToCache writes a prompt's content to a cache file.
// Returns the relative cache path (e.g., "awesome/linux-terminal.md")
func WritePromptToCache(sourceID, promptName, content string) (string, error) {
	// Ensure cache directory exists
	if err := EnsureSourceCacheDir(sourceID); err != nil {
		return "", err
	}

	// Get cache path
	cachePath, err := GetSourceCachePath(sourceID)
	if err != nil {
		return "", err
	}

	// Write content to file
	filename := fmt.Sprintf("%s.md", promptName)
	fullPath := filepath.Join(cachePath, filename)

	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write cache file: %w", err)
	}

	// Return relative path from .pkit directory
	relativePath := filepath.Join("cache", sourceID, filename)
	return relativePath, nil
}

// ComputeFileHash computes SHA256 hash of a file for change detection.
func ComputeFileHash(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:]), nil
}

// ClearSourceCache removes all cached files for a source.
// Used when re-indexing or updating.
func ClearSourceCache(sourceID string) error {
	cachePath, err := GetSourceCachePath(sourceID)
	if err != nil {
		return err
	}

	// Remove directory and all contents
	if err := os.RemoveAll(cachePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clear cache: %w", err)
	}

	return nil
}
