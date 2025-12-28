package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/whisller/pkit/internal/config"
	"github.com/whisller/pkit/internal/index"
	"github.com/whisller/pkit/internal/source"
	"github.com/whisller/pkit/pkg/models"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade [source]",
	Short: "Upgrade source(s) to the latest commit and re-index",
	Long: `Upgrade one or more sources by pulling latest changes from the remote repository
and re-indexing the prompts.

Examples:
  pkit upgrade fabric              # Upgrade specific source
  pkit upgrade --all               # Upgrade all sources with updates
  pkit upgrade fabric --force      # Force upgrade even if no updates`,
	Args: cobra.MaximumNArgs(1),
	RunE: runUpgrade,
}

var (
	upgradeAll     bool
	upgradeForce   bool
	upgradeVerbose bool
)

func init() {
	rootCmd.AddCommand(upgradeCmd)

	upgradeCmd.Flags().BoolVar(&upgradeAll, "all", false, "Upgrade all sources with updates")
	upgradeCmd.Flags().BoolVar(&upgradeForce, "force", false, "Force upgrade even if no updates")
	upgradeCmd.Flags().BoolVarP(&upgradeVerbose, "verbose", "v", false, "Show detailed progress")
}

func runUpgrade(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(cfg.Sources) == 0 {
		return fmt.Errorf("no sources subscribed. Use 'pkit subscribe' first")
	}

	// Get GitHub token
	token, err := config.GetGitHubToken()
	if err != nil && upgradeVerbose {
		fmt.Fprintf(os.Stderr, "→ Warning: failed to get GitHub token: %v\n", err)
		fmt.Fprintln(os.Stderr, "→ Proceeding without authentication")
	}

	// Create source manager
	mgr := source.NewManager(token)

	// Get index path
	indexBasePath, err := config.GetIndexPath()
	if err != nil {
		return fmt.Errorf("failed to get index path: %w", err)
	}

	indexPath := filepath.Join(indexBasePath, "prompts.bleve")

	// Open index
	indexer, err := index.NewIndexer(indexPath)
	if err != nil {
		return fmt.Errorf("failed to open index: %w", err)
	}
	defer indexer.Close()

	// Determine which sources to upgrade
	var sourcesToUpgrade []models.Source

	if upgradeAll {
		// Check all sources for updates
		if upgradeVerbose {
			fmt.Fprintln(os.Stderr, "→ Checking for updates...")
		}

		for _, src := range cfg.Sources {
			if upgradeForce {
				sourcesToUpgrade = append(sourcesToUpgrade, src)
				continue
			}

			hasUpdates, _, err := mgr.CheckForUpdates(&src)
			if err != nil {
				fmt.Fprintf(os.Stderr, "→ Warning: failed to check updates for %s: %v\n", src.ID, err)
				continue
			}

			if hasUpdates {
				sourcesToUpgrade = append(sourcesToUpgrade, src)
			}
		}

		if len(sourcesToUpgrade) == 0 {
			fmt.Fprintln(os.Stderr, "✓ All sources are up to date")
			return nil
		}
	} else {
		// Upgrade specific source
		if len(args) == 0 {
			return fmt.Errorf("source name required or use --all flag")
		}

		sourceID := args[0]
		found := false
		for _, src := range cfg.Sources {
			if src.ID == sourceID {
				sourcesToUpgrade = []models.Source{src}
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("source not found: %s", sourceID)
		}

		// Check for updates if not forcing
		if !upgradeForce {
			hasUpdates, _, err := mgr.CheckForUpdates(&sourcesToUpgrade[0])
			if err != nil {
				return fmt.Errorf("failed to check updates: %w", err)
			}

			if !hasUpdates {
				fmt.Fprintf(os.Stderr, "✓ Source '%s' is already up to date\n", sourceID)
				fmt.Fprintln(os.Stderr, "Use --force to upgrade anyway")
				return nil
			}
		}
	}

	// Upgrade sources (parallel if multiple)
	if len(sourcesToUpgrade) > 1 {
		return upgradeMultipleSources(mgr, indexer, cfg, sourcesToUpgrade)
	}

	// Upgrade single source
	return upgradeSingleSource(mgr, indexer, cfg, &sourcesToUpgrade[0])
}

func upgradeSingleSource(mgr *source.Manager, indexer *index.Indexer, cfg *models.Config, src *models.Source) error {
	if upgradeVerbose {
		fmt.Fprintf(os.Stderr, "→ Upgrading %s...\n", src.ID)
	}

	// Update repository
	newSHA, err := mgr.Update(src)
	if err != nil {
		return fmt.Errorf("failed to update %s: %w", src.ID, err)
	}

	if upgradeVerbose {
		fmt.Fprintf(os.Stderr, "→ Updated to commit %s\n", newSHA[:8])
	}

	// Re-index prompts
	if err := reindexSourcePrompts(indexer, src); err != nil {
		return fmt.Errorf("failed to re-index %s: %w", src.ID, err)
	}

	// Update config with new commit SHA
	for i := range cfg.Sources {
		if cfg.Sources[i].ID == src.ID {
			cfg.Sources[i].CommitSHA = newSHA
			break
		}
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Fprintf(os.Stderr, "✓ Upgraded %s\n", src.ID)
	return nil
}

func upgradeMultipleSources(mgr *source.Manager, indexer *index.Indexer, cfg *models.Config, sources []models.Source) error {
	fmt.Fprintf(os.Stderr, "→ Upgrading %d source(s) in parallel...\n", len(sources))

	var mu sync.Mutex
	updatedSHAs := make(map[string]string)

	g := new(errgroup.Group)

	for _, src := range sources {
		src := src // Capture loop variable
		g.Go(func() error {
			if upgradeVerbose {
				fmt.Fprintf(os.Stderr, "→ Upgrading %s...\n", src.ID)
			}

			// Update repository
			newSHA, err := mgr.Update(&src)
			if err != nil {
				return fmt.Errorf("failed to update %s: %w", src.ID, err)
			}

			// Re-index prompts
			if err := reindexSourcePrompts(indexer, &src); err != nil {
				return fmt.Errorf("failed to re-index %s: %w", src.ID, err)
			}

			mu.Lock()
			updatedSHAs[src.ID] = newSHA
			mu.Unlock()

			if upgradeVerbose {
				fmt.Fprintf(os.Stderr, "✓ Upgraded %s to %s\n", src.ID, newSHA[:8])
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	// Update config with new commit SHAs
	for i := range cfg.Sources {
		if newSHA, ok := updatedSHAs[cfg.Sources[i].ID]; ok {
			cfg.Sources[i].CommitSHA = newSHA
		}
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Fprintf(os.Stderr, "✓ Upgraded %d source(s)\n", len(sources))
	return nil
}

func reindexSourcePrompts(indexer *index.Indexer, src *models.Source) error {
	// Get parser for this source
	p, err := source.GetParser(src.Format)
	if err != nil {
		return fmt.Errorf("failed to get parser: %w", err)
	}

	// Parse prompts
	prompts, err := p.ParsePrompts(src)
	if err != nil {
		return fmt.Errorf("failed to parse prompts: %w", err)
	}

	// Delete old prompts for this source from index
	if err := indexer.DeleteBySource(src.ID); err != nil {
		return fmt.Errorf("failed to delete old prompts: %w", err)
	}

	// Index new prompts
	if err := indexer.IndexPrompts(prompts); err != nil {
		return fmt.Errorf("failed to index prompts: %w", err)
	}

	if upgradeVerbose {
		fmt.Fprintf(os.Stderr, "  ✓ Re-indexed %d prompts\n", len(prompts))
	}

	return nil
}
