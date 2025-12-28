package tag

import (
	"fmt"
	"time"

	"github.com/whisller/pkit/pkg/models"
)

// Manager handles operations on prompt tags.
type Manager struct{}

// NewManager creates a new tag manager.
func NewManager() *Manager {
	return &Manager{}
}

// AddTags adds tags to a prompt (merges with existing tags).
func (m *Manager) AddTags(promptID string, newTags []string) error {
	// Load existing tags
	allTags, err := LoadTags()
	if err != nil {
		return fmt.Errorf("failed to load tags: %w", err)
	}

	// Find existing tags for this prompt
	found := false
	for i := range allTags {
		if allTags[i].PromptID == promptID {
			// Merge tags (deduplicate)
			tagMap := make(map[string]bool)
			for _, t := range allTags[i].Tags {
				tagMap[t] = true
			}
			for _, t := range newTags {
				tagMap[t] = true
			}

			// Convert back to slice
			mergedTags := make([]string, 0, len(tagMap))
			for t := range tagMap {
				mergedTags = append(mergedTags, t)
			}

			allTags[i].Tags = mergedTags
			allTags[i].UpdatedAt = time.Now()
			found = true
			break
		}
	}

	// If not found, create new entry
	if !found {
		now := time.Now()
		allTags = append(allTags, models.PromptTags{
			PromptID:  promptID,
			Tags:      newTags,
			CreatedAt: now,
			UpdatedAt: now,
		})
	}

	// Save
	if err := SaveTags(allTags); err != nil {
		return fmt.Errorf("failed to save tags: %w", err)
	}

	return nil
}

// RemoveTags removes specific tags from a prompt, or all tags if tags slice is empty.
func (m *Manager) RemoveTags(promptID string, tagsToRemove []string) error {
	// Load existing tags
	allTags, err := LoadTags()
	if err != nil {
		return fmt.Errorf("failed to load tags: %w", err)
	}

	// Find and update or remove entry
	found := false
	newAllTags := make([]models.PromptTags, 0, len(allTags))
	for _, pt := range allTags {
		if pt.PromptID == promptID {
			found = true

			// If no specific tags to remove, remove all (don't add to newAllTags)
			if len(tagsToRemove) == 0 {
				continue
			}

			// Remove specific tags
			removeMap := make(map[string]bool)
			for _, t := range tagsToRemove {
				removeMap[t] = true
			}

			remainingTags := make([]string, 0)
			for _, t := range pt.Tags {
				if !removeMap[t] {
					remainingTags = append(remainingTags, t)
				}
			}

			// Only keep entry if there are remaining tags
			if len(remainingTags) > 0 {
				pt.Tags = remainingTags
				pt.UpdatedAt = time.Now()
				newAllTags = append(newAllTags, pt)
			}
		} else {
			newAllTags = append(newAllTags, pt)
		}
	}

	if !found {
		return fmt.Errorf("no tags found for prompt '%s'", promptID)
	}

	// Save
	if err := SaveTags(newAllTags); err != nil {
		return fmt.Errorf("failed to save tags: %w", err)
	}

	return nil
}

// GetTags retrieves tags for a specific prompt.
func (m *Manager) GetTags(promptID string) ([]string, error) {
	// Load existing tags
	allTags, err := LoadTags()
	if err != nil {
		return nil, fmt.Errorf("failed to load tags: %w", err)
	}

	// Find tags for this prompt
	for _, pt := range allTags {
		if pt.PromptID == promptID {
			return pt.Tags, nil
		}
	}

	// No tags found (not an error, just empty)
	return []string{}, nil
}

// ListAllTags returns all prompt tags.
func (m *Manager) ListAllTags() ([]models.PromptTags, error) {
	allTags, err := LoadTags()
	if err != nil {
		return nil, fmt.Errorf("failed to load tags: %w", err)
	}

	return allTags, nil
}
