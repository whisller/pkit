package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/whisller/pkit/internal/alias"
)

var aliasRemoveCmd = &cobra.Command{
	Use:   "remove <alias-name>",
	Short: "Remove an alias",
	Long: `Remove an alias.

Examples:
  pkit alias remove review
  pkit alias remove term`,
	Args: cobra.ExactArgs(1),
	RunE: runAliasRemove,
}

func init() {
	aliasCmd.AddCommand(aliasRemoveCmd)
}

func runAliasRemove(cmd *cobra.Command, args []string) error {
	aliasName := args[0]

	// Remove alias using manager
	manager := alias.NewManager()
	err := manager.RemoveAlias(aliasName)

	if err != nil {
		return fmt.Errorf("failed to remove alias: %w", err)
	}

	fmt.Fprintf(os.Stdout, "Removed alias '%s'\n", aliasName)

	return nil
}
