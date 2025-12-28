package main

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/whisller/pkit/internal/config"
	"github.com/whisller/pkit/internal/source"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of subscribed sources and check for updates",
	Long: `Show status of all subscribed sources including:
- Source name and URL
- Current commit SHA
- Update availability
- Prompt count (if indexed)

Examples:
  pkit status                  # Show status of all sources
  pkit status --check-updates  # Fetch remote changes and check for updates`,
	Args: cobra.NoArgs,
	RunE: runStatus,
}

var (
	statusCheckUpdates bool
	statusVerbose      bool
)

func init() {
	rootCmd.AddCommand(statusCmd)

	statusCmd.Flags().BoolVarP(&statusCheckUpdates, "check-updates", "u", false, "Fetch remote changes and check for updates")
	statusCmd.Flags().BoolVarP(&statusVerbose, "verbose", "v", false, "Show detailed information")
}

func runStatus(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if any sources are subscribed
	if len(cfg.Sources) == 0 {
		fmt.Fprintln(os.Stderr, "No sources subscribed yet")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Subscribe to sources first:")
		fmt.Fprintln(os.Stderr, "  pkit subscribe fabric/patterns")
		fmt.Fprintln(os.Stderr, "  pkit subscribe f/awesome-chatgpt-prompts")
		return nil
	}

	// Get GitHub token
	token, err := config.GetGitHubToken()
	if err != nil && statusVerbose {
		fmt.Fprintf(os.Stderr, "→ Warning: failed to get GitHub token: %v\n", err)
		fmt.Fprintln(os.Stderr, "→ Update checks may be rate limited")
	}

	// Create source manager
	mgr := source.NewManager(token)

	// Display table
	table := tablewriter.NewWriter(os.Stdout)
	table.Header("SOURCE ID", "URL", "COMMIT", "STATUS")

	for _, src := range cfg.Sources {
		// Get current commit
		commitSHA := src.CommitSHA
		if len(commitSHA) > 8 {
			commitSHA = commitSHA[:8]
		}

		status := "Up to date"

		// Check for updates if requested
		if statusCheckUpdates {
			if statusVerbose {
				fmt.Fprintf(os.Stderr, "→ Checking updates for %s...\n", src.ID)
			}

			hasUpdates, remoteSHA, err := mgr.CheckForUpdates(&src)
			if err != nil {
				if statusVerbose {
					fmt.Fprintf(os.Stderr, "→ Warning: failed to check updates for %s: %v\n", src.ID, err)
				}
				status = "Unknown"
			} else if hasUpdates {
				status = fmt.Sprintf("Update available (%s)", remoteSHA[:8])
			}
		}

		table.Append(
			src.ID,
			src.URL,
			commitSHA,
			status,
		)
	}

	// Render the table
	if err := table.Render(); err != nil {
		return fmt.Errorf("failed to render table: %w", err)
	}

	// Print summary
	fmt.Fprintf(os.Stderr, "\nTotal sources: %d\n", len(cfg.Sources))

	if statusCheckUpdates {
		fmt.Fprintln(os.Stderr, "\nUse 'pkit upgrade <source>' to update sources")
		fmt.Fprintln(os.Stderr, "Use 'pkit upgrade --all' to update all sources")
	} else {
		fmt.Fprintln(os.Stderr, "\nUse 'pkit status --check-updates' to check for remote updates")
	}

	return nil
}
