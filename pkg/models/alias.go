package models

import (
	"fmt"
	"time"
)

// Alias represents a user-defined shortcut name for a prompt ID.
type Alias struct {
	// User-defined alias name (unique identifier)
	Name string `yaml:"name" json:"name"`

	// Reference to prompt: <source_id>:<prompt_name>
	PromptID string `yaml:"prompt_id" json:"prompt_id"`

	// Creation timestamp (RFC3339)
	CreatedAt time.Time `yaml:"created_at" json:"created_at"`

	// Last modified timestamp (RFC3339)
	UpdatedAt time.Time `yaml:"updated_at" json:"updated_at"`
}

// Validate checks if the Alias has valid field values.
func (a *Alias) Validate() error {
	if a.Name == "" {
		return fmt.Errorf("alias name cannot be empty")
	}

	if a.PromptID == "" {
		return fmt.Errorf("prompt_id cannot be empty")
	}

	return nil
}
