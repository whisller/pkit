package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/whisller/pkit/internal/alias"
	"github.com/whisller/pkit/internal/config"
	"github.com/whisller/pkit/internal/index"
	"github.com/whisller/pkit/pkg/models"
)

var aliasAddCmd = &cobra.Command{
	Use:   "add <prompt-id> <alias-name>",
	Short: "Create an alias for a prompt",
	Long: `Create a short alias name for a prompt ID.

Examples:
  pkit alias add fabric:code-review review
  pkit alias add awesome:linux-terminal term
  pkit alias add fabric:dialog_with_socrates socrates`,
	Args: cobra.ExactArgs(2),
	RunE: runAliasAdd,
}

func init() {
	aliasCmd.AddCommand(aliasAddCmd)
}

func runAliasAdd(cmd *cobra.Command, args []string) error {
	promptID := args[0]
	aliasName := args[1]

	// Validate alias name format
	if err := alias.ValidateAliasName(aliasName); err != nil {
		return fmt.Errorf("invalid alias name: %w", err)
	}

	// Check alias name uniqueness
	if err := alias.ValidateAliasUnique(aliasName, ""); err != nil {
		return err
	}

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

	_, err = indexer.GetPromptByID(promptID)
	if err != nil {
		return fmt.Errorf("prompt not found: %w", err)
	}

	// Create alias
	a := models.Alias{
		Name:     aliasName,
		PromptID: promptID,
	}

	// Add alias using manager
	manager := alias.NewManager()
	if err := manager.AddAlias(a); err != nil {
		return fmt.Errorf("failed to create alias: %w", err)
	}

	fmt.Fprintf(os.Stdout, "Created alias '%s' for prompt '%s'\n", aliasName, promptID)

	return nil
}
