package models

import (
	"fmt"
	"time"
)

// PromptTags represents tags assigned to a prompt.
type PromptTags struct {
	// Reference to prompt: <source_id>:<prompt_name>
	PromptID string `yaml:"prompt_id" json:"prompt_id"`

	// User-defined tags
	Tags []string `yaml:"tags" json:"tags"`

	// Creation timestamp (RFC3339)
	CreatedAt time.Time `yaml:"created_at" json:"created_at"`

	// Last modified timestamp (RFC3339)
	UpdatedAt time.Time `yaml:"updated_at" json:"updated_at"`
}

// Validate checks if the PromptTags has valid field values.
func (pt *PromptTags) Validate() error {
	if pt.PromptID == "" {
		return fmt.Errorf("prompt_id cannot be empty")
	}

	return nil
}
