package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/whisller/pkit/internal/bookmark"
	"github.com/whisller/pkit/internal/display"
)

var showCmd = &cobra.Command{
	Use:   "show <alias|prompt-id>",
	Short: "Show detailed information about a prompt",
	Long: `Show detailed information about a prompt including metadata and full content.

The show command displays:
- Prompt ID and source
- Name and description
- Tags
- Author information
- Full prompt content

Examples:
  pkit show review                    # Show by alias
  pkit show fabric:code-review        # Show by prompt ID`,
	Args: cobra.ExactArgs(1),
	RunE: runShow,
}

var (
	showJSON    bool
	showVerbose bool
)

func init() {
	rootCmd.AddCommand(showCmd)

	showCmd.Flags().BoolVar(&showJSON, "json", false, "Output as JSON")
	showCmd.Flags().BoolVarP(&showVerbose, "verbose", "v", false, "Show additional metadata")
}

func runShow(cmd *cobra.Command, args []string) error {
	identifier := args[0]

	if showVerbose {
		fmt.Fprintf(os.Stderr, "→ Resolving: %s\n", identifier)
	}

	// Resolve identifier to prompt
	prompt, err := bookmark.ResolveWithContext(identifier)
	if err != nil {
		return fmt.Errorf("failed to resolve '%s': %w", identifier, err)
	}

	if showVerbose {
		fmt.Fprintf(os.Stderr, "→ Found: %s (%s)\n", prompt.Name, prompt.ID)
	}

	// Output based on format
	if showJSON {
		if err := display.PrintPromptJSON(os.Stdout, prompt); err != nil {
			return fmt.Errorf("failed to output JSON: %w", err)
		}
	} else {
		if err := display.PrintPromptWithMetadata(os.Stdout, prompt); err != nil {
			return fmt.Errorf("failed to display prompt: %w", err)
		}
	}

	return nil
}
