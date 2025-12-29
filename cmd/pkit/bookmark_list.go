package main

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/whisller/pkit/internal/bookmark"
)

var bookmarkListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all saved bookmarks",
	Long: `List all saved bookmarks.

Examples:
  pkit bookmark list`,
	Args: cobra.NoArgs,
	RunE: runBookmarkList,
}

func init() {
	bookmarkCmd.AddCommand(bookmarkListCmd)
}

func runBookmarkList(cmd *cobra.Command, args []string) error {
	// Load bookmarks
	manager := bookmark.NewManager()
	bookmarks, err := manager.ListBookmarks()
	if err != nil {
		return fmt.Errorf("failed to load bookmarks: %w", err)
	}

	if len(bookmarks) == 0 {
		_, _ = fmt.Fprintln(os.Stdout, "No bookmarks saved yet. Use 'pkit bookmark add' to create bookmarks.")
		return nil
	}

	// Display bookmarks in a table
	table := tablewriter.NewWriter(os.Stdout)

	table.Options(
		tablewriter.WithRowAutoWrap(1),
	)

	table.Header("PROMPT ID", "USAGE")

	for _, bm := range bookmarks {
		usageStr := fmt.Sprintf("%d", bm.UsageCount)

		_ = table.Append(
			bm.PromptID,
			usageStr,
		)
	}

	if err := table.Render(); err != nil {
		return fmt.Errorf("failed to render table: %w", err)
	}

	return nil
}
