package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/whisller/pkit/internal/tag"
)

var tagRemoveCmd = &cobra.Command{
	Use:   "remove <prompt-id> [tags]",
	Short: "Remove tags from a prompt",
	Long: `Remove specific tags from a prompt, or all tags if no tags specified.

Examples:
  pkit tag remove fabric:code-review dev          # Remove 'dev' tag
  pkit tag remove fabric:code-review "dev, security"  # Remove multiple tags
  pkit tag remove fabric:code-review              # Remove all tags`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runTagRemove,
}

func init() {
	tagCmd.AddCommand(tagRemoveCmd)
}

func runTagRemove(cmd *cobra.Command, args []string) error {
	promptID := args[0]
	tagString := ""
	if len(args) > 1 {
		tagString = args[1]
	}

	// Parse tags (empty means remove all)
	tags := tag.ParseTags(tagString)

	// Remove tags using manager
	manager := tag.NewManager()
	if err := manager.RemoveTags(promptID, tags); err != nil {
		return fmt.Errorf("failed to remove tags: %w", err)
	}

	if len(tags) == 0 {
		fmt.Fprintf(os.Stdout, "Removed all tags from prompt '%s'\n", promptID)
	} else {
		fmt.Fprintf(os.Stdout, "Removed tags from prompt '%s': %s\n", promptID, strings.Join(tags, ", "))
	}

	return nil
}
