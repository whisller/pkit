package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/whisller/pkit/internal/config"
	"github.com/whisller/pkit/internal/source"
)

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with GitHub",
	Long: `Authenticate with GitHub using a Personal Access Token (PAT).

This command will prompt you to enter your GitHub token, which will be
stored securely in your system's keyring.

To create a GitHub Personal Access Token:

  Option 1: Fine-grained token (Recommended - More secure)
    1. Go to: https://github.com/settings/personal-access-tokens/new
    2. Token name: "pkit CLI"
    3. Expiration: 90 days recommended
    4. Repository access: "All repositories" or "Only select repositories"
    5. Permissions → Repository permissions:
       - Contents: Read-only (required)
       - Metadata: Read-only (auto-selected)

  Option 2: Classic token
    1. Go to: https://github.com/settings/tokens
    2. Click "Generate new token (classic)"
    3. Token name: "pkit CLI"
    4. Select scopes:
       - "repo" - Full control of private repositories (for private repos)
       - "public_repo" - Access public repositories (for public repos only)
       - No scopes - Public repos only with rate limit benefits

Examples:
  pkit auth login                    # Interactive prompt
  pkit auth login --token ghp_xxx    # Non-interactive`,
	RunE: runAuthLogin,
}

var loginToken string

func init() {
	authCmd.AddCommand(authLoginCmd)
	authLoginCmd.Flags().StringVar(&loginToken, "token", "", "GitHub Personal Access Token")
}

func runAuthLogin(cmd *cobra.Command, args []string) error {
	var token string
	var err error

	// If token not provided via flag, prompt for it
	if loginToken == "" {
		token, err = promptForToken()
		if err != nil {
			return fmt.Errorf("failed to read token: %w", err)
		}
	} else {
		token = loginToken
	}

	// Validate token format
	token = strings.TrimSpace(token)
	if token == "" {
		return fmt.Errorf("token cannot be empty")
	}

	// Validate token with GitHub API
	fmt.Fprintln(os.Stderr, "Validating token with GitHub API...")
	if err := validateGitHubToken(token); err != nil {
		return fmt.Errorf("token validation failed: %w", err)
	}

	// Store token in keyring
	if err := config.SetGitHubToken(token); err != nil {
		return fmt.Errorf("failed to store token: %w", err)
	}

	// Success message with masked token
	maskedToken := maskToken(token)
	fmt.Fprintf(os.Stdout, "✓ Authenticated successfully as %s\n", maskedToken)
	fmt.Fprintln(os.Stdout, "\nYou can now subscribe to private repositories:")
	fmt.Fprintln(os.Stdout, "  pkit subscribe owner/private-repo")

	return nil
}

// promptForToken prompts the user to enter their GitHub token
func promptForToken() (string, error) {
	fmt.Fprint(os.Stderr, "Enter your GitHub Personal Access Token: ")

	// Read token without echoing to terminal
	tokenBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Fprintln(os.Stderr) // New line after password input
	if err != nil {
		return "", err
	}

	return string(tokenBytes), nil
}

// validateGitHubToken validates the token by making an API call
func validateGitHubToken(token string) error {
	client := source.NewGitHubClient(token)

	// Try to get rate limit info (this endpoint works with any valid token)
	rateLimit, err := client.CheckRateLimit()
	if err != nil {
		return fmt.Errorf("invalid token or network error: %w", err)
	}

	// Check if token is authenticated
	if !rateLimit.Authenticated {
		return fmt.Errorf("token appears to be invalid (unauthenticated)")
	}

	// Show rate limit info
	fmt.Fprintf(os.Stderr, "✓ Token validated (rate limit: %d/%d requests remaining)\n",
		rateLimit.Remaining, rateLimit.Limit)

	return nil
}

// maskToken masks the token for display (shows first 7 chars and last 4)
func maskToken(token string) string {
	if len(token) < 12 {
		return "***"
	}
	return token[:7] + "..." + token[len(token)-4:]
}

// promptForConfirmation asks user for yes/no confirmation
func promptForConfirmation(message string) (bool, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Fprintf(os.Stderr, "%s [y/N]: ", message)

	response, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes", nil
}
