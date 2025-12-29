package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
	"github.com/whisller/pkit/pkg/models"
)

// Package config handles configuration loading, saving, and validation.

const (
	// DefaultConfigDir is the default directory for pkit configuration
	DefaultConfigDir = ".pkit"

	// ConfigFileName is the name of the configuration file
	ConfigFileName = "config.yml"
)

// GetConfigPath returns the full path to the configuration file.
// It uses ~/.pkit/config.yml by default.
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, DefaultConfigDir, ConfigFileName)
	return configPath, nil
}

// GetSourcesPath returns the full path to the sources directory.
// It uses ~/.pkit/sources by default.
func GetSourcesPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	sourcesPath := filepath.Join(homeDir, DefaultConfigDir, "sources")
	return sourcesPath, nil
}

// GetIndexPath returns the full path to the index directory.
// It uses ~/.pkit/index by default.
func GetIndexPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	indexPath := filepath.Join(homeDir, DefaultConfigDir, "index")
	return indexPath, nil
}

// EnsureConfigDir ensures the configuration directory exists.
// Creates ~/.pkit/ if it doesn't exist.
func EnsureConfigDir() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, DefaultConfigDir)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	return nil
}

// Load reads and parses the configuration file from ~/.pkit/config.yml.
// Returns a default configuration if the file doesn't exist.
// Returns an error if the file exists but cannot be read or parsed.
func Load() (*models.Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// If config file doesn't exist, return default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return models.DefaultConfig(), nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var cfg models.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate config
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// Save writes the configuration to ~/.pkit/config.yml using atomic write.
// Uses a temporary file and rename to ensure atomicity.
func Save(cfg *models.Config) error {
	if cfg == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Validate config before saving
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	// Ensure config directory exists
	if err := EnsureConfigDir(); err != nil {
		return err
	}

	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Marshal to YAML with proper indentation
	data, err := yaml.MarshalWithOptions(cfg, yaml.Indent(2))
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to temporary file first (atomic write pattern)
	tmpFile := configPath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write temporary config file: %w", err)
	}

	// Rename temporary file to actual config file (atomic operation)
	if err := os.Rename(tmpFile, configPath); err != nil {
		// Clean up temporary file on error (best effort)
		_ = os.Remove(tmpFile)
		return fmt.Errorf("failed to save config file: %w", err)
	}

	return nil
}
