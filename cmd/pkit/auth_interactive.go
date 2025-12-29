package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"

	"github.com/whisller/pkit/internal/config"
	"github.com/whisller/pkit/internal/source"
)

// handleAuthenticationError handles authentication errors interactively
// Returns a token if user provides one, empty string if user cancels, or error
func handleAuthenticationError(repoURL string) (string, error) {
	fmt.Fprintln(os.Stderr, "\n✗ Authentication required for this repository")
	fmt.Fprintln(os.Stderr, "\nThis repository is private or requires authentication.")

	// Ask if user wants to authenticate
	fmt.Fprint(os.Stderr, "\nWould you like to provide a GitHub token now? [Y/n]: ")
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	response = strings.ToLower(strings.TrimSpace(response))
	if response == "n" || response == "no" {
		fmt.Fprintln(os.Stderr, "\nAuthentication cancelled.")
		fmt.Fprintln(os.Stderr, "You can authenticate later with:")
		fmt.Fprintln(os.Stderr, "  pkit auth login")
		return "", nil
	}

	// Show instructions for creating a token
	fmt.Fprintln(os.Stderr, "\nTo create a GitHub Personal Access Token:")
	fmt.Fprintln(os.Stderr, "  1. Go to: https://github.com/settings/tokens")
	fmt.Fprintln(os.Stderr, "  2. Click 'Generate new token (classic)'")
	fmt.Fprintln(os.Stderr, "  3. Give it a name (e.g., 'pkit CLI')")
	fmt.Fprintln(os.Stderr, "  4. Select scope: 'repo' (for private repositories)")
	fmt.Fprintln(os.Stderr, "  5. Click 'Generate token' and copy it")

	// Prompt for token
	fmt.Fprint(os.Stderr, "\nEnter your GitHub Personal Access Token: ")
	tokenBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Fprintln(os.Stderr) // New line after password input
	if err != nil {
		return "", fmt.Errorf("failed to read token: %w", err)
	}

	token := strings.TrimSpace(string(tokenBytes))
	if token == "" {
		return "", fmt.Errorf("token cannot be empty")
	}

	// Validate token
	fmt.Fprintln(os.Stderr, "Validating token...")
	if err := validateToken(token); err != nil {
		return "", fmt.Errorf("token validation failed: %w", err)
	}

	fmt.Fprintln(os.Stderr, "✓ Token validated successfully")

	// Ask if user wants to save the token
	fmt.Fprint(os.Stderr, "\nSave token for future use? [Y/n]: ")
	saveResponse, err := reader.ReadString('\n')
	if err != nil {
		// Non-fatal error, just don't save
		fmt.Fprintln(os.Stderr, "Note: Token will only be used for this session")
		return token, nil
	}

	saveResponse = strings.ToLower(strings.TrimSpace(saveResponse))
	if saveResponse != "n" && saveResponse != "no" {
		if err := config.SetGitHubToken(token); err != nil {
			// Non-fatal error, just warn user
			fmt.Fprintf(os.Stderr, "Warning: Could not save token: %v\n", err)
			fmt.Fprintln(os.Stderr, "Token will only be used for this session")
		} else {
			fmt.Fprintln(os.Stderr, "✓ Token saved securely in system keyring")
		}
	} else {
		fmt.Fprintln(os.Stderr, "Note: Token will only be used for this session")
	}

	return token, nil
}

// validateToken validates a GitHub token by making an API call
func validateToken(token string) error {
	client := source.NewGitHubClient(token)

	// Check rate limit to validate token
	rateLimit, err := client.CheckRateLimit()
	if err != nil {
		return fmt.Errorf("invalid token or network error: %w", err)
	}

	if !rateLimit.Authenticated {
		return fmt.Errorf("token appears to be invalid (unauthenticated)")
	}

	return nil
}
