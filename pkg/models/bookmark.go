package models

import (
	"fmt"
	"time"
)

// Bookmark represents a user's saved reference to a prompt with custom alias and tags.
type Bookmark struct {
	// User-defined alias (unique identifier)
	Alias string `yaml:"alias" json:"alias" validate:"required,alias,not_reserved"`

	// Reference to prompt: <source_id>:<prompt_name>
	PromptID string `yaml:"prompt_id" json:"prompt_id" validate:"required,prompt_id"`

	// Source identifier (denormalized for faster lookups)
	SourceID string `yaml:"source_id" json:"source_id" validate:"required"`

	// Prompt name (denormalized for faster lookups)
	PromptName string `yaml:"prompt_name" json:"prompt_name" validate:"required"`

	// User-defined tags (comma-separated in commands, array in storage)
	Tags []string `yaml:"tags,omitempty" json:"tags,omitempty" validate:"omitempty,dive,tag"`

	// Optional user notes
	Notes string `yaml:"notes,omitempty" json:"notes,omitempty"`

	// Creation timestamp (RFC3339)
	CreatedAt time.Time `yaml:"created_at" json:"created_at"`

	// Last modified timestamp (RFC3339)
	UpdatedAt time.Time `yaml:"updated_at" json:"updated_at"`

	// Usage count (incremented on each `pkit get`)
	UsageCount int `yaml:"usage_count" json:"usage_count" validate:"gte=0"`

	// Last used timestamp (RFC3339)
	LastUsedAt *time.Time `yaml:"last_used_at,omitempty" json:"last_used_at,omitempty"`
}

// reservedCommands defines built-in command names that cannot be used as aliases
var reservedCommands = map[string]bool{
	"subscribe": true,
	"search":    true,
	"find":      true,
	"list":      true,
	"show":      true,
	"save":      true,
	"get":       true,
	"alias":     true,
	"tag":       true,
	"rm":        true,
	"update":    true,
	"upgrade":   true,
	"status":    true,
	"help":      true,
	"version":   true,
	"init":      true,
	"config":    true,
}

// Validate checks if the Bookmark has valid field values.
// Returns an error if any validation rule is violated.
func (b *Bookmark) Validate() error {
	// Use validator for struct validation
	if err := validate.Struct(b); err != nil {
		return err
	}

	// Additional cross-field validation: PromptID must match SourceID:PromptName
	expectedID := fmt.Sprintf("%s:%s", b.SourceID, b.PromptName)
	if b.PromptID != expectedID {
		return fmt.Errorf("prompt_id %q does not match source_id:prompt_name %q", b.PromptID, expectedID)
	}

	return nil
}
