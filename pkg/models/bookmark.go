package models

import "time"

// Bookmark represents a user's bookmarked prompt with custom alias.
// Full implementation will be added in Phase 2 (Foundation).
type Bookmark struct {
	Alias      string     `yaml:"alias" json:"alias"`
	PromptID   string     `yaml:"prompt_id" json:"prompt_id"`
	SourceID   string     `yaml:"source_id" json:"source_id"`
	PromptName string     `yaml:"prompt_name" json:"prompt_name"`
	Tags       []string   `yaml:"tags" json:"tags"`
	Notes      string     `yaml:"notes" json:"notes"`
	CreatedAt  time.Time  `yaml:"created_at" json:"created_at"`
	UpdatedAt  time.Time  `yaml:"updated_at" json:"updated_at"`
	UsageCount int        `yaml:"usage_count" json:"usage_count"`
	LastUsedAt *time.Time `yaml:"last_used_at,omitempty" json:"last_used_at,omitempty"`
}
