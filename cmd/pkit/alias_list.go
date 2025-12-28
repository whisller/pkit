package main

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/whisller/pkit/internal/alias"
)

var aliasListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all aliases",
	Long: `List all aliases.

Examples:
  pkit alias list`,
	Args: cobra.NoArgs,
	RunE: runAliasList,
}

func init() {
	aliasCmd.AddCommand(aliasListCmd)
}

func runAliasList(cmd *cobra.Command, args []string) error {
	// Load aliases
	manager := alias.NewManager()
	aliases, err := manager.ListAliases()
	if err != nil {
		return fmt.Errorf("failed to load aliases: %w", err)
	}

	if len(aliases) == 0 {
		fmt.Fprintln(os.Stdout, "No aliases created yet. Use 'pkit alias add' to create aliases.")
		return nil
	}

	// Display aliases in a table
	table := tablewriter.NewWriter(os.Stdout)

	table.Options(
		tablewriter.WithRowAutoWrap(1),
	)

	table.Header("ALIAS", "PROMPT ID")

	for _, a := range aliases {
		table.Append(
			a.Name,
			a.PromptID,
		)
	}

	if err := table.Render(); err != nil {
		return fmt.Errorf("failed to render table: %w", err)
	}

	return nil
}
