package source

import (
	"errors"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport"
)

// AuthenticationError represents an authentication failure
type AuthenticationError struct {
	URL string
	Err error
}

func (e *AuthenticationError) Error() string {
	return "authentication required"
}

func (e *AuthenticationError) Unwrap() error {
	return e.Err
}

// IsAuthenticationError checks if an error is an authentication error
func IsAuthenticationError(err error) bool {
	if err == nil {
		return false
	}

	// Check for our custom error type
	var authErr *AuthenticationError
	if errors.As(err, &authErr) {
		return true
	}

	// Check for go-git transport errors
	if errors.Is(err, transport.ErrAuthenticationRequired) {
		return true
	}
	if errors.Is(err, transport.ErrAuthorizationFailed) {
		return true
	}

	// Check error message for common authentication patterns
	errMsg := strings.ToLower(err.Error())
	authPatterns := []string{
		"authentication required",
		"authentication failed",
		"authorization failed",
		"repository not found", // GitHub returns 404 for private repos without auth
		"could not read username",
		"invalid credentials",
	}

	for _, pattern := range authPatterns {
		if strings.Contains(errMsg, pattern) {
			return true
		}
	}

	return false
}

// WrapAuthenticationError wraps an error as an authentication error if applicable
func WrapAuthenticationError(url string, err error) error {
	if err == nil {
		return nil
	}

	if IsAuthenticationError(err) {
		return &AuthenticationError{
			URL: url,
			Err: err,
		}
	}

	return err
}
