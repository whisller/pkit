package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/whisller/pkit/internal/config"
	"github.com/whisller/pkit/internal/source"
)

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check authentication status",
	Long: `Check your GitHub authentication status and token validity.

This command will:
  - Check if a token is stored
  - Validate the token with GitHub API
  - Show rate limit information
  - Display token permissions

Examples:
  pkit auth status           # Check current authentication
  pkit auth status --verbose # Show detailed token information`,
	RunE: runAuthStatus,
}

var authStatusVerbose bool

func init() {
	authCmd.AddCommand(authStatusCmd)
	authStatusCmd.Flags().BoolVarP(&authStatusVerbose, "verbose", "v", false, "Show detailed information")
}

func runAuthStatus(cmd *cobra.Command, args []string) error {
	// Check if token exists
	hasToken := config.HasGitHubToken()
	if !hasToken {
		fmt.Fprintln(os.Stdout, "✗ Not authenticated")
		fmt.Fprintln(os.Stdout, "\nTo authenticate:")
		fmt.Fprintln(os.Stdout, "  pkit auth login")
		fmt.Fprintln(os.Stdout, "\nOr set environment variable:")
		fmt.Fprintln(os.Stdout, "  export GITHUB_TOKEN=ghp_your_token_here")
		return nil
	}

	// Get token
	token, err := config.GetGitHubToken()
	if err != nil {
		return fmt.Errorf("failed to retrieve token: %w", err)
	}

	if token == "" {
		fmt.Fprintln(os.Stdout, "✗ Not authenticated (token is empty)")
		return nil
	}

	// Validate token with GitHub API
	client := source.NewGitHubClient(token)
	rateLimit, err := client.CheckRateLimit()
	if err != nil {
		fmt.Fprintln(os.Stdout, "✗ Authentication failed")
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintln(os.Stderr, "\nYour token may be invalid or expired. Please login again:")
		fmt.Fprintln(os.Stderr, "  pkit auth login")
		return nil
	}

	// Show authentication status
	fmt.Fprintln(os.Stdout, "✓ Authenticated with GitHub")
	fmt.Fprintf(os.Stdout, "  Token: %s\n", maskToken(token))
	fmt.Fprintf(os.Stdout, "  Status: %s\n", getAuthenticationStatus(rateLimit.Authenticated))

	// Show rate limit information
	fmt.Fprintln(os.Stdout, "\nRate Limit:")
	fmt.Fprintf(os.Stdout, "  Limit: %d requests/hour\n", rateLimit.Limit)
	fmt.Fprintf(os.Stdout, "  Remaining: %d\n", rateLimit.Remaining)
	fmt.Fprintf(os.Stdout, "  Resets at: %s\n", rateLimit.ResetAt.Format("2006-01-02 15:04:05 MST"))

	// Calculate usage percentage
	usagePercent := 0
	if rateLimit.Limit > 0 {
		usagePercent = ((rateLimit.Limit - rateLimit.Remaining) * 100) / rateLimit.Limit
	}
	fmt.Fprintf(os.Stdout, "  Usage: %d%%\n", usagePercent)

	// Show warnings if rate limit is low
	if rateLimit.Remaining < 100 {
		fmt.Fprintln(os.Stderr, "\n⚠ Warning: Low rate limit remaining")
		if !rateLimit.Authenticated {
			fmt.Fprintln(os.Stderr, "  Consider authenticating to get higher rate limits")
		}
	}

	// Verbose output
	if authStatusVerbose {
		fmt.Fprintln(os.Stdout, "\nToken Storage:")
		fmt.Fprintf(os.Stdout, "  Method: System keyring (with GITHUB_TOKEN fallback)\n")
		fmt.Fprintf(os.Stdout, "  Service: pkit\n")
		fmt.Fprintf(os.Stdout, "  Username: github\n")

		fmt.Fprintln(os.Stdout, "\nCapabilities:")
		fmt.Fprintln(os.Stdout, "  ✓ Access public repositories")
		fmt.Fprintln(os.Stdout, "  ✓ Higher rate limits (5000 requests/hour)")
		fmt.Fprintln(os.Stdout, "\nPrivate Repository Access:")
		fmt.Fprintln(os.Stdout, "  Requires one of:")
		fmt.Fprintln(os.Stdout, "    - Fine-grained token: Contents (Read-only) + Metadata (Read-only)")
		fmt.Fprintln(os.Stdout, "    - Classic token: 'repo' scope")
		fmt.Fprintln(os.Stdout, "\n  To verify private repo access, try:")
		fmt.Fprintln(os.Stdout, "    pkit subscribe owner/private-repo")
	}

	// Show next steps
	fmt.Fprintln(os.Stdout, "\nYou can now subscribe to private repositories:")
	fmt.Fprintln(os.Stdout, "  pkit subscribe owner/private-repo")

	return nil
}

func getAuthenticationStatus(authenticated bool) string {
	if authenticated {
		return "Authenticated (token valid)"
	}
	return "Unauthenticated (token may be invalid)"
}
