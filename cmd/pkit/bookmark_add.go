package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/whisller/pkit/internal/bookmark"
	"github.com/whisller/pkit/internal/config"
	"github.com/whisller/pkit/internal/index"
	"github.com/whisller/pkit/pkg/models"
)

var bookmarkAddCmd = &cobra.Command{
	Use:   "add <prompt-id>",
	Short: "Add a prompt to bookmarks",
	Long: `Add a prompt to your bookmarks for quick access.

Examples:
  pkit bookmark add fabric:code-review
  pkit bookmark add awesome:linux-terminal --notes "My favorite"`,
	Args: cobra.ExactArgs(1),
	RunE: runBookmarkAdd,
}

var bookmarkAddNotes string

func init() {
	bookmarkCmd.AddCommand(bookmarkAddCmd)

	bookmarkAddCmd.Flags().StringVar(&bookmarkAddNotes, "notes", "", "Optional notes")
}

func runBookmarkAdd(cmd *cobra.Command, args []string) error {
	promptID := args[0]

	// Check if prompt exists in index
	indexBasePath, err := config.GetIndexPath()
	if err != nil {
		return fmt.Errorf("failed to get index path: %w", err)
	}

	indexPath := filepath.Join(indexBasePath, "prompts.bleve")
	indexer, err := index.NewIndexer(indexPath)
	if err != nil {
		return fmt.Errorf("failed to open index: %w", err)
	}
	defer indexer.Close()

	prompt, err := indexer.GetPromptByID(promptID)
	if err != nil {
		return fmt.Errorf("prompt not found: %w", err)
	}

	// Create bookmark
	bm := models.Bookmark{
		PromptID: promptID,
		Notes:    bookmarkAddNotes,
	}

	// Validate bookmark
	if err := bookmark.ValidateBookmark(&bm); err != nil {
		return fmt.Errorf("invalid bookmark: %w", err)
	}

	// Add bookmark using manager
	manager := bookmark.NewManager()
	if err := manager.AddBookmark(bm); err != nil {
		return fmt.Errorf("failed to add bookmark: %w", err)
	}

	fmt.Fprintf(os.Stdout, "Bookmarked prompt '%s'\n", prompt.Name)

	return nil
}
