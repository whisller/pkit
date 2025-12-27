package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/whisller/pkit/internal/bookmark"
	"github.com/whisller/pkit/internal/display"
)

var getCmd = &cobra.Command{
	Use:   "get <alias|prompt-id>",
	Short: "Get prompt content for piping to execution tools",
	Long: `Get prompt content for piping to execution tools like claude, llm, fabric, and mods.

The get command resolves a bookmark alias or prompt ID and outputs ONLY the prompt
content to stdout - no headers, no metadata, no formatting. This ensures clean piping.

Examples:
  pkit get review                    # Get bookmarked prompt by alias
  pkit get fabric:summarize          # Get prompt by ID
  pkit get review | claude -p "test" # Pipe to claude
  pkit get review --json             # Output as JSON`,
	Args: cobra.ExactArgs(1),
	RunE: runGet,
}

var (
	getJSON    bool
	getVerbose bool
	getDebug   bool
)

func init() {
	rootCmd.AddCommand(getCmd)

	getCmd.Flags().BoolVar(&getJSON, "json", false, "Output prompt metadata as JSON")
	getCmd.Flags().BoolVarP(&getVerbose, "verbose", "v", false, "Show operation details to stderr")
	getCmd.Flags().BoolVar(&getDebug, "debug", false, "Show full trace to stderr")
}

func runGet(cmd *cobra.Command, args []string) error {
	identifier := args[0]

	if getVerbose || getDebug {
		fmt.Fprintf(os.Stderr, "→ Resolving: %s\n", identifier)
	}

	// Resolve identifier to prompt
	prompt, err := bookmark.ResolveWithContext(identifier)
	if err != nil {
		return fmt.Errorf("failed to resolve '%s': %w", identifier, err)
	}

	if getVerbose || getDebug {
		fmt.Fprintf(os.Stderr, "→ Found: %s (%s)\n", prompt.Name, prompt.ID)
	}

	// Output based on format
	if getJSON {
		if err := display.PrintPromptJSON(os.Stdout, prompt); err != nil {
			return fmt.Errorf("failed to output JSON: %w", err)
		}
	} else {
		// CRITICAL: Output ONLY the content for piping
		if err := display.PrintPromptText(os.Stdout, prompt); err != nil {
			return fmt.Errorf("failed to output prompt: %w", err)
		}
	}

	return nil
}
