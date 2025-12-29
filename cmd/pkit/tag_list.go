package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/whisller/pkit/internal/tag"
)

var tagListCmd = &cobra.Command{
	Use:   "list [prompt-id]",
	Short: "List tags for a prompt or all tagged prompts",
	Long: `List tags for a specific prompt, or list all tagged prompts if no prompt-id specified.

Examples:
  pkit tag list                        # List all tagged prompts
  pkit tag list fabric:code-review     # List tags for specific prompt`,
	Args: cobra.MaximumNArgs(1),
	RunE: runTagList,
}

func init() {
	tagCmd.AddCommand(tagListCmd)
}

func runTagList(cmd *cobra.Command, args []string) error {
	manager := tag.NewManager()

	// If prompt ID provided, show tags for that prompt
	if len(args) > 0 {
		promptID := args[0]
		tags, err := manager.GetTags(promptID)
		if err != nil {
			return fmt.Errorf("failed to get tags: %w", err)
		}

		if len(tags) == 0 {
			// Output to stdout - error extremely rare (stdout closed/redirected)
			_, _ = fmt.Fprintf(os.Stdout, "No tags found for prompt '%s'\n", promptID)
			return nil
		}

		// Output to stdout - error extremely rare (stdout closed/redirected)
		_, _ = fmt.Fprintf(os.Stdout, "Tags for '%s': %s\n", promptID, strings.Join(tags, ", "))
		return nil
	}

	// Otherwise, list all tagged prompts
	allTags, err := manager.ListAllTags()
	if err != nil {
		return fmt.Errorf("failed to list tags: %w", err)
	}

	if len(allTags) == 0 {
		// Output to stdout - error extremely rare (stdout closed/redirected)
		_, _ = fmt.Fprintln(os.Stdout, "No tags found. Use 'pkit tag add' to tag prompts.")
		return nil
	}

	// Display in a table
	table := tablewriter.NewWriter(os.Stdout)

	table.Options(
		tablewriter.WithRowAutoWrap(1),
	)

	table.Header("PROMPT ID", "TAGS")

	for _, pt := range allTags {
		tagsStr := strings.Join(pt.Tags, ", ")
		// table.Append is in-memory operation, error extremely rare
		_ = table.Append(
			pt.PromptID,
			tagsStr,
		)
	}

	if err := table.Render(); err != nil {
		return fmt.Errorf("failed to render table: %w", err)
	}

	return nil
}
