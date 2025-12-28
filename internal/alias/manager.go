package alias

import (
	"fmt"
	"time"

	"github.com/whisller/pkit/pkg/models"
)

// Manager handles CRUD operations on aliases.
type Manager struct{}

// NewManager creates a new alias manager.
func NewManager() *Manager {
	return &Manager{}
}

// AddAlias adds a new alias to the collection.
// Returns error if alias name already exists.
func (m *Manager) AddAlias(alias models.Alias) error {
	// Load existing aliases
	aliases, err := LoadAliases()
	if err != nil {
		return fmt.Errorf("failed to load aliases: %w", err)
	}

	// Check if alias name already exists
	for _, existing := range aliases {
		if existing.Name == alias.Name {
			return fmt.Errorf("alias '%s' already exists", alias.Name)
		}
	}

	// Set timestamps
	now := time.Now()
	alias.CreatedAt = now
	alias.UpdatedAt = now

	// Add alias
	aliases = append(aliases, alias)

	// Save
	if err := SaveAliases(aliases); err != nil {
		return fmt.Errorf("failed to save aliases: %w", err)
	}

	return nil
}

// UpdateAlias updates an existing alias by name.
// Returns error if alias not found.
func (m *Manager) UpdateAlias(name string, updater func(*models.Alias) error) error {
	// Load existing aliases
	aliases, err := LoadAliases()
	if err != nil {
		return fmt.Errorf("failed to load aliases: %w", err)
	}

	// Find and update alias
	found := false
	for i := range aliases {
		if aliases[i].Name == name {
			// Apply update
			if err := updater(&aliases[i]); err != nil {
				return fmt.Errorf("failed to update alias: %w", err)
			}
			aliases[i].UpdatedAt = time.Now()
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("alias '%s' not found", name)
	}

	// Save
	if err := SaveAliases(aliases); err != nil {
		return fmt.Errorf("failed to save aliases: %w", err)
	}

	return nil
}

// RemoveAlias removes an alias by name.
// Returns error if alias not found.
func (m *Manager) RemoveAlias(name string) error {
	// Load existing aliases
	aliases, err := LoadAliases()
	if err != nil {
		return fmt.Errorf("failed to load aliases: %w", err)
	}

	// Find and remove alias
	found := false
	newAliases := make([]models.Alias, 0, len(aliases))
	for _, alias := range aliases {
		if alias.Name == name {
			found = true
			continue
		}
		newAliases = append(newAliases, alias)
	}

	if !found {
		return fmt.Errorf("alias '%s' not found", name)
	}

	// Save
	if err := SaveAliases(newAliases); err != nil {
		return fmt.Errorf("failed to save aliases: %w", err)
	}

	return nil
}

// GetAlias retrieves an alias by name.
// Returns error if alias not found.
func (m *Manager) GetAlias(name string) (*models.Alias, error) {
	// Load existing aliases
	aliases, err := LoadAliases()
	if err != nil {
		return nil, fmt.Errorf("failed to load aliases: %w", err)
	}

	// Find alias
	for _, alias := range aliases {
		if alias.Name == name {
			return &alias, nil
		}
	}

	return nil, fmt.Errorf("alias '%s' not found", name)
}

// ListAliases returns all aliases.
func (m *Manager) ListAliases() ([]models.Alias, error) {
	// Load existing aliases
	aliases, err := LoadAliases()
	if err != nil {
		return nil, fmt.Errorf("failed to load aliases: %w", err)
	}

	return aliases, nil
}
