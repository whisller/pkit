package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/whisller/pkit/pkg/models"
)

// MarkdownParser parses generic Markdown files from any repository.
// This is a fallback parser for repositories that don't match Fabric or awesome-chatgpt formats.
type MarkdownParser struct{}

// NewMarkdownParser creates a new generic markdown parser instance.
func NewMarkdownParser() *MarkdownParser {
	return &MarkdownParser{}
}

// Name returns the parser name.
func (p *MarkdownParser) Name() string {
	return "markdown"
}

// CanParse checks if the source path contains any .md files.
// This always returns true as a fallback, but should be checked last.
func (p *MarkdownParser) CanParse(sourcePath string) bool {
	// Check if directory contains any .md files
	hasMarkdown := false
	_ = filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if filepath.Ext(path) == ".md" {
			baseName := strings.ToLower(filepath.Base(path))
			// Skip common non-prompt files
			if baseName != "readme.md" && baseName != "license.md" && baseName != "contributing.md" {
				hasMarkdown = true
				return filepath.SkipAll // Found at least one, stop walking
			}
		}
		return nil
	})
	return hasMarkdown
}

// ParsePrompts extracts all prompts from markdown files.
func (p *MarkdownParser) ParsePrompts(source *models.Source) ([]models.Prompt, error) {
	var prompts []models.Prompt

	// Walk source directory looking for .md files
	err := filepath.Walk(source.LocalPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only process .md files
		if filepath.Ext(path) != ".md" {
			return nil
		}

		// Skip README and common non-prompt files
		baseName := strings.ToLower(filepath.Base(path))
		if baseName == "readme.md" || baseName == "license.md" || baseName == "contributing.md" {
			return nil
		}

		// Read file
		content, err := os.ReadFile(path)
		if err != nil {
			// Log warning but continue with other files
			fmt.Fprintf(os.Stderr, "Warning: failed to read %s: %v\n", path, err)
			return nil
		}

		// Extract metadata
		name := strings.TrimSuffix(filepath.Base(path), ".md")
		name = slugify(name)

		description := extractDescription(content, 150)
		relPath, _ := filepath.Rel(source.LocalPath, path)

		prompt := models.Prompt{
			ID:          fmt.Sprintf("%s:%s", source.ID, name),
			SourceID:    source.ID,
			Name:        name,
			Content:     string(content),
			Description: description,
			Tags:        []string{},
			Author:      "",
			Version:     "",
			FilePath:    relPath,
			IndexedAt:   time.Now(),
			UpdatedAt:   info.ModTime(),
		}

		prompts = append(prompts, prompt)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	if len(prompts) == 0 {
		return nil, fmt.Errorf("no markdown prompts found in %s", source.LocalPath)
	}

	return prompts, nil
}
