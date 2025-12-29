package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/whisller/pkit/internal/config"
)

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove GitHub authentication",
	Long: `Remove your stored GitHub authentication token.

This command will delete the token from your system's keyring.
Note: If you set GITHUB_TOKEN environment variable, you'll need to
unset it manually.

Examples:
  pkit auth logout       # Remove stored token
  pkit auth logout -f    # Force removal without confirmation`,
	RunE: runAuthLogout,
}

var logoutForce bool

func init() {
	authCmd.AddCommand(authLogoutCmd)
	authLogoutCmd.Flags().BoolVarP(&logoutForce, "force", "f", false, "Skip confirmation prompt")
}

func runAuthLogout(cmd *cobra.Command, args []string) error {
	// Check if token exists
	hasToken := config.HasGitHubToken()
	if !hasToken {
		fmt.Fprintln(os.Stdout, "No authentication token found")
		return nil
	}

	// Get token for display (masked)
	token, err := config.GetGitHubToken()
	if err != nil {
		return fmt.Errorf("failed to retrieve token: %w", err)
	}

	maskedToken := maskToken(token)

	// Confirm logout unless force flag is set
	if !logoutForce {
		fmt.Fprintf(os.Stderr, "This will remove your stored GitHub token (%s)\n", maskedToken)
		confirmed, err := promptForConfirmation("Are you sure you want to logout?")
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}

		if !confirmed {
			fmt.Fprintln(os.Stdout, "Logout cancelled")
			return nil
		}
	}

	// Delete token from keyring
	if err := config.DeleteGitHubToken(); err != nil {
		return fmt.Errorf("failed to delete token: %w", err)
	}

	fmt.Fprintln(os.Stdout, "✓ Logged out successfully")
	fmt.Fprintln(os.Stdout, "\nNote: If you set GITHUB_TOKEN environment variable,")
	fmt.Fprintln(os.Stdout, "you'll need to unset it manually:")
	fmt.Fprintln(os.Stdout, "  unset GITHUB_TOKEN")

	return nil
}
