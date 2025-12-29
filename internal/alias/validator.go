package alias

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/whisller/pkit/pkg/models"
)

const (
	// AliasesFileName is the name of the aliases file
	AliasesFileName = "aliases.yml"
)

// AliasesFile represents the structure of the aliases YAML file
type AliasesFile struct {
	Aliases []models.Alias `yaml:"aliases"`
}

// GetAliasesPath returns the full path to the aliases file.
// It uses ~/.pkit/aliases.yml by default.
func GetAliasesPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	aliasesPath := filepath.Join(homeDir, ".pkit", AliasesFileName)
	return aliasesPath, nil
}

// Reserved command names that cannot be used as aliases
var reservedAliases = map[string]bool{
	"get":        true,
	"search":     true,
	"find":       true,
	"subscribe":  true,
	"bookmark":   true,
	"bookmarks":  true,
	"alias":      true,
	"aliases":    true,
	"tag":        true,
	"unbookmark": true,
	"unalias":    true,
	"reindex":    true,
	"help":       true,
	"version":    true,
	"status":     true,
	"upgrade":    true,
	"show":       true,
}

// ValidateAliasName validates an alias name format and checks for reserved words.
func ValidateAliasName(name string) error {
	if name == "" {
		return fmt.Errorf("alias name cannot be empty")
	}

	// Check length
	if len(name) < 2 {
		return fmt.Errorf("alias name must be at least 2 characters long")
	}

	if len(name) > 50 {
		return fmt.Errorf("alias name must be no more than 50 characters long")
	}

	// Check format: alphanumeric, hyphens, underscores only
	validAlias := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validAlias.MatchString(name) {
		return fmt.Errorf("alias name can only contain letters, numbers, hyphens, and underscores")
	}

	// Check reserved words
	if reservedAliases[strings.ToLower(name)] {
		return fmt.Errorf("alias name '%s' is reserved and cannot be used", name)
	}

	return nil
}

// ValidateAliasUnique checks if an alias name is unique among existing aliases.
func ValidateAliasUnique(name string, excludeName string) error {
	aliases, err := LoadAliases()
	if err != nil {
		return fmt.Errorf("failed to load aliases: %w", err)
	}

	for _, alias := range aliases {
		// Skip the alias being updated
		if excludeName != "" && alias.Name == excludeName {
			continue
		}

		if alias.Name == name {
			return fmt.Errorf("alias '%s' already exists", name)
		}
	}

	return nil
}

// LoadAliases loads aliases from the file.
// Returns an empty slice if the file doesn't exist.
func LoadAliases() ([]models.Alias, error) {
	path, err := GetAliasesPath()
	if err != nil {
		return nil, err
	}

	// If file doesn't exist, return empty slice
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return []models.Alias{}, nil
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read aliases file: %w", err)
	}

	// Parse YAML
	var aliasesFile AliasesFile
	if err := yaml.Unmarshal(data, &aliasesFile); err != nil {
		return nil, fmt.Errorf("failed to parse aliases file: %w", err)
	}

	return aliasesFile.Aliases, nil
}

// SaveAliases saves aliases to the file using atomic write.
func SaveAliases(aliases []models.Alias) error {
	path, err := GetAliasesPath()
	if err != nil {
		return err
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create aliases directory: %w", err)
	}

	// Create aliases file structure
	aliasesFile := AliasesFile{
		Aliases: aliases,
	}

	// Marshal to YAML
	data, err := yaml.MarshalWithOptions(aliasesFile, yaml.Indent(2))
	if err != nil {
		return fmt.Errorf("failed to marshal aliases: %w", err)
	}

	// Write to temporary file first (atomic write pattern)
	tmpFile := path + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write temporary aliases file: %w", err)
	}

	// Rename temporary file to actual aliases file (atomic operation)
	if err := os.Rename(tmpFile, path); err != nil {
		// Clean up temporary file on error (best effort)
		_ = os.Remove(tmpFile)
		return fmt.Errorf("failed to save aliases file: %w", err)
	}

	return nil
}
