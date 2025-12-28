package tag

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/whisller/pkit/pkg/models"
)

const (
	// TagsFileName is the name of the tags file
	TagsFileName = "tags.yml"
)

// TagsFile represents the structure of the tags YAML file
type TagsFile struct {
	PromptTags []models.PromptTags `yaml:"prompt_tags"`
}

// GetTagsPath returns the full path to the tags file.
// It uses ~/.pkit/tags.yml by default.
func GetTagsPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	tagsPath := filepath.Join(homeDir, ".pkit", TagsFileName)
	return tagsPath, nil
}

// LoadTags loads prompt tags from the file.
// Returns an empty slice if the file doesn't exist.
func LoadTags() ([]models.PromptTags, error) {
	path, err := GetTagsPath()
	if err != nil {
		return nil, err
	}

	// If file doesn't exist, return empty slice
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return []models.PromptTags{}, nil
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read tags file: %w", err)
	}

	// Parse YAML
	var tagsFile TagsFile
	if err := yaml.Unmarshal(data, &tagsFile); err != nil {
		return nil, fmt.Errorf("failed to parse tags file: %w", err)
	}

	return tagsFile.PromptTags, nil
}

// SaveTags saves prompt tags to the file using atomic write.
func SaveTags(promptTags []models.PromptTags) error {
	path, err := GetTagsPath()
	if err != nil {
		return err
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create tags directory: %w", err)
	}

	// Create tags file structure
	tagsFile := TagsFile{
		PromptTags: promptTags,
	}

	// Marshal to YAML
	data, err := yaml.MarshalWithOptions(tagsFile, yaml.Indent(2))
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	// Write to temporary file first (atomic write pattern)
	tmpFile := path + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write temporary tags file: %w", err)
	}

	// Rename temporary file to actual tags file (atomic operation)
	if err := os.Rename(tmpFile, path); err != nil {
		// Clean up temporary file on error
		os.Remove(tmpFile)
		return fmt.Errorf("failed to save tags file: %w", err)
	}

	return nil
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
