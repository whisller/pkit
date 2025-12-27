package models

import "time"

// Source represents a subscribed GitHub repository.
// Full implementation will be added in Phase 2 (Foundation).
type Source struct {
	ID          string    `yaml:"id" json:"id"`
	Name        string    `yaml:"name" json:"name"`
	URL         string    `yaml:"url" json:"url"`
	Format      string    `yaml:"format" json:"format"`
	LocalPath   string    `yaml:"local_path" json:"local_path"`
	LastIndexed time.Time `yaml:"last_indexed" json:"last_indexed"`
	CommitSHA   string    `yaml:"commit_sha" json:"commit_sha"`
	PromptCount int       `yaml:"prompt_count" json:"prompt_count"`
}
