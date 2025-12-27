package config

import (
	"fmt"
	"os"

	"github.com/zalando/go-keyring"
)

const (
	// KeyringService is the service name for storing tokens in the system keyring
	KeyringService = "pkit"

	// KeyringUser is the username for the token entry
	KeyringUser = "github"

	// EnvVarToken is the environment variable name for the GitHub token fallback
	EnvVarToken = "GITHUB_TOKEN"
)

// GetGitHubToken retrieves the GitHub personal access token.
// It first tries to read from the system keyring, then falls back to the GITHUB_TOKEN environment variable.
// Returns an empty string if no token is found (not an error - allows unauthenticated requests).
func GetGitHubToken() (string, error) {
	// Try system keyring first
	token, err := keyring.Get(KeyringService, KeyringUser)
	if err == nil && token != "" {
		return token, nil
	}

	// Check if keyring error is just "not found" vs actual error
	if err != nil && err != keyring.ErrNotFound {
		// Log warning but continue to fallback
		// Non-fatal error: keyring might not be available on this system
		fmt.Fprintf(os.Stderr, "Warning: keyring access failed: %v\n", err)
	}

	// Fallback to environment variable
	token = os.Getenv(EnvVarToken)
	if token != "" {
		return token, nil
	}

	// No token found - return empty string (not an error)
	// This allows unauthenticated GitHub API requests (60 req/hour limit)
	return "", nil
}

// SetGitHubToken stores the GitHub personal access token in the system keyring.
// Returns an error if the keyring is not available or the operation fails.
func SetGitHubToken(token string) error {
	if token == "" {
		return fmt.Errorf("token cannot be empty")
	}

	err := keyring.Set(KeyringService, KeyringUser, token)
	if err != nil {
		return fmt.Errorf("failed to store token in keyring: %w", err)
	}

	return nil
}

// DeleteGitHubToken removes the GitHub personal access token from the system keyring.
// Returns an error if the keyring is not available or the operation fails.
// Returns nil if the token doesn't exist (idempotent).
func DeleteGitHubToken() error {
	err := keyring.Delete(KeyringService, KeyringUser)
	if err != nil && err != keyring.ErrNotFound {
		return fmt.Errorf("failed to delete token from keyring: %w", err)
	}

	return nil
}

// HasGitHubToken checks if a GitHub token is available (either in keyring or environment).
// Returns true if a token is found, false otherwise.
func HasGitHubToken() bool {
	token, err := GetGitHubToken()
	return err == nil && token != ""
}
