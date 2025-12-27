package source

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/whisller/pkit/internal/config"
	"github.com/whisller/pkit/pkg/models"
)

// LoadPromptContent loads the full content of a prompt from its source file.
// This is used instead of storing content in the index to save space.
func LoadPromptContent(prompt *models.Prompt) error {
	if prompt.Content != "" {
		// Content already loaded
		return nil
	}

	// Get the source for this prompt
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Find the source by ID
	var source *models.Source
	for _, src := range cfg.Sources {
		if src.ID == prompt.SourceID {
			source = &src
			break
		}
	}

	if source == nil {
		return fmt.Errorf("source not found: %s", prompt.SourceID)
	}

	// Determine full file path based on whether it's a cache path or source path
	var fullPath string

	if strings.HasPrefix(prompt.FilePath, "cache/") {
		// Cache path: resolve from ~/.pkit/
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		fullPath = filepath.Join(homeDir, ".pkit", prompt.FilePath)
	} else {
		// Source path: resolve from source.LocalPath
		fullPath = filepath.Join(source.LocalPath, prompt.FilePath)
	}

	// Read the file
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("failed to read prompt file %s: %w", fullPath, err)
	}

	// Set the content
	prompt.Content = string(content)
	return nil
}
