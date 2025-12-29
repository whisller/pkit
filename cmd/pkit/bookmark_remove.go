package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/whisller/pkit/internal/bookmark"
)

var bookmarkRemoveCmd = &cobra.Command{
	Use:   "remove <prompt-id>",
	Short: "Remove a prompt from bookmarks",
	Long: `Remove a prompt from your bookmarks.

Examples:
  pkit bookmark remove fabric:code-review
  pkit bookmark remove awesome:linux-terminal`,
	Args: cobra.ExactArgs(1),
	RunE: runBookmarkRemove,
}

func init() {
	bookmarkCmd.AddCommand(bookmarkRemoveCmd)
}

func runBookmarkRemove(cmd *cobra.Command, args []string) error {
	promptID := args[0]

	// Remove bookmark using manager
	manager := bookmark.NewManager()
	err := manager.RemoveBookmark(promptID)

	if err != nil {
		return fmt.Errorf("failed to remove bookmark: %w", err)
	}

	// Output to stdout - error extremely rare (stdout closed/redirected)
	_, _ = fmt.Fprintf(os.Stdout, "Removed bookmark for prompt '%s'\n", promptID)

	return nil
}
