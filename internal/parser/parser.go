package parser

import (
	"github.com/whisller/pkit/pkg/models"
)

// Package parser handles format-specific prompt parsing (Fabric, awesome-chatgpt-prompts, markdown).

// Parser defines the interface for parsing prompts from different source formats.
// Each source format (Fabric, awesome-chatgpt-prompts, markdown) implements this interface.
type Parser interface {
	// ParsePrompts extracts all prompts from a source directory.
	// Returns a slice of Prompt models and any error encountered.
	ParsePrompts(source *models.Source) ([]models.Prompt, error)

	// CanParse checks if this parser can handle the given source path.
	// Used for auto-detection of source formats.
	CanParse(sourcePath string) bool

	// Name returns the parser name for logging and debugging.
	Name() string
}
