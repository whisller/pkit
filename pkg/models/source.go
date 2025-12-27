package models

import (
	"time"
)

// Source represents a subscribed GitHub repository containing prompts.
type Source struct {
	// Unique identifier (e.g., "fabric", "awesome-chatgpt-prompts")
	ID string `yaml:"id" json:"id" validate:"required,source_id"`

	// Display name for the source
	Name string `yaml:"name" json:"name" validate:"required"`

	// Full GitHub URL (e.g., "https://github.com/danielmiessler/fabric")
	URL string `yaml:"url" json:"url" validate:"required,url,github_url"`

	// Short form used in subscribe command (e.g., "fabric/patterns")
	ShortName string `yaml:"short_name,omitempty" json:"short_name,omitempty"`

	// Local filesystem path (~/.pkit/sources/<id>)
	LocalPath string `yaml:"local_path" json:"local_path" validate:"required"`

	// Format type determines which parser to use
	// Valid values: "fabric_pattern", "awesome_chatgpt", "markdown"
	Format string `yaml:"format" json:"format" validate:"required,oneof=fabric_pattern awesome_chatgpt markdown"`

	// Current git commit SHA
	CommitSHA string `yaml:"commit_sha" json:"commit_sha" validate:"omitempty,git_sha"`

	// Last update check timestamp (RFC3339)
	LastChecked time.Time `yaml:"last_checked" json:"last_checked"`

	// Last successful index timestamp (RFC3339)
	LastIndexed time.Time `yaml:"last_indexed" json:"last_indexed"`

	// Number of prompts indexed from this source
	PromptCount int `yaml:"prompt_count" json:"prompt_count" validate:"gte=0"`

	// Subscription timestamp (RFC3339)
	SubscribedAt time.Time `yaml:"subscribed_at" json:"subscribed_at"`

	// Whether updates are available upstream
	UpdateAvailable bool `yaml:"update_available" json:"update_available"`

	// Upstream commit SHA if update available
	UpstreamSHA string `yaml:"upstream_sha,omitempty" json:"upstream_sha,omitempty" validate:"omitempty,git_sha"`
}

// Validate checks if the Source has valid field values.
// Returns an error if any validation rule is violated.
func (s *Source) Validate() error {
	return validate.Struct(s)
}
