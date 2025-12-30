package source

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/whisller/pkit/internal/parser"
	"github.com/whisller/pkit/pkg/models"
)

// Package source handles source repository management (git operations, format detection).

// Manager handles source subscription, updates, and format detection.
type Manager struct {
	token string
}

// NewManager creates a new source manager.
func NewManager(token string) *Manager {
	return &Manager{
		token: token,
	}
}

// DetectSourceFormat auto-detects the format of a source repository.
// Checks for Fabric patterns, awesome-chatgpt CSV, or defaults to markdown.
func DetectSourceFormat(localPath string) string {
	// Check for Fabric patterns structure (data/patterns)
	if _, err := os.Stat(filepath.Join(localPath, "data", "patterns")); err == nil {
		return "fabric_pattern"
	}

	// Check for awesome-chatgpt-prompts CSV
	if _, err := os.Stat(filepath.Join(localPath, "prompts.csv")); err == nil {
		return "awesome_chatgpt"
	}

	// Default to generic markdown
	return "markdown"
}

// GetParser returns the appropriate parser for a source format.
func GetParser(format string) (parser.Parser, error) {
	switch format {
	case "fabric_pattern":
		return parser.NewFabricParser(), nil
	case "awesome_chatgpt":
		return parser.NewAwesomeChatGPTParser(), nil
	case "markdown":
		return parser.NewMarkdownParser(), nil
	default:
		return nil, fmt.Errorf("unknown format: %s", format)
	}
}

// Subscribe subscribes to a new source repository.
// Clones the repository, detects format, and returns the Source model.
func (m *Manager) Subscribe(url, localPath string) (*models.Source, error) {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create parent directory: %w", err)
	}

	// Clone repository
	commitSHA, err := CloneRepository(url, localPath, m.token)
	if err != nil {
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	// Detect format
	format := DetectSourceFormat(localPath)

	// Extract source ID from URL
	sourceID := ExtractSourceIDFromURL(url)

	// Create Source model
	source := &models.Source{
		ID:        sourceID,
		Name:      sourceID, // Can be customized later
		URL:       url,
		LocalPath: localPath,
		Format:    format,
		CommitSHA: commitSHA,
	}

	return source, nil
}

// SubscribeMultiple subscribes to multiple sources in parallel using errgroup.
func (m *Manager) SubscribeMultiple(sources []struct {
	URL       string
	LocalPath string
}) (map[string]*models.Source, error) {
	var mu sync.Mutex
	results := make(map[string]*models.Source)

	g := new(errgroup.Group)

	for _, src := range sources {
		src := src // Capture loop variable
		g.Go(func() error {
			source, err := m.Subscribe(src.URL, src.LocalPath)
			if err != nil {
				return fmt.Errorf("failed to subscribe to %s: %w", src.URL, err)
			}

			mu.Lock()
			results[source.ID] = source
			mu.Unlock()

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return results, nil
}

// CheckForUpdates checks if a source has updates available.
// Returns true if updates are available, along with the remote commit SHA.
func (m *Manager) CheckForUpdates(source *models.Source) (hasUpdates bool, remoteSHA string, err error) {
	return CheckForUpdates(source.LocalPath, m.token)
}

// Update updates an existing source repository.
// Pulls latest changes and returns the new commit SHA.
func (m *Manager) Update(source *models.Source) (string, error) {
	// Pull latest changes
	commitSHA, err := PullRepository(source.LocalPath, m.token)
	if err != nil {
		return "", fmt.Errorf("failed to update repository: %w", err)
	}

	return commitSHA, nil
}

// ExtractSourceIDFromURL extracts a source ID from a GitHub URL.
// Example: "https://github.com/danielmiessler/fabric" -> "danielmiessler/fabric"
func ExtractSourceIDFromURL(url string) string {
	// Remove protocol
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "github.com/")

	// Remove .git suffix
	url = strings.TrimSuffix(url, ".git")

	// Split by / and take owner/repo (last 2 parts)
	parts := strings.Split(url, "/")
	if len(parts) >= 2 {
		// Return owner/repo format
		return parts[len(parts)-2] + "/" + parts[len(parts)-1]
	} else if len(parts) == 1 {
		// Fallback to just repo name if only one part
		return parts[0]
	}

	return "unknown"
}
