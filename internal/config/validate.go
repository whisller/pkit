package config

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/whisller/pkit/pkg/models"
)

// ValidateConfigFile validates the configuration file at the given path.
// Returns an error if the file is corrupted or contains invalid data.
// Returns nil if the file doesn't exist (will use default config).
func ValidateConfigFile(path string) error {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil // Empty state is OK - will use default config
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("cannot read config file: %w", err)
	}

	// Parse YAML
	var cfg models.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("corrupted config file: %w", err)
	}

	// Validate config using model validation
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	return nil
}

// ValidateConfig validates a Config struct.
// This is a convenience wrapper around models.Config.Validate().
func ValidateConfig(cfg *models.Config) error {
	if cfg == nil {
		return fmt.Errorf("config cannot be nil")
	}

	return cfg.Validate()
}

// ValidateAndFixConfig validates the config and attempts to fix common issues.
// Returns the fixed config and any errors that couldn't be fixed.
func ValidateAndFixConfig(cfg *models.Config) (*models.Config, error) {
	if cfg == nil {
		return models.DefaultConfig(), nil
	}

	// Check version
	if cfg.Version == "" {
		cfg.Version = "1.0"
	}

	// Fix rate limit warning threshold if out of range
	if cfg.GitHub.RateLimitWarningThreshold < 50 || cfg.GitHub.RateLimitWarningThreshold > 95 {
		cfg.GitHub.RateLimitWarningThreshold = 80 // Reset to default
	}

	// Fix max results if out of range
	if cfg.Search.MaxResults < 10 || cfg.Search.MaxResults > 1000 {
		cfg.Search.MaxResults = 50 // Reset to default
	}

	// Fix table style if invalid
	validTableStyles := map[string]bool{"simple": true, "rounded": true, "unicode": true}
	if !validTableStyles[cfg.Display.TableStyle] {
		cfg.Display.TableStyle = "rounded" // Reset to default
	}

	// Fix date format if invalid
	validDateFormats := map[string]bool{"rfc3339": true, "relative": true, "short": true}
	if !validDateFormats[cfg.Display.DateFormat] {
		cfg.Display.DateFormat = "relative" // Reset to default
	}

	// Validate sources - cannot auto-fix these
	sourceIDs := make(map[string]bool)
	for i, src := range cfg.Sources {
		if sourceIDs[src.ID] {
			return nil, fmt.Errorf("duplicate source ID %q at index %d", src.ID, i)
		}
		sourceIDs[src.ID] = true

		if err := src.Validate(); err != nil {
			return nil, fmt.Errorf("invalid source %q: %w", src.ID, err)
		}
	}

	return cfg, nil
}

// CheckConfigIntegrity performs a comprehensive integrity check on the config file.
// Returns a list of warnings (non-fatal issues) and any fatal errors.
func CheckConfigIntegrity(path string) (warnings []string, err error) {
	warnings = []string{}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return warnings, nil // Empty state is OK
	}

	// Read and parse
	data, err := os.ReadFile(path)
	if err != nil {
		return warnings, fmt.Errorf("cannot read config file: %w", err)
	}

	var cfg models.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return warnings, fmt.Errorf("corrupted config file: %w", err)
	}

	// Check for deprecated fields or unusual values
	if cfg.Version != "1.0" {
		warnings = append(warnings, fmt.Sprintf("unexpected config version: %s (expected 1.0)", cfg.Version))
	}

	// Check if sources have local paths that don't exist
	for _, src := range cfg.Sources {
		if src.LocalPath != "" {
			if _, err := os.Stat(src.LocalPath); os.IsNotExist(err) {
				warnings = append(warnings, fmt.Sprintf("source %q local path does not exist: %s", src.ID, src.LocalPath))
			}
		}
	}

	// Validate the config
	if err := cfg.Validate(); err != nil {
		return warnings, err
	}

	return warnings, nil
}
