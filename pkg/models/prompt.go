package models

import (
	"fmt"
	"time"
)

// Prompt represents a single prompt discovered from a source.
type Prompt struct {
	// Unique identifier: <source_id>:<prompt_name> (e.g., "fabric:summarize")
	ID string `json:"id" validate:"required,prompt_id"`

	// Source identifier this prompt belongs to
	SourceID string `json:"source_id" validate:"required"`

	// Prompt name (unique within source)
	Name string `json:"name" validate:"required,prompt_name"`

	// Full prompt text content
	Content string `json:"content" validate:"required"`

	// Brief description for search results and list views (~150 chars)
	// Extracted by parser or derived from first paragraph
	Description string `json:"description"`

	// Tags/categories extracted by parser or derived
	Tags []string `json:"tags,omitempty" validate:"omitempty,dive,tag"`

	// Author information if available from source
	// May be empty if source format doesn't provide it
	Author string `json:"author,omitempty"`

	// Version information if provided by source
	// May be empty if source format doesn't provide it
	Version string `json:"version,omitempty"`

	// Relative file path within source repository
	FilePath string `json:"file_path" validate:"required"`

	// Format-specific metadata (JSON blob)
	// Store extra fields that don't fit standard schema
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// When this prompt was first indexed
	IndexedAt time.Time `json:"indexed_at"`

	// When this prompt was last updated in source
	UpdatedAt time.Time `json:"updated_at"`
}

// Validate checks if the Prompt has valid field values.
// Returns an error if any validation rule is violated.
func (p *Prompt) Validate() error {
	// Use validator for struct validation
	if err := validate.Struct(p); err != nil {
		return err
	}

	// Additional cross-field validation: ID must match SourceID:Name
	expectedID := fmt.Sprintf("%s:%s", p.SourceID, p.Name)
	if p.ID != expectedID {
		return fmt.Errorf("prompt ID %q does not match source_id:name %q", p.ID, expectedID)
	}

	return nil
}
