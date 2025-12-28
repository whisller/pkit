package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/whisller/pkit/internal/tag"
)

var tagAddCmd = &cobra.Command{
	Use:   "add <prompt-id> <tags>",
	Short: "Add tags to a prompt",
	Long: `Add tags to a prompt. Tags will be merged with existing tags.

Examples:
  pkit tag add fabric:code-review dev,security
  pkit tag add fabric:code-review "dev, security"       # Spaces are OK
  pkit tag add awesome:linux-terminal "linux, dev, shell"`,
	Args: cobra.MinimumNArgs(2),
	RunE: runTagAdd,
}

func init() {
	tagCmd.AddCommand(tagAddCmd)
}

func runTagAdd(cmd *cobra.Command, args []string) error {
	promptID := args[0]
	// Join all remaining args in case shell split them
	tagString := strings.Join(args[1:], " ")

	// Parse tags
	tags := tag.ParseTags(tagString)

	if len(tags) == 0 {
		return fmt.Errorf("no valid tags provided")
	}

	// Add tags using manager
	manager := tag.NewManager()
	if err := manager.AddTags(promptID, tags); err != nil {
		return fmt.Errorf("failed to add tags: %w", err)
	}

	fmt.Fprintf(os.Stdout, "Added tags to prompt '%s': %s\n", promptID, strings.Join(tags, ", "))

	return nil
}
