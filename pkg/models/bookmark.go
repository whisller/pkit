package models

import (
	"fmt"
	"time"
)

// Bookmark represents a user's saved reference to a prompt for quick access.
type Bookmark struct {
	// Reference to prompt: <source_id>:<prompt_name>
	PromptID string `yaml:"prompt_id" json:"prompt_id" validate:"required,prompt_id"`

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

// Validate checks if the Bookmark has valid field values.
// Returns an error if any validation rule is violated.
func (b *Bookmark) Validate() error {
	if b.PromptID == "" {
		return fmt.Errorf("prompt_id cannot be empty")
	}

	return nil
}
