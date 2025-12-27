package models

import (
	"fmt"
	"time"
)

// Config represents user configuration including sources, GitHub token reference, and preferences.
type Config struct {
	// Config file version (for future migrations post-stabilization)
	Version string `yaml:"version" json:"version" validate:"required"`

	// List of subscribed sources
	Sources []Source `yaml:"sources" json:"sources" validate:"dive"`

	// GitHub configuration
	GitHub GitHubConfig `yaml:"github" json:"github" validate:"required"`

	// Search preferences
	Search SearchConfig `yaml:"search" json:"search" validate:"required"`

	// Display preferences
	Display DisplayConfig `yaml:"display" json:"display" validate:"required"`

	// Cache settings
	Cache CacheConfig `yaml:"cache" json:"cache" validate:"required"`
}

// GitHubConfig contains GitHub API configuration
type GitHubConfig struct {
	// Whether to use authenticated requests (token from keyring)
	UseAuth bool `yaml:"use_auth" json:"use_auth"`

	// Rate limit warning threshold (percentage, 0-100)
	RateLimitWarningThreshold int `yaml:"rate_limit_warning_threshold" json:"rate_limit_warning_threshold" validate:"gte=50,lte=95"`

	// Last known rate limit state (ephemeral, not critical to persist)
	LastRateLimit *RateLimit `yaml:"last_rate_limit,omitempty" json:"last_rate_limit,omitempty"`
}

// SearchConfig contains search preferences
type SearchConfig struct {
	// Maximum search results to display
	MaxResults int `yaml:"max_results" json:"max_results" validate:"gte=10,lte=1000"`

	// Fuzzy matching enabled
	FuzzyMatch bool `yaml:"fuzzy_match" json:"fuzzy_match"`

	// Case sensitive search
	CaseSensitive bool `yaml:"case_sensitive" json:"case_sensitive"`
}

// DisplayConfig contains display preferences
type DisplayConfig struct {
	// Use color in output
	Color bool `yaml:"color" json:"color"`

	// Table style (simple, rounded, unicode)
	TableStyle string `yaml:"table_style" json:"table_style" validate:"required,table_style"`

	// Date format (rfc3339, relative, short)
	DateFormat string `yaml:"date_format" json:"date_format" validate:"required,date_format"`
}

// CacheConfig contains cache settings
type CacheConfig struct {
	// Enable search index caching
	Enabled bool `yaml:"enabled" json:"enabled"`

	// Auto-rebuild index on source changes
	AutoRebuild bool `yaml:"auto_rebuild" json:"auto_rebuild"`
}

// RateLimit tracks GitHub API rate limit consumption
type RateLimit struct {
	// Maximum requests allowed in window
	Limit int `json:"limit"`

	// Remaining requests in current window
	Remaining int `json:"remaining"`

	// When the rate limit window resets (Unix timestamp)
	ResetAt time.Time `json:"reset_at"`

	// Resource type (core, search, graphql)
	Resource string `json:"resource"`

	// Whether authenticated request was used
	Authenticated bool `json:"authenticated"`

	// Last updated timestamp
	UpdatedAt time.Time `json:"updated_at"`
}

// PercentageUsed calculates the percentage of rate limit consumed (0-100)
func (r *RateLimit) PercentageUsed() int {
	if r.Limit == 0 {
		return 0
	}
	return int((float64(r.Limit-r.Remaining) / float64(r.Limit)) * 100)
}

// ShouldWarn checks if the warning threshold has been exceeded
func (r *RateLimit) ShouldWarn(threshold int) bool {
	return r.PercentageUsed() >= threshold
}

// TimeUntilReset returns the duration until rate limit resets
func (r *RateLimit) TimeUntilReset() time.Duration {
	return time.Until(r.ResetAt)
}

// DefaultConfig returns a Config with default values
func DefaultConfig() *Config {
	return &Config{
		Version: "1.0",
		Sources: []Source{},
		GitHub: GitHubConfig{
			UseAuth:                   false,
			RateLimitWarningThreshold: 80,
		},
		Search: SearchConfig{
			MaxResults:    50,
			FuzzyMatch:    true,
			CaseSensitive: false,
		},
		Display: DisplayConfig{
			Color:      true,
			TableStyle: "rounded",
			DateFormat: "relative",
		},
		Cache: CacheConfig{
			Enabled:     true,
			AutoRebuild: true,
		},
	}
}

// Validate checks if the Config has valid field values.
// Returns an error if any validation rule is violated.
func (c *Config) Validate() error {
	// Use validator for struct validation
	if err := validate.Struct(c); err != nil {
		return err
	}

	// Additional validation: Check for duplicate source IDs
	sourceIDs := make(map[string]bool)
	for i, src := range c.Sources {
		if sourceIDs[src.ID] {
			return fmt.Errorf("duplicate source ID %q at index %d", src.ID, i)
		}
		sourceIDs[src.ID] = true
	}

	return nil
}
