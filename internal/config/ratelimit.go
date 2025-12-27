package config

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/whisller/pkit/pkg/models"
)

// ExtractRateLimitFromResponse extracts rate limit information from GitHub API response headers.
// Returns a RateLimit struct with the extracted information, or nil if headers are missing.
func ExtractRateLimitFromResponse(resp *http.Response, authenticated bool) *models.RateLimit {
	if resp == nil {
		return nil
	}

	// Extract rate limit headers
	limitStr := resp.Header.Get("X-RateLimit-Limit")
	remainingStr := resp.Header.Get("X-RateLimit-Remaining")
	resetStr := resp.Header.Get("X-RateLimit-Reset")
	resource := resp.Header.Get("X-RateLimit-Resource")

	// If any header is missing, return nil
	if limitStr == "" || remainingStr == "" || resetStr == "" {
		return nil
	}

	// Parse values
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return nil
	}

	remaining, err := strconv.Atoi(remainingStr)
	if err != nil {
		return nil
	}

	resetUnix, err := strconv.ParseInt(resetStr, 10, 64)
	if err != nil {
		return nil
	}

	// Default resource to "core" if not specified
	if resource == "" {
		resource = "core"
	}

	return &models.RateLimit{
		Limit:         limit,
		Remaining:     remaining,
		ResetAt:       time.Unix(resetUnix, 0),
		Resource:      resource,
		Authenticated: authenticated,
		UpdatedAt:     time.Now(),
	}
}

// CheckAndWarnRateLimit checks the rate limit and prints a warning if the threshold is exceeded.
// Returns the rate limit information.
func CheckAndWarnRateLimit(rateLimit *models.RateLimit, threshold int) {
	if rateLimit == nil {
		return
	}

	if !rateLimit.ShouldWarn(threshold) {
		return
	}

	// Calculate time until reset
	timeUntil := rateLimit.TimeUntilReset()
	timeUntilStr := formatDuration(timeUntil)

	// Print warning to stderr
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Warning: GitHub API rate limit at %d%% (%d/%d requests remaining)\n",
		rateLimit.PercentageUsed(), rateLimit.Remaining, rateLimit.Limit)
	fmt.Fprintf(os.Stderr, "  Resets at: %s (in %s)\n",
		rateLimit.ResetAt.Format("2006-01-02 15:04:05"), timeUntilStr)

	// If unauthenticated, provide instructions to increase limit
	if !rateLimit.Authenticated {
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "To increase rate limit to 5000 requests/hour:\n")
		fmt.Fprintf(os.Stderr, "  1. Create GitHub personal access token: https://github.com/settings/tokens\n")
		fmt.Fprintf(os.Stderr, "  2. Store token securely: pkit config set-token\n")
		fmt.Fprintf(os.Stderr, "  3. Enable authentication: pkit config set github.use_auth true\n")
	}

	fmt.Fprintf(os.Stderr, "\n")
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < 0 {
		return "now"
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%d hours %d minutes", hours, minutes)
	}
	if minutes > 0 {
		return fmt.Sprintf("%d minutes", minutes)
	}
	return fmt.Sprintf("%d seconds", seconds)
}

// SaveRateLimitToConfig updates the rate limit in the config and saves it.
// This is optional - rate limit can be ephemeral, but persisting helps with status display.
func SaveRateLimitToConfig(cfg *models.Config, rateLimit *models.RateLimit) error {
	if cfg == nil || rateLimit == nil {
		return nil
	}

	// Update rate limit in config
	cfg.GitHub.LastRateLimit = rateLimit

	// Save config (non-fatal if it fails - rate limit is ephemeral)
	if err := Save(cfg); err != nil {
		// Log warning but don't fail the operation
		fmt.Fprintf(os.Stderr, "Warning: failed to save rate limit to config: %v\n", err)
	}

	return nil
}
