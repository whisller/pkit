package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/whisller/pkit/internal/bookmark"
	"github.com/whisller/pkit/internal/config"
	"github.com/whisller/pkit/internal/display"
	"github.com/whisller/pkit/internal/index"
	"github.com/whisller/pkit/internal/tui"
	"github.com/whisller/pkit/pkg/models"
)

var findCmd = &cobra.Command{
	Use:   "find [query]",
	Short: "Interactively find and select a prompt",
	Long: `Find launches an interactive TUI for browsing and selecting prompts.

The find command provides:
- Real-time fuzzy search filtering
- Interactive prompt selection with arrow keys
- Keyboard shortcuts for common actions:
  - Enter: Select prompt (outputs ID)
  - Ctrl+G: Get prompt content
  - Ctrl+S: Bookmark prompt
  - Ctrl+T: Add tags to prompt
  - Q/Esc: Quit

When not running in a TTY (e.g., piped), falls back to search command behavior.

Examples:
  pkit find                         # Launch interactive finder
  pkit find code                    # Pre-filter by "code"
  pkit find --get | claude          # Interactive select + auto-get`,
	Args: cobra.MaximumNArgs(1),
	RunE: runFind,
}

var (
	findGet     bool
	findVerbose bool
)

func init() {
	rootCmd.AddCommand(findCmd)

	findCmd.Flags().BoolVarP(&findGet, "get", "g", false, "Automatically get the selected prompt content")
	findCmd.Flags().BoolVarP(&findVerbose, "verbose", "v", false, "Show detailed progress")
}

func runFind(cmd *cobra.Command, args []string) (err error) {
	// Check if stdout is a TTY
	isTTY := isatty.IsTerminal(os.Stdout.Fd())

	if !isTTY {
		// Fall back to search-like behavior when piped
		if findVerbose {
			fmt.Fprintln(os.Stderr, "→ Not in TTY, falling back to search mode")
		}
		return runFindFallback(args)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if any sources are subscribed
	if len(cfg.Sources) == 0 {
		return fmt.Errorf(`no sources subscribed

Subscribe to sources first:
  pkit subscribe fabric/patterns
  pkit subscribe f/awesome-chatgpt-prompts

Then use find:
  pkit find`)
	}

	// Get index path
	indexBasePath, err := config.GetIndexPath()
	if err != nil {
		return fmt.Errorf("failed to get index path: %w", err)
	}

	// Open index
	indexPath := filepath.Join(indexBasePath, "prompts.bleve")
	indexer, err := index.NewIndexer(indexPath)
	if err != nil {
		return fmt.Errorf("failed to open index: %w", err)
	}
	defer func() {
		if closeErr := indexer.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close index: %w", closeErr)
		}
	}()

	// Search for prompts
	query := ""
	if len(args) > 0 {
		query = args[0]
	}

	searchOpts := index.SearchOptions{
		Query:      query,
		MaxResults: 1000, // Get all prompts for interactive filtering
	}

	results, err := indexer.Search(searchOpts)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if len(results) == 0 {
		fmt.Fprintln(os.Stderr, "No prompts found")
		return nil
	}

	// Extract prompts for TUI
	prompts := make([]models.Prompt, len(results))
	for i, result := range results {
		prompts[i] = result.Prompt
	}

	// Run interactive finder
	selectedID, action, err := tui.RunFinder(prompts)
	if err != nil {
		return fmt.Errorf("finder error: %w", err)
	}

	if selectedID == "" {
		// User quit without selection
		return nil
	}

	// Handle the selected action
	switch action {
	case "get":
		return handleFindGet(selectedID)
	case "bookmark":
		return handleFindBookmark(selectedID)
	case "tag":
		return handleFindTag(selectedID)
	case "select":
		// If --get flag is set, get the prompt
		if findGet {
			return handleFindGet(selectedID)
		}
		// Otherwise just output the ID
		_, _ = fmt.Fprintln(os.Stdout, selectedID)
		return nil
	default:
		_, _ = fmt.Fprintln(os.Stdout, selectedID)
		return nil
	}
}

func handleFindGet(promptID string) error {
	// Resolve and output prompt content
	prompt, err := bookmark.ResolveWithContext(promptID)
	if err != nil {
		return fmt.Errorf("failed to resolve prompt: %w", err)
	}

	return display.PrintPromptText(os.Stdout, prompt)
}

func handleFindBookmark(promptID string) error {
	// Add bookmark
	mgr := bookmark.NewManager()

	bm := models.Bookmark{
		PromptID: promptID,
		Notes:    "Added via interactive finder",
	}

	if err := mgr.AddBookmark(bm); err != nil {
		return fmt.Errorf("failed to bookmark: %w", err)
	}

	fmt.Fprintf(os.Stderr, "✓ Bookmarked: %s\n", promptID)
	return nil
}

func handleFindTag(promptID string) error {
	// This is a simplified version - in a full implementation,
	// you'd want to prompt the user for tags
	fmt.Fprintf(os.Stderr, "To add tags, use: pkit tag add %s <tags>\n", promptID)
	return nil
}

func runFindFallback(args []string) (err error) {
	// Simple fallback: search and output prompt IDs
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(cfg.Sources) == 0 {
		return fmt.Errorf("no sources subscribed")
	}

	indexBasePath, err := config.GetIndexPath()
	if err != nil {
		return fmt.Errorf("failed to get index path: %w", err)
	}

	indexPath := filepath.Join(indexBasePath, "prompts.bleve")
	indexer, err := index.NewIndexer(indexPath)
	if err != nil {
		return fmt.Errorf("failed to open index: %w", err)
	}
	defer func() {
		if closeErr := indexer.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close index: %w", closeErr)
		}
	}()

	query := ""
	if len(args) > 0 {
		query = args[0]
	}

	searchOpts := index.SearchOptions{
		Query:      query,
		MaxResults: 50,
	}

	results, err := indexer.Search(searchOpts)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	// Output prompt IDs (one per line for piping) - error extremely rare
	for _, result := range results {
		_, _ = fmt.Fprintln(os.Stdout, result.Prompt.ID)
	}

	return nil
}
