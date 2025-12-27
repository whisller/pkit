package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/whisller/pkit/pkg/models"
)

// FabricParser parses Fabric patterns from danielmiessler/fabric repository.
// Fabric patterns are stored in patterns/*/system.md files.
type FabricParser struct{}

// NewFabricParser creates a new Fabric parser instance.
func NewFabricParser() *FabricParser {
	return &FabricParser{}
}

// Name returns the parser name.
func (p *FabricParser) Name() string {
	return "fabric_pattern"
}

// CanParse checks if the source path contains Fabric patterns.
func (p *FabricParser) CanParse(sourcePath string) bool {
	// Check for data/patterns directory (new Fabric structure)
	patternsDir := filepath.Join(sourcePath, "data", "patterns")
	if info, err := os.Stat(patternsDir); err == nil && info.IsDir() {
		return true
	}
	// Fallback to patterns directory (old structure)
	patternsDir = filepath.Join(sourcePath, "patterns")
	info, err := os.Stat(patternsDir)
	return err == nil && info.IsDir()
}

// ParsePrompts extracts all prompts from Fabric patterns directory.
func (p *FabricParser) ParsePrompts(source *models.Source) ([]models.Prompt, error) {
	var prompts []models.Prompt

	// Walk source directory looking for system.md files in patterns/*/
	// Try data/patterns first (new structure), then patterns (old structure)
	patternsDir := filepath.Join(source.LocalPath, "data", "patterns")
	if _, err := os.Stat(patternsDir); os.IsNotExist(err) {
		patternsDir = filepath.Join(source.LocalPath, "patterns")
	}

	// Check if patterns directory exists
	if _, err := os.Stat(patternsDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("patterns directory not found: %w", err)
	}

	entries, err := os.ReadDir(patternsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read patterns directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		systemFile := filepath.Join(patternsDir, entry.Name(), "system.md")
		if _, err := os.Stat(systemFile); os.IsNotExist(err) {
			continue
		}

		// Read file content
		content, err := os.ReadFile(systemFile)
		if err != nil {
			// Log warning but continue with other patterns
			fmt.Fprintf(os.Stderr, "Warning: failed to read %s: %v\n", systemFile, err)
			continue
		}

		// Extract metadata from content
		name := entry.Name()
		description := extractDescription(content, 150)

		// Get file mod time for UpdatedAt
		fileInfo, _ := os.Stat(systemFile)
		updatedAt := time.Now()
		if fileInfo != nil {
			updatedAt = fileInfo.ModTime()
		}

		prompt := models.Prompt{
			ID:          fmt.Sprintf("%s:%s", source.ID, name),
			SourceID:    source.ID,
			Name:        name,
			Content:     string(content),
			Description: description,
			Tags:        []string{},
			Author:      "",
			Version:     "",
			FilePath:    fmt.Sprintf("patterns/%s/system.md", name),
			IndexedAt:   time.Now(),
			UpdatedAt:   updatedAt,
		}

		prompts = append(prompts, prompt)
	}

	if len(prompts) == 0 {
		return nil, fmt.Errorf("no patterns found in %s", patternsDir)
	}

	return prompts, nil
}

// extractDescription extracts the first paragraph from content, truncating to maxLen.
// It skips headers and returns the first content paragraph after a header.
func extractDescription(content []byte, maxLen int) string {
	text := string(content)
	lines := strings.Split(text, "\n")

	var paragraph []string
	inContent := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip headers
		if strings.HasPrefix(line, "#") {
			inContent = true
			continue
		}

		if inContent {
			if line == "" && len(paragraph) > 0 {
				break // End of first paragraph
			}
			if line != "" {
				paragraph = append(paragraph, line)
			}
		}
	}

	desc := strings.Join(paragraph, " ")

	// Truncate if too long
	if len(desc) > maxLen {
		return desc[:maxLen-3] + "..."
	}

	if desc == "" {
		return "Fabric pattern"
	}

	return desc
}
