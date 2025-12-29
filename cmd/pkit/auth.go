package main

import (
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage GitHub authentication",
	Long: `Manage GitHub authentication for accessing private repositories.

pkit uses GitHub Personal Access Tokens (PATs) to authenticate with GitHub.
Tokens are stored securely in your system's keyring (macOS Keychain,
Linux Secret Service, Windows Credential Manager).

Examples:
  pkit auth login        # Set GitHub token interactively
  pkit auth status       # Check token validity and permissions
  pkit auth logout       # Remove stored token`,
}

func init() {
	rootCmd.AddCommand(authCmd)
}
