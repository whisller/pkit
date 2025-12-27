package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/whisller/pkit/internal/config"
	"github.com/whisller/pkit/internal/index"
	"github.com/whisller/pkit/internal/source"
	"github.com/whisller/pkit/pkg/models"
)

var reindexCmd = &cobra.Command{
	Use:   "reindex",
	Short: "Rebuild the search index for all subscribed sources",
	Long: `Reindex rebuilds the search index by re-parsing all subscribed sources.

This is useful when:
- The index schema has changed (after pkit upgrades)
- The index is corrupted
- Source files have been updated manually

Examples:
  pkit reindex                    # Reindex all sources
  pkit reindex --source fabric    # Reindex only fabric source`,
	RunE: runReindex,
}

var (
	reindexSource  string
	reindexVerbose bool
)

func init() {
	rootCmd.AddCommand(reindexCmd)

	reindexCmd.Flags().StringVar(&reindexSource, "source", "", "Reindex only this source")
	reindexCmd.Flags().BoolVarP(&reindexVerbose, "verbose", "v", false, "Show detailed progress")
}

func runReindex(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(cfg.Sources) == 0 {
		return fmt.Errorf("no sources subscribed. Use 'pkit subscribe' first")
	}

	// Get index path
	indexBasePath, err := config.GetIndexPath()
	if err != nil {
		return fmt.Errorf("failed to get index path: %w", err)
	}

	indexPath := filepath.Join(indexBasePath, "prompts.bleve")

	// Delete old index
	if reindexVerbose {
		fmt.Println("Deleting old index...")
	}
	if err := index.DeleteIndex(indexPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete old index: %w", err)
	}

	// Create new index
	if reindexVerbose {
		fmt.Println("Creating new index...")
	}
	indexer, err := index.NewIndexer(indexPath)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer indexer.Close()

	// Reindex each source
	sourcesToReindex := cfg.Sources
	if reindexSource != "" {
		// Filter to specific source
		found := false
		for _, src := range cfg.Sources {
			if src.ID == reindexSource {
				sourcesToReindex = []models.Source{src}
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("source not found: %s", reindexSource)
		}
	}

	for _, src := range sourcesToReindex {
		if reindexVerbose {
			fmt.Printf("Reindexing %s...\n", src.ID)
		}

		// Get parser for this source
		p, err := source.GetParser(src.Format)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to get parser for %s: %v\n", src.ID, err)
			continue
		}

		// Parse prompts
		prompts, err := p.ParsePrompts(&src)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to parse %s: %v\n", src.ID, err)
			continue
		}

		// Index prompts
		if err := indexer.IndexPrompts(prompts); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to index %s: %v\n", src.ID, err)
			continue
		}

		if reindexVerbose {
			fmt.Printf("  ✓ Indexed %d prompts\n", len(prompts))
		}
	}

	fmt.Printf("✓ Reindexed %d source(s)\n", len(sourcesToReindex))
	return nil
}
